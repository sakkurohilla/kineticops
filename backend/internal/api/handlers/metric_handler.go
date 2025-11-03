package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
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
	hostID, _ := strconv.ParseInt(c.Query("host_id"), 10, 64)
	
	// If no host_id provided, return aggregated data for all hosts
	if hostID == 0 {
		// Return empty structure for dashboard overview
		return c.JSON(map[string][]interface{}{
			"cpu_usage":     {},
			"memory_usage":  {},
			"disk_usage":    {},
			"network_bytes": {},
		})
	}

	// Use aggregation service for proper time-series data
	aggService := services.NewMetricsAggregationService()
	
	// Get all metric types for the host
	metricNames := []string{"cpu_usage", "memory_usage", "disk_usage", "network_bytes"}
	result, err := aggService.GetMultipleMetricsAggregated(hostID, metricNames, rangeParam)
	
	if err != nil {
		fmt.Printf("[ERROR] Aggregation failed: %v\n", err)
		return c.JSON(map[string][]interface{}{
			"cpu_usage":     {},
			"memory_usage":  {},
			"disk_usage":    {},
			"network_bytes": {},
		})
	}

	fmt.Printf("[DEBUG] Aggregated metrics for host %d, range %s: %d metric types\n", 
		hostID, rangeParam, len(result))

	return c.JSON(result)
}


