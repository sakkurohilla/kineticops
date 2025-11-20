package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://192.168.2.54:3000,http://127.0.0.1:3000",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-CSRF-Token,X-Requested-With",
		AllowCredentials: true, // Required for cookies (CSRF tokens)
		ExposeHeaders:    "Content-Length,Set-Cookie",
		MaxAge:           86400,
	})
}
