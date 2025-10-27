package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

// GetCurrentUser returns the authenticated user's profile
func GetCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	user, err := services.GetUserByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}

// UpdateUser updates authenticated user's own profile
func UpdateUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if err := services.UpdateUserProfile(userID, req.Email, req.Username); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "Profile updated successfully"})
}

// DeleteUser soft deletes authenticated user's own account
func DeleteUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	if err := services.DeleteUser(userID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "Account deleted successfully"})
}

// ListUsers returns all users (admin only)
func ListUsers(c *fiber.Ctx) error {
	// TODO: Add admin role check
	users, err := services.GetAllUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch users"})
	}

	return c.JSON(users)
}

// ✅ NEW: GetUserByID - Admin gets specific user
func GetUserByID(c *fiber.Ctx) error {
	// TODO: Add admin role check

	userIDParam := c.Params("id")
	userID, err := strconv.ParseInt(userIDParam, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	user, err := services.GetUserByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}

// ✅ NEW: UpdateUserByID - Admin updates any user
func UpdateUserByID(c *fiber.Ctx) error {
	// TODO: Add admin role check

	userIDParam := c.Params("id")
	userID, err := strconv.ParseInt(userIDParam, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if err := services.UpdateUserProfile(userID, req.Email, req.Username); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "User updated successfully"})
}

// ✅ NEW: DeleteUserByID - Admin deletes any user
func DeleteUserByID(c *fiber.Ctx) error {
	// TODO: Add admin role check

	userIDParam := c.Params("id")
	userID, err := strconv.ParseInt(userIDParam, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	if err := services.DeleteUser(userID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "User deleted successfully"})
}
