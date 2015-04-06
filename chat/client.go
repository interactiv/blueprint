package main

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	//room is the room the client is chatting in
	room *Room
	//channel in which messages are sent
	send chan []byte
	//websocket connection
	socket *websocket.Conn
}

func (c *Client) read() {
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

func (c *Client) write() {
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
