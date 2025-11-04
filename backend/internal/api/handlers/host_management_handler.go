package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

type UpdateHostRequest struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Group       string `json:"group"`
}

// UpdateHostDetails allows updating host display name and other details
func UpdateHostDetails(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id")
	if tenantID == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	hostID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid host ID"})
	}

	var req UpdateHostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate input
	if req.DisplayName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Display name is required"})
	}

	// Update host
	updates := map[string]interface{}{
		"description": req.DisplayName, // Use description field for custom name
	}
	
	if req.Group != "" {
		updates["group"] = req.Group
	}

	result := postgres.DB.Model(&models.Host{}).
		Where("id = ? AND tenant_id = ?", hostID, tenantID).
		Updates(updates)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update host"})
	}

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Host not found"})
	}

	return c.JSON(fiber.Map{"message": "Host updated successfully"})
}

// GetHostDetails returns host information
func GetHostDetails(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id")
	if tenantID == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	hostID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid host ID"})
	}

	var host models.Host
	err = postgres.DB.Where("id = ? AND tenant_id = ?", hostID, tenantID).First(&host).Error
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Host not found"})
	}

	return c.JSON(fiber.Map{
		"id": host.ID,
		"hostname": host.Hostname,
		"display_name": host.Description,
		"ip": host.IP,
		"os": host.OS,
		"status": host.Status,
		"agent_status": host.AgentStatus,
		"group": host.Group,
		"last_seen": host.LastSeen,
	})
}