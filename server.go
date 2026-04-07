package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type server struct {
	rooms    map[string]*room
	clients  map[net.Addr]*client
	commands chan commands
	mu       sync.RWMutex
	auth     *AuthManager
}

func newServer() *server {
	return &server{
		rooms:    make(map[string]*room),
		clients:  make(map[net.Addr]*client),
		commands: make(chan commands, 100),
		auth:     NewAuthManager(),
	}
}

func (s *server) run() {
	for cmd := range s.commands {
		switch cmd.id {
		case CMD_REGISTER:
			s.register(cmd.client, cmd.args)
		case CMD_LOGIN:
			s.login(cmd.client, cmd.args)
		case CMD_LOGOUT:
			s.logout(cmd.client)
		case CMD_NICK:
			s.nick(cmd.client, cmd.args)
		case CMD_JOIN:
			s.join(cmd.client, cmd.args)
		case CMD_ROOMS:
			s.listRooms(cmd.client)
		case CMD_MSG:
			s.msg(cmd.client, cmd.args)
		case CMD_QUIT:
			s.quit(cmd.client)
		case CMD_USERS:
			s.listUsers(cmd.client)
		case CMD_DM:
			s.dm(cmd.client, cmd.args)
		case CMD_STATUS:
			s.setStatus(cmd.client, cmd.args)
		case CMD_HELP:
			s.help(cmd.client)
		case CMD_HISTORY:
			s.history(cmd.client, cmd.args)
		}
	}
}

func (s *server) newClient(conn net.Conn) *client {
	log.Printf("new user has joined: %s", conn.RemoteAddr().String())

	c := &client{
		conn:     conn,
		nick:     "anonymous",
		status:   "online",
		commands: s.commands,
	}
	s.mu.Lock()
	s.clients[conn.RemoteAddr()] = c
	s.mu.Unlock()
	return c
}

func (s *server) nick(c *client, args []string) {
	if len(args) < 2 {
		c.msg("Username is required. Usage: /nick NAME")
		return
	}

	newNick := args[1]

	// Validate nick length
	if len(newNick) > 20 {
		c.msg("Username must be 20 characters or less")
		return
	}

	if len(newNick) < 2 {
		c.msg("Username must be at least 2 characters")
		return
	}

	oldNick := c.nick
	c.nick = newNick
	c.msg(fmt.Sprintf("COMPLETED. Username changed from '%s' to '%s'", oldNick, newNick))

	// Notify room if user is in one
	if c.room != nil {
		c.room.broadcast(c, fmt.Sprintf("[SYSTEM] %s is now known as %s", oldNick, newNick))
	}
}

func (s *server) join(c *client, args []string) {
	if !s.isAuthenticated(c) {
		return
	}

	if len(args) < 2 {
		c.msg("Room name is required. usage: /join ROOM_NAME")
		return
	}

	roomName := args[1]

	s.mu.Lock()
	r, ok := s.rooms[roomName]
	if !ok {
		r = &room{
			name:    roomName,
			members: make(map[net.Addr]*client),
			mu:      sync.RWMutex{},
		}
		s.rooms[roomName] = r
	}
	s.mu.Unlock()

	r.mu.Lock()
	r.members[c.conn.RemoteAddr()] = c
	r.mu.Unlock()

	s.quitCurrentRoom(c)
	c.room = r

	r.broadcast(c, fmt.Sprintf("[SYSTEM] %s joined the room", c.nick))

	c.msg(fmt.Sprintf("✓ Welcome to #%s", roomName))
}

func (s *server) listRooms(c *client) {
	s.mu.RLock()
	var rooms []string
	for name := range s.rooms {
		rooms = append(rooms, name)
	}
	s.mu.RUnlock()

	if len(rooms) == 0 {
		c.msg("[INFO] No rooms available. Create one by using /join ROOM_NAME")
		return
	}

	c.msg(fmt.Sprintf("[ROOMS] Available: %s", strings.Join(rooms, ", ")))
}

func (s *server) msg(c *client, args []string) {
	if !s.isAuthenticated(c) {
		return
	}

	if c.room == nil {
		c.msg("[ERROR] You must join a room first to send messages. Use /join ROOM_NAME")
		return
	}

	if len(args) < 2 {
		c.msg("[ERROR] Message is required. Usage: /msg MESSAGE")
		return
	}

	msg := strings.Join(args[1:], " ")
	c.room.addMessage(c, msg)
	c.room.broadcast(c, fmt.Sprintf("[%s] %s: %s", time.Now().Format("15:04:05"), c.nick, msg))
}

func (s *server) quit(c *client) {
	log.Printf("user has left the chat: %s", c.conn.RemoteAddr().String())

	s.quitCurrentRoom(c)
	delete(s.clients, c.conn.RemoteAddr())

	c.msg("We will see you soon =(")
	c.conn.Close()
}

func (s *server) quitCurrentRoom(c *client) {
	if c.room != nil {
		c.room.mu.Lock()
		delete(c.room.members, c.conn.RemoteAddr())
		isEmpty := len(c.room.members) == 0
		c.room.mu.Unlock()

		c.room.broadcast(c, fmt.Sprintf("[SYSTEM] %s has left the room", c.nick))

		// Clean up empty rooms
		if isEmpty {
			s.mu.Lock()
			delete(s.rooms, c.room.name)
			s.mu.Unlock()
			log.Printf("room %s was deleted (empty)", c.room.name)
		}
	}
}

func (s *server) listUsers(c *client) {
	if !s.isAuthenticated(c) {
		return
	}

	if c.room == nil {
		c.msg("[ERROR] You are not in a room. Use /join ROOM_NAME to join a room first.")
		return
	}

	c.room.mu.RLock()
	var users []string
	for _, member := range c.room.members {
		status := ""
		if member.status != "" && member.status != "online" {
			status = fmt.Sprintf(" (%s)", member.status)
		}
		users = append(users, member.nick+status)
	}
	c.room.mu.RUnlock()

	if len(users) == 0 {
		c.msg("[USERS] No other users in this room")
	} else {
		c.msg(fmt.Sprintf("[USERS in #%s] %s", c.room.name, strings.Join(users, ", ")))
	}
}

func (s *server) dm(c *client, args []string) {
	if !s.isAuthenticated(c) {
		return
	}

	if len(args) < 3 {
		c.msg("[ERROR] Usage: /dm USERNAME MESSAGE")
		return
	}

	targetNick := args[1]
	message := strings.Join(args[2:], " ")

	if targetNick == c.nick {
		c.msg("[ERROR] You cannot send a direct message to yourself")
		return
	}

	s.mu.RLock()
	var targetClient *client
	for _, client := range s.clients {
		if client.nick == targetNick {
			targetClient = client
			break
		}
	}
	s.mu.RUnlock()

	if targetClient == nil {
		c.msg(fmt.Sprintf("[ERROR] User '%s' not found", targetNick))
		return
	}

	timestamp := time.Now().Format("15:04:05")
	targetClient.msg(fmt.Sprintf("[%s] [DM from %s]: %s", timestamp, c.nick, message))
	c.msg(fmt.Sprintf("[%s] [DM to %s]: %s", timestamp, targetNick, message))
}

func (s *server) setStatus(c *client, args []string) {
	if !s.isAuthenticated(c) {
		return
	}

	if len(args) < 2 {
		c.msg("[ERROR] Usage: /status STATUS (e.g., away, busy, offline)")
		return
	}

	oldStatus := c.status
	status := strings.Join(args[1:], " ")
	c.status = status

	c.msg(fmt.Sprintf("✓ Your status changed from '%s' to '%s'", oldStatus, status))

	if c.room != nil {
		c.room.broadcast(c, fmt.Sprintf("[SYSTEM] %s is now %s", c.nick, status))
	}
}

func (s *server) help(c *client) {
	if !c.authenticated {
		helpText := `
╔════════════════════════════════════════╗
║    Welcome to TCP-Chat (v2.0 Auth)     ║
╠════════════════════════════════════════╣
║                                        ║
║     AUTHENTICATION REQUIRED             ║
║  /register <user> <pass> - Create acc ║
║  /login <user> <pass>    - Sign in    ║
║                                        ║
║  Example:                              ║
║    /register alice mypassword          ║
║    /login alice mypassword             ║
║                                        ║
║  /help     - Show this message         ║
║  /quit     - Exit the chat             ║
╚════════════════════════════════════════╝
`
		c.msg(helpText)
		return
	}

	// Authenticated user help
	helpText := `
╔════════════════════════════════════════╗
║       TCP-Chat Commands (v2.0)         ║
╠════════════════════════════════════════╣
║                                        ║
║ ACCOUNT:                               ║
║  /logout              - Sign out       ║
║                                        ║
║ ROOMS:                                 ║
║  /join <room>        - Join/create     ║
║  /rooms              - List all rooms  ║
║                                        ║
║ MESSAGING:                             ║
║  /msg <text>         - Send to room    ║
║  /dm <user> <text>   - Private message ║
║  /history [n]        - Show messages   ║
║                                        ║
║ PROFILE:                               ║
║  /nick <name>        - Change username ║
║  /status <status>    - Set status      ║
║  /users              - List room users ║
║                                        ║
║ /quit                - Exit the chat   ║
╚════════════════════════════════════════╝
`
	c.msg(helpText)
}

func (s *server) history(c *client, args []string) {
	if !s.isAuthenticated(c) {
		return
	}

	if c.room == nil {
		c.msg("[ERROR] You must join a room first. Use /join ROOM_NAME")
		return
	}

	count := 10 // Default to last 10 messages
	if len(args) > 1 {
		// Parse the count argument if provided
		if num, err := parseint(args[1]); err == nil && num > 0 {
			count = num
		}
	}

	messages := c.room.getMessages(count)
	if len(messages) == 0 {
		c.msg("[HISTORY] No messages in this room yet")
		return
	}

	c.msg(fmt.Sprintf("[HISTORY] Last %d messages in #%s:", len(messages), c.room.name))
	for _, msg := range messages {
		c.msg(fmt.Sprintf("  [%s] %s: %s", msg.timestamp.Format("15:04:05"), msg.sender, msg.content))
	}
}

func parseint(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// ============ Authentication Commands ============

func (s *server) register(c *client, args []string) {
	if len(args) < 3 {
		c.msg("[ERROR] Usage: /register USERNAME PASSWORD")
		return
	}

	username := args[1]
	password := strings.Join(args[2:], " ")

	err := s.auth.Register(username, password)
	if err != nil {
		c.msg(fmt.Sprintf("[ERROR] %v", err))
		return
	}

	c.msg(fmt.Sprintf("✓ Account created for '%s'. Use /login to sign in.", username))
}

func (s *server) login(c *client, args []string) {
	if c.authenticated {
		c.msg("[ERROR] You are already logged in. Use /logout first.")
		return
	}

	if len(args) < 3 {
		c.msg("[ERROR] Usage: /login USERNAME PASSWORD")
		return
	}

	username := args[1]
	password := strings.Join(args[2:], " ")

	// Attempt login
	token, err := s.auth.Login(username, password)
	if err != nil {
		c.msg(fmt.Sprintf("[ERROR] %v", err))
		return
	}

	// Update client state
	c.authenticated = true
	c.username = strings.ToLower(username)
	c.sessionToken = token
	c.nick = username // Set nick to username by default

	c.msg(fmt.Sprintf("✓ Welcome back, %s! You are now authenticated.", username))
	c.msg(fmt.Sprintf("[INFO] Session token: %s (expires in 24 hours)", token[:16]+"..."))
}

func (s *server) logout(c *client) {
	if !c.authenticated {
		c.msg("[ERROR] You are not logged in.")
		return
	}

	username := c.username
	token := c.sessionToken

	// Logout
	err := s.auth.Logout(token)
	if err != nil {
		c.msg(fmt.Sprintf("[ERROR] Logout failed: %v", err))
		return
	}

	// Quit current room if in one
	s.quitCurrentRoom(c)

	// Reset client state
	c.authenticated = false
	c.username = ""
	c.sessionToken = ""
	c.nick = "anonymous"

	c.msg(fmt.Sprintf("✓ You have been logged out, %s.", username))
}

// isAuthenticated checks if a client is logged in
func (s *server) isAuthenticated(c *client) bool {
	if !c.authenticated {
		c.msg("[ERROR] You must be logged in to use this command. Use /login or /register.")
		return false
	}
	return true
}
