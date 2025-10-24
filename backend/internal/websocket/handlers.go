package websocket

import (
	"fmt"

	"github.com/gofiber/contrib/websocket"
	"github.com/golang-jwt/jwt/v4"
)

func WsHandler(hub *Hub, jwtSecret string) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		tokenStr := c.Query("token")
		if tokenStr == "" {
			c.WriteMessage(websocket.TextMessage, []byte("Missing JWT"))
			c.Close()
			return
		}

		fmt.Println("[DEBUG] WebSocket received JWTSecret:", jwtSecret)

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err != nil {
			fmt.Println("[DEBUG] JWT parse error:", err)
		}
		if token == nil || !token.Valid {
			c.WriteMessage(websocket.TextMessage, []byte("Invalid JWT"))
			c.Close()
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.WriteMessage(websocket.TextMessage, []byte("Invalid token claims"))
			c.Close()
			return
		}
		userIDfloat, ok := claims["user_id"].(float64)
		if !ok {
			c.WriteMessage(websocket.TextMessage, []byte("Missing user_id in token"))
			c.Close()
			return
		}
		userID := int64(userIDfloat)

		client := &Client{hub: hub, conn: c, send: make(chan []byte, 256), userID: userID}
		hub.register <- client
		go client.WritePump()
		client.ReadPump()
	}
}
