package main

import (
	"net"
	"sync"
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
	mu       sync.RWMutex
}

func (r *room) broadcast(sender *client, msg string) {
	r.mu.RLock()
	members := make(map[net.Addr]*client)
	for addr, m := range r.members {
		members[addr] = m
	}
	r.mu.RUnlock()

	for addr, m := range members {
		if sender.conn.RemoteAddr() != addr {
			m.msg(msg)
		}
	}
}

func (r *room) addMessage(sender *client, content string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = append(r.messages, Message{
		sender:    sender.nick,
		content:   content,
		timestamp: time.Now(),
	})
	// Keep only last 100 messages
	if len(r.messages) > 100 {
		r.messages = r.messages[1:]
	}
}

func (r *room) getMessages(count int) []Message {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if count <= 0 || count > len(r.messages) {
		count = len(r.messages)
	}
	result := make([]Message, count)
	copy(result, r.messages[len(r.messages)-count:])
	return result
}
