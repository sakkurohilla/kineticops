package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/sakkurohilla/kineticops/backend/internal/auth"
)

// WsHandler upgrades HTTP connections to websockets and validates JWT tokens.
// It uses the central auth package to validate tokens so we avoid version
// mismatches between jwt libraries.
func WsHandler(hub *Hub, jwtSecret string) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		// recover from any panic during websocket handling to avoid crashing the whole server
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[ERROR] panic in websocket handler: %v\n", r)
				// best-effort notify client and close
				_ = c.WriteMessage(websocket.TextMessage, []byte("internal server error"))
				c.Close()
			}
		}()
		// We expect the client to send an initial JSON auth message after
		// the websocket handshake. This avoids placing tokens in the URL
		// which can be logged by intermediaries and servers.
		// Example initial message: { "type": "auth", "token": "<jwt>" }

		remote := "unknown"
		if c.RemoteAddr() != nil {
			remote = c.RemoteAddr().String()
		}
		fmt.Printf("[DEBUG] WebSocket incoming connection from %s (awaiting auth)\n", remote)

		// Set a short read deadline to avoid clients leaving connections open
		// without authenticating.
		_ = c.SetReadDeadline(time.Now().Add(10 * time.Second))
		mt, msg, err := c.ReadMessage()
		if err != nil {
			fmt.Printf("[DEBUG] WebSocket failed to read auth message from %s: %v\n", remote, err)
			_ = c.WriteMessage(websocket.TextMessage, []byte("failed to read auth message"))
			c.Close()
			return
		}
		// reset deadline after initial read
		_ = c.SetReadDeadline(time.Time{})

		// Expect a JSON object with type=auth and token
		var init struct {
			Type  string `json:"type"`
			Token string `json:"token"`
		}
		if mt != websocket.TextMessage {
			_ = c.WriteMessage(websocket.TextMessage, []byte("expected text auth message"))
			c.Close()
			return
		}
		if err := json.Unmarshal(msg, &init); err != nil {
			_ = c.WriteMessage(websocket.TextMessage, []byte("invalid auth payload"))
			c.Close()
			return
		}
		if init.Type != "auth" || init.Token == "" {
			_ = c.WriteMessage(websocket.TextMessage, []byte("auth required"))
			c.Close()
			return
		}

		// Validate token
		claims, err := auth.ValidateJWT(init.Token)
		if err != nil {
			fmt.Printf("[DEBUG] WebSocket token validation failed from %s: %v\n", remote, err)
			_ = c.WriteMessage(websocket.TextMessage, []byte("Invalid JWT"))
			c.Close()
			return
		}

		userID := claims.UserID
		fmt.Printf("[DEBUG] WebSocket authenticated user=%d from %s\n", userID, remote)

		// send explicit auth_ok so clients can rely on auth before using socket
		_ = c.WriteMessage(websocket.TextMessage, []byte("{\"type\":\"auth_ok\"}"))

		client := &Client{hub: hub, conn: c, send: make(chan []byte, 256), userID: userID}
		hub.register <- client
		go client.WritePump()
		client.ReadPump()
	}
}
