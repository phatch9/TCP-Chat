package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type server struct {
	rooms    map[string]*room
	clients  map[net.Addr]*client
	commands chan commands
}

func newServer() *server {
	return &server{
		rooms:    make(map[string]*room),
		clients:  make(map[net.Addr]*client),
		commands: make(chan commands),
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
	s.clients[conn.RemoteAddr()] = c
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

	r, ok := s.rooms[roomName]
	if !ok {
		r = &room{
			name:    roomName,
			members: make(map[net.Addr]*client),
		}
		s.rooms[roomName] = r
	}
	r.members[c.conn.RemoteAddr()] = c

	s.quitCurrentRoom(c)
	c.room = r

	r.broadcast(c, fmt.Sprintf("%s joined the room", c.nick))

	c.msg(fmt.Sprintf("welcome to %s", roomName))
}

func (s *server) listRooms(c *client) {
	var rooms []string
	for name := range s.rooms {
		rooms = append(rooms, name)
	}

	c.msg(fmt.Sprintf("available rooms: %s", strings.Join(rooms, ", ")))
}

func (s *server) msg(c *client, args []string) {
	if len(args) < 2 {
		c.msg("Message is required, usage: /msg MSG")
		return
	}

	msg := strings.Join(args[1:], " ")
	c.room.broadcast(c, c.nick+": "+msg)
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
		oldRoom := s.rooms[c.room.name]
		delete(s.rooms[c.room.name].members, c.conn.RemoteAddr())
		oldRoom.broadcast(c, fmt.Sprintf("%s has left the room", c.nick))
	}
}

func (s *server) listUsers(c *client) {
	if c.room == nil {
		c.msg("You are not in a room. Use /join ROOM_NAME to join a room first.")
		return
	}

	var users []string
	for _, member := range c.room.members {
		status := ""
		if member.status != "" && member.status != "online" {
			status = " (" + member.status + ")"
		}
		users = append(users, member.nick+status)
	}

	if len(users) == 0 {
		c.msg("No other users in this room")
	} else {
		c.msg(fmt.Sprintf("Users in %s: %s", c.room.name, strings.Join(users, ", ")))
	}
}

func (s *server) dm(c *client, args []string) {
	if len(args) < 3 {
		c.msg("Usage: /dm USERNAME MESSAGE")
		return
	}

	targetNick := args[1]
	message := strings.Join(args[2:], " ")

	var targetClient *client
	for _, client := range s.clients {
		if client.nick == targetNick {
			targetClient = client
			break
		}
	}

	if targetClient == nil {
		c.msg(fmt.Sprintf("User '%s' not found", targetNick))
		return
	}

	if targetClient == c {
		c.msg("You cannot send a direct message to yourself")
		return
	}

	targetClient.msg(fmt.Sprintf("[DM from %s]: %s", c.nick, message))
	c.msg(fmt.Sprintf("[DM to %s]: %s", targetNick, message))
}

func (s *server) setStatus(c *client, args []string) {
	if len(args) < 2 {
		c.msg("Usage: /status STATUS (e.g., away, busy, offline)")
		return
	}

	status := strings.Join(args[1:], " ")
	c.status = status
	c.msg(fmt.Sprintf("Your status is now: %s", status))

	if c.room != nil {
		c.room.broadcast(c, fmt.Sprintf("%s is now %s", c.nick, status))
	}
}

func (s *server) help(c *client) {
	helpText := `
Available Commands:
  /nick <name>          - Set your username
  /join <room>          - Join a chat room
  /rooms                - List all available rooms
  /users                - List users in current room
  /msg <message>        - Send message to room
  /dm <user> <msg>      - Send private message to user
  /status <status>      - Set your status (e.g., away, busy)
  /quit                 - Exit the chat
  /help                 - Display this help message
`
	c.msg(helpText)
}
