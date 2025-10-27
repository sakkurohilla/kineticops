package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

// CreateHost creates a new host
func CreateHost(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	var req models.Host
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	req.TenantID = tid.(int64)
	req.RegToken = "host-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	req.LastSeen = time.Now()

	if err := postgres.CreateHost(postgres.DB, &req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot create host"})
	}

	return c.JSON(req)
}

// ListHosts returns all hosts for the authenticated user
func ListHosts(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	tenantID := tid.(int64)
	limit := c.Query("limit", "10")
	offset := c.Query("offset", "0")

	hosts, err := services.ListHosts(tenantID, limit, offset)
	if err != nil {
		// Return empty array instead of error when no hosts exist
		return c.JSON([]models.Host{})
	}

	return c.JSON(hosts)
}

// GetHost retrieves a single host by ID
func GetHost(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	tenantID := tid.(int64)
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	host, err := postgres.GetHost(postgres.DB, id)
	if err != nil || host.TenantID != tenantID {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	return c.JSON(host)
}

// UpdateHost updates host fields
func UpdateHost(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	tenantID := tid.(int64)
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	host, err := postgres.GetHost(postgres.DB, id)
	if err != nil || host.TenantID != tenantID {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	var fields map[string]interface{}
	if err := c.BodyParser(&fields); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	err = postgres.UpdateHost(postgres.DB, id, fields)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot update host"})
	}

	return c.JSON(fiber.Map{"msg": "host updated"})
}

// DeleteHost deletes a host
func DeleteHost(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	tenantID := tid.(int64)
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	host, err := postgres.GetHost(postgres.DB, id)
	if err != nil || host.TenantID != tenantID {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	err = postgres.DeleteHost(postgres.DB, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot delete host"})
	}

	return c.JSON(fiber.Map{"msg": "host deleted"})
}

// HostHeartbeat updates host heartbeat/status
func HostHeartbeat(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	tenantID := tid.(int64)
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	host, err := postgres.GetHost(postgres.DB, id)
	if err != nil || host.TenantID != tenantID {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	err = postgres.UpdateHost(postgres.DB, id, map[string]interface{}{
		"agent_status": "online",
		"last_seen":    time.Now(),
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot update status"})
	}

	return c.JSON(fiber.Map{"msg": "host heartbeat updated"})
}
