package main

import (
	"github.com/gorilla/websocket"
	"github.com/interactiv/blueprints/trace"
	"github.com/stretchr/objx"
	"log"
	"net/http"
)

type Room struct {
	// messages are forwarded to people in that room
	forward chan *message
	join    chan *Client
	leave   chan *Client
	clients map[*Client]bool
	tracer  trace.Tracer
	avatar  Avatar
}

// infinite loop meant to be run in the background
func (r *Room) run() {
	for {
		select {
		case client := <-r.join:
			//joining
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <-r.leave:
			//leaving
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client leaving")
		case msg := <-r.forward:
			r.tracer.Trace("Message received: ", msg)
			//forward message to all clients
			for client := range r.clients {
				select {
				case client.send <- msg:
					log.Printf("message sent")
					//send messaage
					r.tracer.Trace(" -- Message sent to client")
				default:
					//failed to send
					log.Printf("failed to send message")
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace(" -- Failed to send message,closing client")
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
	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("Failed to get auth cookie: ", err)
		return
	}
	client := &Client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize),
		room:     r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}

// newRoom makes a new room thati s ready to go.
func newRoom(avatar Avatar) *Room {
	return &Room{
		forward: make(chan *message),
		join:    make(chan *Client),
		leave:   make(chan *Client),
		clients: make(map[*Client]bool),
		tracer:  trace.NewNull(),
		avatar:  avatar,
	}
}
