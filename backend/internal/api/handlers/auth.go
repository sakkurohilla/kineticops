package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sakkurohilla/kineticops/backend/config"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

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

// LOGIN user
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
	cfg := config.Load()
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID,
		"username": req.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	})
	refreshTokenStr, _ := refreshToken.SignedString([]byte(cfg.JWTSecret))
	services.LogEvent(userID, "login", req.Username)
	return c.JSON(fiber.Map{"token": token, "refresh_token": refreshTokenStr})
}

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

// REFRESH TOKEN endpoint
func RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}
	cfg := config.Load()
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid refresh token"})
	}
	claims := token.Claims.(jwt.MapClaims)
	userID, _ := claims["user_id"].(float64)
	username, _ := claims["username"].(string)
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  int64(userID),
		"username": username,
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"iat":      time.Now().Unix(),
	})
	tokenStr, _ := newToken.SignedString([]byte(cfg.JWTSecret))
	services.LogEvent(int64(userID), "token_refresh", username)
	return c.JSON(fiber.Map{"token": tokenStr})
}
