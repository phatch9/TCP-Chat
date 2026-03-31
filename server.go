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
}

func newServer() *server {
	return &server{
		rooms:    make(map[string]*room),
		clients:  make(map[net.Addr]*client),
		commands: make(chan commands, 100),
	}
}

func (s *server) run() {
	for cmd := range s.commands {
		switch cmd.id {
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
		c.msg("nick is required. usage: /nick NAME")
		return
	}

	c.nick = args[1]
	c.msg(fmt.Sprintf("all right, I will call you %s", c.nick))
}

func (s *server) join(c *client, args []string) {
	if len(args) < 2 {
		c.msg("room name is required. usage: /join ROOM_NAME")
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
	helpText := `
╔════════════════════════════════════════╗
║       Available Commands               ║
╠════════════════════════════════════════╣
║ /nick <name>       - Set your username ║
║ /join <room>       - Join a chat room  ║
║ /rooms             - List all rooms    ║
║ /users             - List room members ║
║ /msg <message>     - Send to room      ║
║ /dm <user> <msg>   - Private message   ║
║ /status <status>   - Set your status   ║
║ /history [n]       - Show last n msgs  ║
║ /quit              - Exit the chat     ║
║ /help              - Show this message ║
╚════════════════════════════════════════╝
`
	c.msg(helpText)
}

func (s *server) history(c *client, args []string) {
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
