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

// You can add other handlers: trigger alert manually, list alerts, ack/close alerts, etc.
