package handlers

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
	"github.com/sakkurohilla/kineticops/backend/internal/telemetry"
	ws "github.com/sakkurohilla/kineticops/backend/internal/websocket"
)

type CollectLogRequest struct {
	HostID      int64             `json:"host_id"`
	Timestamp   time.Time         `json:"timestamp"`
	Level       string            `json:"level"`
	Message     string            `json:"message"`
	Meta        map[string]string `json:"meta"`
	Correlation string            `json:"correlation_id"`
}

func CollectLog(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	var req CollectLogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}
	log := &models.Log{
		TenantID:  tid.(int64),
		HostID:    req.HostID,
		Timestamp: req.Timestamp,
		Level:     req.Level,
		Message:   req.Message,
		Meta:      req.Meta,
		CorrelID:  req.Correlation,
	}
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}
	if err := services.CollectLog(context.Background(), log); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot store log"})
	}

	// Broadcast new log to connected websocket clients (realtime tail)
	// Wrap with a small envelope so frontend can easily filter by type
	payload := map[string]interface{}{
		"type": "log",
		"log":  log,
	}
	if b, err := json.Marshal(payload); err == nil {
		// Best-effort: remember message for new clients and broadcast
		if gh := ws.GetGlobalHub(); gh != nil {
			gh.RememberMessage(b)
		}
		ws.BroadcastToClients(b)
		telemetry.IncWSBroadcast(context.Background(), 1)
	}
	return c.Status(201).JSON(fiber.Map{"msg": "Log stored"})
}

func SearchLogs(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	// Build filters
	filters := make(map[string]interface{})
	if host := c.Query("host_id"); host != "" {
		id, _ := strconv.ParseInt(host, 10, 64)
		filters["host_id"] = id
	}
	if lvl := c.Query("level"); lvl != "" {
		filters["level"] = lvl
	}
	start, _ := time.Parse(time.RFC3339, c.Query("start"))
	end, _ := time.Parse(time.RFC3339, c.Query("end"))
	if !start.IsZero() && !end.IsZero() {
		filters["timestamp"] = map[string]interface{}{
			"$gte": start,
			"$lte": end,
		}
	}
	text := c.Query("search")
	limit, _ := strconv.Atoi(c.Query("limit"))
	skip, _ := strconv.Atoi(c.Query("skip"))
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	logs, err := services.SearchLogs(context.Background(), tid.(int64), filters, text, limit, skip)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot search logs"})
	}
	// fetch total count for pagination
	total, err := services.CountLogs(context.Background(), tid.(int64), filters, text)
	if err != nil {
		// non-fatal: return logs without total
		return c.JSON(fiber.Map{"limit": limit, "skip": skip, "logs": logs})
	}
	return c.JSON(fiber.Map{"total": total, "limit": limit, "skip": skip, "logs": logs})
}

// TriggerLogRetention runs the retention cleanup immediately. Admin-only.
func TriggerLogRetention(c *fiber.Ctx) error {
	// require admin scope in middleware.AuthRequired (already enforced by route)
	days, _ := strconv.Atoi(c.Query("days"))
	if days <= 0 {
		days = 30
	}
	if err := services.EnforceLogRetention(days); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Retention failed"})
	}
	return c.JSON(fiber.Map{"msg": "Retention started", "days": days})
}
