package main

import (
	"github.com/gorilla/websocket"
	"time"
)

type Client struct {
	//room is the room the client is chatting in
	room *Room
	//channel in which messages are sent
	send chan *message
	//websocket connection
	socket *websocket.Conn
	// user data holds infos about user
	userData map[string]interface{}
}

func (c *Client) read() {
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil {
			msg.When = time.Now()
			msg.Name = c.userData["name"].(string)
			msg.Avatar = c.userData["avatar_url"].(string)
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}
func (c *Client) write() {
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
