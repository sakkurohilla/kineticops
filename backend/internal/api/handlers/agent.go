package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

var agentService *services.AgentService

func InitAgentHandlers(as *services.AgentService) {
	agentService = as
}

// RegisterAgent - POST /api/v1/agents/register
func RegisterAgent(c *fiber.Ctx) error {
	var req struct {
		Token    string               `json:"token"`
		Metadata models.AgentMetadata `json:"metadata"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	// Validate token and mark agent as registered
	agent, err := agentService.GetAgentStatus(0) // This needs to be updated to use token
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Agent not found"})
	}

	return c.JSON(fiber.Map{
		"status":   "registered",
		"agent_id": agent.ID,
	})
}

// Heartbeat - POST /api/v1/agents/heartbeat
func AgentHeartbeat(c *fiber.Ctx) error {
	var heartbeat models.AgentHeartbeat
	if err := c.BodyParser(&heartbeat); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if agentService == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Agent service not initialized"})
	}

	err := agentService.RegisterAgentHeartbeat(&heartbeat)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// GetAgentStatus - GET /api/v1/agents/{id}/status
func GetAgentStatus(c *fiber.Ctx) error {
	hostID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid host ID"})
	}

	if agentService == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Agent service not initialized"})
	}

	agent, err := agentService.CheckAgentStatus(hostID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Agent not found"})
	}

	return c.JSON(agent)
}

// GetAgentServices - GET /api/v1/agents/{id}/services
func GetAgentServices(c *fiber.Ctx) error {
	agentID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid agent ID"})
	}

	if agentService == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Agent service not initialized"})
	}

	services, err := agentService.FetchServices(agentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(services)
}

// ExecuteCommand - POST /api/v1/agents/{id}/execute
func ExecuteAgentCommand(c *fiber.Ctx) error {
	agentID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid agent ID"})
	}

	var req struct {
		Command string `json:"command"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	// This would execute command via agent
	// For now, return success
	return c.JSON(fiber.Map{
		"success":  true,
		"output":   "Command executed successfully",
		"agent_id": agentID,
	})
}
