package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

var workflowService *services.WorkflowService

func InitWorkflowHandlers(ws *services.WorkflowService) {
	workflowService = ws
}

// CreateWorkflowSession - POST /api/v1/workflow/session
func CreateWorkflowSession(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req models.WorkflowSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if workflowService == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Workflow service not initialized"})
	}

	userIDInt := int(userID.(int64))
	response, err := workflowService.CreateWorkflowSession(&req, userIDInt)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(response)
}

// DiscoverServices - POST /api/v1/workflow/{hostId}/discover
func DiscoverServices(c *fiber.Ctx) error {
	hostID, err := strconv.Atoi(c.Params("hostId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid host ID"})
	}

	sessionToken := c.Get("X-Session-Token")
	if sessionToken == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Session token required"})
	}

	if workflowService == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Workflow service not initialized"})
	}

	// Get real-time services from the global hub's last known state
	// This uses the same websocket data that the Services page uses
	services, err := workflowService.DiscoverServicesRealtime(hostID, sessionToken)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"services": services,
	})
}

// ControlService - POST /api/v1/services/{serviceId}/control
func ControlService(c *fiber.Ctx) error {
	serviceID, err := strconv.Atoi(c.Params("serviceId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid service ID"})
	}

	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	sessionToken := c.Get("X-Session-Token")
	if sessionToken == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Session token required"})
	}

	var req models.ServiceControlRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if workflowService == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Workflow service not initialized"})
	}

	userIDInt := int(userID.(int64))
	response, err := workflowService.ControlService(serviceID, req.Action, sessionToken, userIDInt)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(response)
}

// GetServiceStatus - GET /api/v1/services/{serviceId}/status
func GetServiceStatus(c *fiber.Ctx) error {
	_, err := strconv.Atoi(c.Params("serviceId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid service ID"})
	}

	sessionToken := c.Get("X-Session-Token")
	if sessionToken == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Session token required"})
	}

	// Get real-time service status
	return c.Status(501).JSON(fiber.Map{"error": "Service status not implemented"})
}

// GetWorkflowData - GET /api/v1/hosts/{hostId}/workflow
func GetWorkflowData(c *fiber.Ctx) error {
	hostID, err := strconv.Atoi(c.Params("hostId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid host ID"})
	}

	sessionToken := c.Get("X-Session-Token")
	if sessionToken == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Session token required"})
	}

	if workflowService == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Workflow service not initialized"})
	}

	data, err := workflowService.GetHostWorkflowRealtime(hostID, sessionToken)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(data)
}

// CloseWorkflowSession - DELETE /api/v1/workflow/session
func CloseWorkflowSession(c *fiber.Ctx) error {
	sessionToken := c.Get("X-Session-Token")
	if sessionToken == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Session token required"})
	}

	if workflowService == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Workflow service not initialized"})
	}

	err := workflowService.CloseSession(sessionToken)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "session closed"})
}
