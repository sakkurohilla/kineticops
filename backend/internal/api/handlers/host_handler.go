package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

var hostAgentService *services.AgentService

func InitHostAgentService(as *services.AgentService) {
	hostAgentService = as
}



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
		// Handle unique constraint (duplicate hostname for tenant) gracefully
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return c.Status(409).JSON(fiber.Map{"error": "Host with this hostname already exists"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Cannot create host"})
	}

	return c.JSON(req)
}

// ListHosts returns all hosts for the authenticated tenant
func ListHosts(c *fiber.Ctx) error {
	// Allow public listing: if tenant_id is not provided (unauthenticated),
	// pass tenantID=0 to services.ListHosts which will return all hosts.
	var tenantID int64 = 0
	if tid := c.Locals("tenant_id"); tid != nil {
		tenantID = tid.(int64)
	}

	limitStr := c.Query("limit", "10")
	offsetStr := c.Query("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	hosts, err := services.ListHosts(tenantID, limit, offset)
	if err != nil {
		return c.JSON([]models.Host{})
	}

	return c.JSON(hosts)
}

// GetHost retrieves a single host by ID
func GetHost(c *fiber.Ctx) error {
	// If unauthenticated, allow fetching host publicly.
	var tenantID int64 = 0
	if tid := c.Locals("tenant_id"); tid != nil {
		tenantID = tid.(int64)
	}
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	host, err := postgres.GetHost(postgres.DB, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}
	if tenantID != 0 && host.TenantID != tenantID {
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

// TestSSHConnection tests SSH connection before saving
func TestSSHConnection(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	var req struct {
		IP         string `json:"ip"`
		Port       int    `json:"port"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		PrivateKey string `json:"private_key"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if req.Port == 0 {
		req.Port = 22
	}

	err := services.TestSSHConnectionWithKey(req.IP, req.Port, req.Username, req.Password, req.PrivateKey)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "SSH connection successful",
	})
}

// GetHostMetrics returns recent metrics for a host
func GetHostMetrics(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	hostID, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	var metrics []map[string]interface{}
	err := postgres.DB.Raw(`
		SELECT * FROM host_metrics 
		WHERE host_id = ? 
		ORDER BY timestamp DESC 
		LIMIT 100
	`, hostID).Scan(&metrics).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot fetch metrics"})
	}

	return c.JSON(metrics)
}

// GetHostLatestMetrics returns only the most recent metric
func GetHostLatestMetrics(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	hostID, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	var metric map[string]interface{}
	err := postgres.DB.Raw(`
		SELECT * FROM host_metrics 
		WHERE host_id = ? 
		ORDER BY timestamp DESC 
		LIMIT 1
	`, hostID).Scan(&metric).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot fetch metrics"})
	}

	if len(metric) == 0 {
		// No metrics yet for this host - return null (client should handle this as 'no data yet')
		return c.JSON(nil)
	}

	return c.JSON(metric)
}

// GetHostMetricsTimeRange returns metrics for a host between start and end timestamps (RFC3339)
func GetHostMetricsTimeRange(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	hostID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	startStr := c.Query("start")
	endStr := c.Query("end")

	var start, end time.Time
	var err error
	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid start timestamp"})
		}
	} else {
		// default to 24 hours ago
		start = time.Now().Add(-24 * time.Hour)
	}
	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid end timestamp"})
		}
	} else {
		end = time.Now()
	}

	var metrics []map[string]interface{}
	err = postgres.DB.Raw(`
		SELECT * FROM host_metrics WHERE host_id = ? AND timestamp BETWEEN ? AND ? ORDER BY timestamp ASC
	`, hostID, start, end).Scan(&metrics).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot fetch metrics"})
	}

	// Return empty array when no data
	if len(metrics) == 0 {
		return c.JSON([]map[string]interface{}{})
	}

	return c.JSON(metrics)
}

// GetHostDashboardMetrics returns summarized metrics for host dashboard visualization
func GetHostDashboardMetrics(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	hostID, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	var metrics []struct {
		CPUUsage    float64   `json:"cpu_usage"`
		MemoryUsage float64   `json:"memory_usage"`
		DiskUsage   float64   `json:"disk_usage"`
		NetworkIn   float64   `json:"network_in"`
		NetworkOut  float64   `json:"network_out"`
		Uptime      int64     `json:"uptime"`
		LoadAverage string    `json:"load_average"`
		Timestamp   time.Time `json:"timestamp"`
	}

	err := postgres.DB.Table("host_metrics").
		Where("host_id = ?", hostID).
		Order("timestamp DESC").
		Limit(100).
		Find(&metrics).Error

	if err != nil {
		// Return empty array on error - NO MOCK DATA
		return c.JSON([]map[string]interface{}{})
	}

	// Return empty array if no metrics found - NO MOCK DATA
	if len(metrics) == 0 {
		return c.JSON([]map[string]interface{}{})
	}

	return c.JSON(metrics)
}

// CollectHostNow triggers an immediate collection for the given host id (debug/testing).
func CollectHostNow(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	tenantID := tid.(int64)

	hostID, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	host, err := postgres.GetHost(postgres.DB, hostID)
	if err != nil || host.TenantID != tenantID {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	metric, err := services.CollectHostMetrics(host)
	if err != nil {
		// mark host offline and return error
		services.UpdateHostStatus(host.ID, "offline")
		return c.Status(500).JSON(fiber.Map{"error": "collect failed", "detail": err.Error()})
	}

	if err := services.SaveHostMetrics(metric); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "save failed", "detail": err.Error()})
	}

	services.UpdateHostStatus(host.ID, "online")

	return c.JSON(fiber.Map{"success": true, "metric": metric})
}

// CreateHostWithAgent creates a host with agent setup
func CreateHostWithAgent(c *fiber.Ctx) error {
	var req models.AgentSetupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if hostAgentService == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Agent service not initialized"})
	}

	response, err := hostAgentService.SetupAgent(&req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(response)
}

// GetHostServices returns discovered services for a host
func GetHostServices(c *fiber.Ctx) error {
	hostID, _ := strconv.Atoi(c.Params("id"))
	
	if hostAgentService == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Agent service not initialized"})
	}

	services, err := hostAgentService.GetHostServices(hostID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(services)
}
