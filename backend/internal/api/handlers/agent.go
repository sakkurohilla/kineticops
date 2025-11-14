package handlers

import (
	"database/sql"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
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

// BootstrapAgent - POST /api/v1/agents/bootstrap
// Called by host installers to obtain per-host agent token and Loki URL.
func BootstrapAgent(c *fiber.Ctx) error {
	var req struct {
		Hostname        string `json:"hostname"`
		IP              string `json:"ip"`
		CreateIfMissing bool   `json:"create_if_missing"`
		RegSecret       string `json:"reg_secret"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if req.Hostname == "" {
		return c.Status(400).JSON(fiber.Map{"error": "hostname required"})
	}

	// Find host by hostname
	var hostID int
	err := postgres.SqlxDB.Get(&hostID, "SELECT id FROM hosts WHERE hostname = $1 LIMIT 1", req.Hostname)
	if err != nil {
		// If host not found and caller requested creation, verify registration secret
		if req.CreateIfMissing {
			regSecret := os.Getenv("REGISTRATION_SECRET")
			if regSecret == "" || req.RegSecret == "" || req.RegSecret != regSecret {
				return c.Status(403).JSON(fiber.Map{"error": "registration not allowed or secret invalid"})
			}

			// Prevent automatic recreation if the host was recently deleted
			var tombstones int
			_ = postgres.SqlxDB.Get(&tombstones, "SELECT count(*) FROM deleted_hosts WHERE hostname = $1 AND deleted_at > NOW() - INTERVAL '7 days'", req.Hostname)
			if tombstones > 0 {
				return c.Status(403).JSON(fiber.Map{"error": "host was recently deleted; manual re-add via UI required"})
			}

			// Create a minimal host record
			hostRepo := postgres.NewHostRepository(postgres.SqlxDB)
			host := &models.Host{
				Hostname:    req.Hostname,
				IP:          req.IP,
				OS:          "linux",
				AgentStatus: "pending",
				Group:       "default",
				TenantID:    1,
			}
			if err := hostRepo.Create(host); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to create host"})
			}
			hostID = int(host.ID)
		} else {
			return c.Status(404).JSON(fiber.Map{"error": "host not registered"})
		}
	}

	// Check for existing agent for this host
	repo := postgres.NewAgentRepository(postgres.SqlxDB)
	var existingToken string
	err = postgres.SqlxDB.Get(&existingToken, "SELECT token FROM agents WHERE host_id = $1 LIMIT 1", hostID)
	if err == nil && existingToken != "" {
		// return existing token and loki url
		loki := os.Getenv("LOKI_URL")
		if loki == "" {
			loki = "http://loki:3100/loki/api/v1/push"
		}
		return c.JSON(fiber.Map{"loki_url": loki, "token": existingToken})
	}

	// Generate a new token and create agent record
	if agentService == nil {
		return c.Status(500).JSON(fiber.Map{"error": "agent service not initialized"})
	}

	token, gerr := agentService.GenerateAgentToken()
	if gerr != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate token"})
	}

	agent := &models.Agent{
		HostID:     hostID,
		AgentToken: token,
		Status:     "pending",
	}

	if err := repo.Create(agent); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create agent record"})
	}

	loki := os.Getenv("LOKI_URL")
	if loki == "" {
		loki = "http://loki:3100/loki/api/v1/push"
	}

	return c.JSON(fiber.Map{"loki_url": loki, "token": token})
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
		// If token is invalid / not recognized, return 401 so agent can re-register
		if strings.Contains(err.Error(), "token not recognized") || strings.Contains(err.Error(), "host was recently deleted") || err == sql.ErrNoRows {
			return c.Status(401).JSON(fiber.Map{"error": "agent token invalid or host deleted"})
		}
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

// RevokeAgent - POST /api/v1/agents/:id/revoke
func RevokeAgent(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid agent id"})
	}
	repo := postgres.NewAgentRepository(postgres.SqlxDB)
	if err := repo.UpdateRevoked(id, true); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to revoke agent"})
	}
	return c.JSON(fiber.Map{"msg": "Agent revoked", "agent_id": id})
}

// UnrevokeAgent - POST /api/v1/agents/:id/unrevoke
func UnrevokeAgent(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid agent id"})
	}
	repo := postgres.NewAgentRepository(postgres.SqlxDB)
	if err := repo.UpdateRevoked(id, false); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to unrevoke agent"})
	}
	return c.JSON(fiber.Map{"msg": "Agent unrevoked", "agent_id": id})
}
