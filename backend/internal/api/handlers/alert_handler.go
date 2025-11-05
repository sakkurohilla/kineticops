package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

type AlertRuleRequest struct {
	MetricName          string  `json:"metric_name"`
	Operator            string  `json:"operator"`
	Threshold           float64 `json:"threshold"`
	Window              int     `json:"window"`
	Frequency           int     `json:"frequency"`
	NotificationWebhook string  `json:"notification_webhook"`
	EscalationPolicy    string  `json:"escalation_policy"`
}

func CreateAlertRule(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	var req AlertRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}
	rule := &models.AlertRule{
		TenantID:            tid.(int64),
		MetricName:          req.MetricName,
		Operator:            req.Operator,
		Threshold:           req.Threshold,
		Window:              req.Window,
		Frequency:           req.Frequency,
		NotificationWebhook: req.NotificationWebhook,
		EscalationPolicy:    req.EscalationPolicy,
		CreatedAt:           time.Now(),
	}
	if err := services.CreateAlertRule(rule); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot create alert rule"})
	}
	return c.Status(201).JSON(rule)
}

func ListAlertRules(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	rules, err := services.ListAlertRules(tid.(int64))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot list alert rules"})
	}
	return c.JSON(rules)
}

func ListAlerts(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	var alerts []models.Alert
	err := postgres.DB.Where("tenant_id = ?", tid.(int64)).Order("triggered_at desc").Find(&alerts).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot list alerts"})
	}
	return c.JSON(alerts)
}

// UpdateAlert updates an alert's fields (used for silencing/ack)
func UpdateAlert(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing id"})
	}

	var fields map[string]interface{}
	if err := c.BodyParser(&fields); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	// Ensure tenant+ownership: load alert first
	var alert models.Alert
	if err := postgres.DB.Where("id = ? AND tenant_id = ?", idParam, tid.(int64)).First(&alert).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	if err := postgres.DB.Model(&models.Alert{}).Where("id = ?", alert.ID).Updates(fields).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot update alert"})
	}

	return c.JSON(fiber.Map{"msg": "alert updated"})
}

// You can add other handlers: trigger alert manually, list alerts, ack/close alerts, etc.

// GetAlertStats returns alert statistics for dashboard
func GetAlertStats(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		// For now, return public stats when no auth
		var stats = struct {
			Total        int64 `json:"total"`
			Open         int64 `json:"open"`
			Acknowledged int64 `json:"acknowledged"`
			Silenced     int64 `json:"silenced"`
			Resolved     int64 `json:"resolved"`
			Critical     int64 `json:"critical"`
			High         int64 `json:"high"`
			Medium       int64 `json:"medium"`
			Low          int64 `json:"low"`
		}{
			Total:        0,
			Open:         0,
			Acknowledged: 0,
			Silenced:     0,
			Resolved:     0,
			Critical:     0,
			High:         0,
			Medium:       0,
			Low:          0,
		}
		return c.JSON(stats)
	}

	tenantID := tid.(int64)

	var stats struct {
		Total        int64 `json:"total"`
		Open         int64 `json:"open"`
		Acknowledged int64 `json:"acknowledged"`
		Silenced     int64 `json:"silenced"`
		Resolved     int64 `json:"resolved"`
		Critical     int64 `json:"critical"`
		High         int64 `json:"high"`
		Medium       int64 `json:"medium"`
		Low          int64 `json:"low"`
	}

	// Count from existing alerts table
	postgres.DB.Model(&models.Alert{}).Where("tenant_id = ?", tenantID).Count(&stats.Total)
	postgres.DB.Model(&models.Alert{}).Where("tenant_id = ? AND is_resolved = false", tenantID).Count(&stats.Open)

	// For now, return basic stats - enhance later when you have more alert data
	stats.Resolved = stats.Total - stats.Open
	stats.Acknowledged = 0
	stats.Silenced = 0
	stats.Critical = 0
	stats.High = 0
	stats.Medium = 0
	stats.Low = 0

	return c.JSON(stats)
}
