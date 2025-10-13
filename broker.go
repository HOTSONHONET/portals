package main

import (
	"fmt"
	"strings"
	"sync"
)

type Broker struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

func NewBroker() *Broker {
	return &Broker{
		clients: map[chan string]struct{}{},
	}
}

func (b *Broker) Add(c chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[c] = struct{}{}
}

func (b *Broker) Remove(c chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.clients, c)
	close(c)
}

func convert2sseEvent(event, html string) string {
	html = strings.ReplaceAll(html, "\r\n", "\n")
	html = strings.ReplaceAll(html, "\n", "\ndata: ")
	return fmt.Sprintf("event: %v\ndata: %v\n\n", event, html)
}

func (b *Broker) Broadcast(event, html string) {
	msg := convert2sseEvent(event, html)
	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.clients {
		// Sending data to active players (channels) which can take data
		select {
		case ch <- msg:
		default:
		}
	}
}
