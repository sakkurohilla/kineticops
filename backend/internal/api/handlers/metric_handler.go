package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

type CollectMetricRequest struct {
	HostID int64             `json:"host_id"`
	Name   string            `json:"name"`
	Value  float64           `json:"value"`
	Labels map[string]string `json:"labels"`
}

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
	data, err := services.ListMetrics(hostID, name, start, end, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot list metrics"})
	}
	return c.JSON(data)
}

func LatestMetric(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	hostID, _ := strconv.ParseInt(c.Query("host_id"), 10, 64)
	name := c.Query("name")
	data, err := services.LatestMetric(hostID, name)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot get latest metric"})
	}
	return c.JSON(data)
}

func PrometheusExport(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).SendString("Unauthorized")
	}
	// TODO: Replace with dynamic query and formatting as your real use case grows
	return c.SendString(`# HELP host_cpu_usage CPU usage
# TYPE host_cpu_usage gauge
host_cpu_usage{host_id="1"} 42.0
`)
}
