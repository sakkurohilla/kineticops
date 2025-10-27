package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

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
	if token != "" {
		return c.JSON(fiber.Map{
			"msg":   "Password reset token generated. Check your email.",
			"token": token, // REMOVE THIS IN PRODUCTION
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

	// Validate password strength
	if len(req.NewPassword) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 6 characters"})
	}

	err := services.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "Password reset successful. You can now login with your new password."})
}

// ChangePassword allows authenticated users to change their password
func ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Both current and new password are required"})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "New password must be at least 6 characters"})
	}

	err := services.ChangeUserPassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	services.LogEvent(userID, "password_changed", "")
	return c.JSON(fiber.Map{"msg": "Password changed successfully"})
}
