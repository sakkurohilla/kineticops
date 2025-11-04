package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

type MetricsAggregationHandler struct {
	aggregationService *services.AggregationService
}

func NewMetricsAggregationHandler() *MetricsAggregationHandler {
	return &MetricsAggregationHandler{
		aggregationService: services.NewAggregationService(),
	}
}

// GetTimeSeriesMetrics handles time series queries with custom aggregation
func (h *MetricsAggregationHandler) GetTimeSeriesMetrics(c *fiber.Ctx) error {
	// Extract parameters
	hostID, _ := strconv.ParseInt(c.Query("host_id"), 10, 64)
	metricName := c.Query("metric")
	startTime := c.Query("start")
	endTime := c.Query("end")
	interval := c.Query("interval", "")
	function := c.Query("function", "avg")

	if hostID == 0 || metricName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "host_id and metric are required"})
	}

	// Parse time parameters
	var start, end time.Time
	var err error

	if startTime != "" {
		start, err = time.Parse(time.RFC3339, startTime)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid start time format"})
		}
	} else {
		start = time.Now().Add(-24 * time.Hour)
	}

	if endTime != "" {
		end, err = time.Parse(time.RFC3339, endTime)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid end time format"})
		}
	} else {
		end = time.Now()
	}

	// Build query
	query := services.AggregationQuery{
		MetricName: metricName,
		HostID:     hostID,
		StartTime:  start,
		EndTime:    end,
		Interval:   interval,
		Function:   function,
	}

	// Execute query
	points, err := h.aggregationService.QueryTimeSeries(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"metric": metricName,
		"host_id": hostID,
		"start_time": start.Format(time.RFC3339),
		"end_time": end.Format(time.RFC3339),
		"interval": interval,
		"function": function,
		"data": points,
	})
}

// GetDashboardMetrics returns optimized metrics for dashboard
func (h *MetricsAggregationHandler) GetDashboardMetrics(c *fiber.Ctx) error {
	hostID, _ := strconv.ParseInt(c.Query("host_id"), 10, 64)
	timeRange := c.Query("range", "24h")

	if hostID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "host_id is required"})
	}

	metrics, err := h.aggregationService.GetMetricsForDashboard(hostID, timeRange)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"host_id": hostID,
		"time_range": timeRange,
		"metrics": metrics,
	})
}

// GetLatestMetrics returns current metric values
func (h *MetricsAggregationHandler) GetLatestMetrics(c *fiber.Ctx) error {
	hostID, _ := strconv.ParseInt(c.Query("host_id"), 10, 64)

	if hostID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "host_id is required"})
	}

	metrics, err := h.aggregationService.GetLatestMetrics(hostID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"host_id": hostID,
		"timestamp": time.Now().Format(time.RFC3339),
		"metrics": metrics,
	})
}