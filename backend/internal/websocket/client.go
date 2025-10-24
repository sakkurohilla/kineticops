package websocket

import (
	"github.com/gofiber/contrib/websocket"
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int64
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// No need to process incoming for now.
	}
}

func (c *Client) WritePump() {
	for msg := range c.send {
		c.conn.WriteMessage(websocket.TextMessage, msg)
	}
}

//func (c *Client) WritePump() {
//	defer c.conn.Close()
//	for {
//		select {
//		case msg, ok := <-c.send:
//			if !ok {
//				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
//				return
//			}
//			_ = c.conn.WriteMessage(websocket.TextMessage, msg)
//		}
//	}
//}
