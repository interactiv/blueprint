package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Room struct {
	// messages are forwarded to people in that room
	forward chan []byte
	join    chan *Client
	leave   chan *Client
	clients map[*Client]bool
}

// infinite loop meant to be run in the background
func (r *Room) run() {
	for {
		select {
		case client := <-r.join:
			//joining
			log.Printf("client joining %+v",client)
			r.clients[client] = true
		case client := <-r.leave:
			//leaving
			log.Printf("client leaving %+v",client)
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			//forward message to all clients
			log.Printf("forward message to clients %+v",msg)
			for client := range r.clients {
				select {
				case client.send <- msg:
					log.Printf("message sent")
					//send messaage
				default:
					//failed to send
					log.Printf("failed to send message")
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

// ServeHTTP is the request handler
func (r *Room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &Client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}

// newRoom makes a new room thati s ready to go.
func newRoom() *Room {
	return &Room{
		forward: make(chan []byte),
		join:    make(chan *Client),
		leave:   make(chan *Client),
		clients: make(map[*Client]bool),
	}
}
