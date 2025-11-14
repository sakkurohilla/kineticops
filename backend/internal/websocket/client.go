package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/telemetry"
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int64
	// simple token-bucket limiter fields (tokens refill over time)
	limiterMu     sync.Mutex
	limiterTokens float64
	limiterLast   time.Time
	limiterRate   float64 // tokens per second
	limiterBurst  float64 // max tokens
}

// AllowSend attempts to consume `n` tokens from the client's bucket and
// returns true when there are enough tokens. It is safe for concurrent use.
func (c *Client) AllowSend(n int) bool {
	c.limiterMu.Lock()
	defer c.limiterMu.Unlock()
	now := time.Now()
	if c.limiterLast.IsZero() {
		c.limiterLast = now
		// initialize tokens to full burst on first use
		c.limiterTokens = c.limiterBurst
	}
	elapsed := now.Sub(c.limiterLast).Seconds()
	if elapsed > 0 {
		c.limiterTokens += elapsed * c.limiterRate
		if c.limiterTokens > c.limiterBurst {
			c.limiterTokens = c.limiterBurst
		}
		c.limiterLast = now
	}
	if float64(n) <= c.limiterTokens {
		c.limiterTokens -= float64(n)
		return true
	}
	return false
}

func (c *Client) ReadPump() {
	// Configure read deadline and pong handler for heartbeat
	defer func() {
		// recover to avoid goroutine panics bringing down process
		if r := recover(); r != nil {
			logging.Errorf("[WS CLIENT] panic in ReadPump user=%d: %v", c.userID, r)
		}
		// ensure unregister and close
		select {
		case c.hub.unregister <- c:
		default:
		}
		_ = c.conn.Close()
	}()

	// read timeout settings
	const pongWait = 60 * time.Second
	// limit maximum incoming message size to protect from large payloads
	const maxMessageSize = 64 * 1024 // 64KB
	// try to set read limit if supported by underlying conn
	// (fasthttp/websocket supports SetReadLimit)
	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		logging.Warnf("[WS CLIENT] SetReadDeadline error user=%d: %v", c.userID, err)
	}
	c.conn.SetPongHandler(func(appData string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			// log read error for diagnostics
			logging.Warnf("[WS CLIENT] read error user=%d remote=%v: %v", c.userID, c.conn.RemoteAddr(), err)
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
		if r := recover(); r != nil {
			logging.Errorf("[WS CLIENT] panic in WritePump user=%d: %v", c.userID, r)
		}
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// channel closed
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				// log write error
				logging.Errorf("[WS CLIENT] write error user=%d remote=%v: %v", c.userID, c.conn.RemoteAddr(), err)
				telemetry.IncWSSendErrors(context.TODO(), 1)
				telemetry.IncClientDisconnect()
				// If write fails, unregister and close connection
				select {
				case c.hub.unregister <- c:
				default:
				}
				_ = c.conn.Close()
				return
			}
		case <-ticker.C:
			// send ping
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				logging.Errorf("[WS CLIENT] ping write error user=%d remote=%v: %v", c.userID, c.conn.RemoteAddr(), err)
				select {
				case c.hub.unregister <- c:
				default:
				}
				_ = c.conn.Close()
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
