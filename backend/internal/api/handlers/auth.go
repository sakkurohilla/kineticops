package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/auth"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

// REGISTER user
func Register(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "All fields required"})
	}
	if err := services.RegisterUser(req.Username, req.Email, req.Password); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	services.LogEvent(0, "register", req.Username)
	return c.JSON(fiber.Map{"msg": "User registered. Please verify your email (mock)."})
}

// LOGIN user and get tokens
func Login(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}
	token, userID, err := services.LoginUser(req.Username, req.Password)
	if err != nil {
		services.LogEvent(0, "failed_login", req.Username)
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}
	// Generate refresh token
	refreshToken, _ := auth.GenerateJWT(userID, req.Username, 24*60*60) // 24 hr expiry
	services.LogEvent(userID, "login", req.Username)
	return c.JSON(fiber.Map{"token": token, "refresh_token": refreshToken})
}

// PASSWORD RESET (mock)
func ForgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}
	if err := services.ForgotPassword(req.Email); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not process password reset"})
	}
	services.LogEvent(0, "password_reset", req.Email)
	return c.JSON(fiber.Map{"msg": "If your email exists in our system, a reset link was sent. (mock)"})
}

// REFRESH TOKEN endpoint (real implementation)
func RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}
	claims, err := auth.ValidateJWT(req.RefreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid refresh token"})
	}
	token, _ := auth.GenerateJWT(claims.UserID, claims.Username, 15*60) // 15 min expiry
	services.LogEvent(claims.UserID, "token_refresh", claims.Username)
	return c.JSON(fiber.Map{"token": token})
}
