package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sakkurohilla/kineticops/backend/config"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

// Register handles user registration
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
	return c.JSON(fiber.Map{"msg": "User registered successfully. Please login."})
}

// Login handles user authentication
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

// ForgotPassword initiates password reset process
func ForgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if req.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email is required"})
	}

	token, err := services.GeneratePasswordResetToken(req.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not process password reset"})
	}

	// In development: return token directly
	// In production: send email with reset link
	// For now, we return the token for testing
	if token != "" {
		return c.JSON(fiber.Map{
			"msg":   "Password reset token generated. Check your email.",
			"token": token, // REMOVE THIS IN PRODUCTION - only for development
		})
	}

	return c.JSON(fiber.Map{
		"msg": "If your email exists in our system, a reset link was sent.",
	})
}

// VerifyResetToken validates a password reset token
func VerifyResetToken(c *fiber.Ctx) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	resetToken, err := services.VerifyPasswordResetToken(req.Token)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"valid":      true,
		"email":      resetToken.Email,
		"expires_at": resetToken.ExpiresAt,
	})
}

// ResetPassword handles password reset with token
func ResetPassword(c *fiber.Ctx) error {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if req.Token == "" || req.NewPassword == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Token and new password are required"})
	}

	// Validate password strength (minimum 6 characters for now)
	if len(req.NewPassword) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 6 characters"})
	}

	err := services.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "Password reset successful. You can now login with your new password."})
}

// GetCurrentUser returns the current authenticated user
func GetCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	user, err := services.GetUserByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}

// RefreshToken handles token refresh
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
