package websocket

import (
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int64
}

func (c *Client) ReadPump() {
	// Configure read deadline and pong handler for heartbeat
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// read timeout settings
	const pongWait = 60 * time.Second
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			// log read error for diagnostics
			log.Printf("[WS CLIENT] read error user=%d remote=%v: %v", c.userID, c.conn.RemoteAddr(), err)
			break
		}
		// No need to process incoming for now.
	}
}

func (c *Client) WritePump() {
	// send ping periodically to keep connection alive and detect dead peers
	const (
		writeWait  = 10 * time.Second
		pingPeriod = 30 * time.Second
	)

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// channel closed
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				// log write error
				log.Printf("[WS CLIENT] write error user=%d remote=%v: %v", c.userID, c.conn.RemoteAddr(), err)
				// If write fails, unregister and close connection
				c.hub.unregister <- c
				c.conn.Close()
				return
			}
		case <-ticker.C:
			// send ping
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("[WS CLIENT] ping write error user=%d remote=%v: %v", c.userID, c.conn.RemoteAddr(), err)
				c.hub.unregister <- c
				c.conn.Close()
				return
			}
		}
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
