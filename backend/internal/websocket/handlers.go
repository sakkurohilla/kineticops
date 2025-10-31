package websocket

import (
	"fmt"

	"github.com/gofiber/contrib/websocket"
	"github.com/sakkurohilla/kineticops/backend/internal/auth"
)

// WsHandler upgrades HTTP connections to websockets and validates JWT tokens.
// It uses the central auth package to validate tokens so we avoid version
// mismatches between jwt libraries.
func WsHandler(hub *Hub, jwtSecret string) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		tokenStr := c.Query("token")
		if tokenStr == "" {
			c.WriteMessage(websocket.TextMessage, []byte("Missing JWT"))
			c.Close()
			return
		}

		remote := "unknown"
		if c.RemoteAddr() != nil {
			remote = c.RemoteAddr().String()
		}
		fmt.Printf("[DEBUG] WebSocket incoming connection from %s\n", remote)

		// Use auth.ValidateJWT to parse and validate the token using the
		// same library and claim types used by the rest of the server.
		claims, err := auth.ValidateJWT(tokenStr)
		if err != nil {
			fmt.Printf("[DEBUG] WebSocket token validation failed from %s: %v\n", remote, err)
			c.WriteMessage(websocket.TextMessage, []byte("Invalid JWT"))
			c.Close()
			return
		}

		userID := claims.UserID
		fmt.Printf("[DEBUG] WebSocket authenticated user=%d from %s\n", userID, remote)

		client := &Client{hub: hub, conn: c, send: make(chan []byte, 256), userID: userID}
		hub.register <- client
		go client.WritePump()
		client.ReadPump()
	}
}
