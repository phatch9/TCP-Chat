package main

import (
	"net"
	"time"
)

type Message struct {
	sender    string
	content   string
	timestamp time.Time
}

type room struct {
	name     string
	members  map[net.Addr]*client
	messages []Message
}

func (r *room) broadcast(sender *client, msg string) {
	for addr, m := range r.members {
		if sender.conn.RemoteAddr() != addr {
			m.msg(msg)
		}
	}
}

func (r *room) addMessage(sender *client, content string) {
	r.messages = append(r.messages, Message{
		sender:    sender.nick,
		content:   content,
		timestamp: time.Now(),
	})
	// Keep only last 50 messages
	if len(r.messages) > 50 {
		r.messages = r.messages[1:]
	}
}
