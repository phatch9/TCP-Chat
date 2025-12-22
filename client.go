package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type client struct {
	conn     net.Conn
	nick     string
	room     *room
	status   string
	commands chan<- commands
}

func (c *client) readInput() {
	for {
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			return
		}

		msg = strings.Trim(msg, "\r\n")

		args := strings.Split(msg, " ")
		cmd := strings.TrimSpace(args[0])

		switch cmd {
		case "/nick":
			c.commands <- commands{
				id:     CMD_NICK,
				client: c,
				args:   args,
			}
		case "/join":
			c.commands <- commands{
				id:     CMD_JOIN,
				client: c,
				args:   args,
			}
		case "/rooms":
			c.commands <- commands{
				id:     CMD_ROOMS,
				client: c,
			}
		case "/msg":
			c.commands <- commands{
				id:     CMD_MSG,
				client: c,
				args:   args,
			}
		case "/quit":
			c.commands <- commands{
				id:     CMD_QUIT,
				client: c,
			}
		case "/users":
			c.commands <- commands{
				id:     CMD_USERS,
				client: c,
			}
		case "/dm":
			c.commands <- commands{
				id:     CMD_DM,
				client: c,
				args:   args,
			}
		case "/status":
			c.commands <- commands{
				id:     CMD_STATUS,
				client: c,
				args:   args,
			}
		case "/help":
			c.commands <- commands{
				id:     CMD_HELP,
				client: c,
			}
		default:
			c.err(fmt.Errorf("unknown command: %s", cmd))
		}
	}
}

func (c *client) err(err error) {
	c.conn.Write([]byte("err: " + err.Error() + "\n"))
}

func (c *client) msg(msg string) {
	c.conn.Write([]byte("> " + msg + "\n"))
}
