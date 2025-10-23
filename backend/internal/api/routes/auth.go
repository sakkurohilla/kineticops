package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/auth"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"gorm.io/gorm"
)

func RegisterAuthRoutes(app *fiber.App) {
	// REGISTER user
	app.Post("/auth/register", func(c *fiber.Ctx) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Email    string `json:"email"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
		}
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Hash failed"})
		}

		user := models.User{
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: hash,
		}
		// Store in Postgres (using GORM)
		if err := postgres.DB.Create(&user).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "User creation failed"})
		}
		return c.JSON(fiber.Map{"msg": "User registered"})
	})

	// LOGIN with user from DB
	app.Post("/auth/login", func(c *fiber.Ctx) error {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&creds); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
		}

		var user models.User
		result := postgres.DB.Where("username = ?", creds.Username).First(&user)
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
		} else if result.Error != nil {
			return c.Status(500).JSON(fiber.Map{"error": "DB error"})
		}

		if !auth.CheckPasswordHash(creds.Password, user.PasswordHash) {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
		}
		token, _ := auth.GenerateJWT(user.ID, user.Username, time.Minute*15)
		return c.JSON(fiber.Map{"token": token})
	})
}
