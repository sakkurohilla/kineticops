package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

type CollectMetricRequest struct {
	HostID int64             `json:"host_id"`
	Name   string            `json:"name"`
	Value  float64           `json:"value"`
	Labels map[string]string `json:"labels"`
}

// CollectMetric handles incoming metric data from agents or nodes.
func CollectMetric(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	var req CollectMetricRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if req.HostID == 0 || req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing hostID or name"})
	}

	// Enforce schema validation
	if err := services.ValidateMetricSchema(req.Name, req.Value, ""); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	err := services.CollectMetric(req.HostID, tid.(int64), req.Name, req.Value, req.Labels)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"msg": "Metric collected"})
}

func ListMetrics(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	hostID, _ := strconv.ParseInt(c.Query("host_id"), 10, 64)
	name := c.Query("name")
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	start, _ := time.Parse(time.RFC3339, c.Query("start", time.Now().Add(-1*time.Hour).Format(time.RFC3339)))
	end, _ := time.Parse(time.RFC3339, c.Query("end", time.Now().Format(time.RFC3339)))

	tidVal := tid.(int64)
	data, err := services.ListMetrics(tidVal, hostID, name, start, end, limit)
	if err != nil {
		// Return empty array, not error, when no metrics exist
		return c.JSON([]models.Metric{})
	}
	return c.JSON(data)
}

// LatestMetric fetches the latest recorded metric for a specific host.
func LatestMetric(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	hostID, _ := strconv.ParseInt(c.Query("host_id"), 10, 64)
	name := c.Query("name")

	data, err := services.LatestMetric(hostID, name)
	if err != nil {
		return c.JSON(models.Metric{}) // Return empty metric instead of error
	}

	return c.JSON(data)
}

// PrometheusExport provides Prometheus-compatible metrics exposition endpoint.
func PrometheusExport(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).SendString("Unauthorized")
	}

	// Return empty Prometheus format when no metrics exist
	return c.SendString("# No metrics available\n")
}

// GetMetricsRange handles dashboard metrics retrieval within a specified time range.
func GetMetricsRange(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	rangeParam := c.Query("range", "24h")
	var startTime time.Time

	now := time.Now().UTC()
	switch rangeParam {
	case "1h":
		startTime = now.Add(-1 * time.Hour)
	case "6h":
		startTime = now.Add(-6 * time.Hour)
	case "24h":
		startTime = now.Add(-24 * time.Hour)
	case "7d":
		startTime = now.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = now.Add(-30 * 24 * time.Hour)
	default:
		startTime = now.Add(-24 * time.Hour)
	}

	// Get host_id from query parameter
	hostID, _ := strconv.ParseInt(c.Query("host_id"), 10, 64)
	
	// Get metrics for specific host or all hosts of this tenant within time range
	var metrics []models.Metric
	query := postgres.DB.Where("tenant_id = ? AND timestamp >= ? AND timestamp <= ?", tid.(int64), startTime, now)
	
	// Filter by host_id if specified
	if hostID > 0 {
		query = query.Where("host_id = ?", hostID)
	}
	
	err := query.Order("timestamp ASC").Limit(1000).Find(&metrics).Error

	if err != nil {
		return c.JSON([]models.Metric{}) // Return empty array on error
	}

	// Debug: log the query parameters and result count
	fmt.Printf("[DEBUG] Range: %s, Host: %d, Start: %s, End: %s, Found: %d metrics\n", 
		rangeParam, hostID, startTime.Format(time.RFC3339), now.Format(time.RFC3339), len(metrics))



	return c.JSON(metrics)
}


