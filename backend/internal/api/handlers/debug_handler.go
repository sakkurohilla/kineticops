package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	ws "github.com/sakkurohilla/kineticops/backend/internal/websocket"
)

// DebugWSSend accepts a JSON body to burst-send messages to connected websocket clients.
// Example body: { "count": 1000, "host_id": 123, "interval_ms": 1 }
func DebugWSSend(c *fiber.Ctx) error {
	var req struct {
		Count      int   `json:"count"`
		HostID     int64 `json:"host_id"`
		IntervalMs int   `json:"interval_ms"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad request"})
	}
	if req.Count <= 0 {
		req.Count = 1000
	}
	if req.IntervalMs < 0 {
		req.IntervalMs = 0
	}

	go func(count int, hid int64, interval int) {
		seq := uint64(time.Now().UnixNano())
		for i := 0; i < count; i++ {
			payload := map[string]interface{}{
				"host_id":      hid,
				"seq":          seq + uint64(i),
				"timestamp":    time.Now().UTC().Format(time.RFC3339Nano),
				"cpu_usage":    mathRandFloat(5, 90),
				"memory_usage": mathRandFloat(10, 90),
			}
			b, _ := json.Marshal(payload)
			ws.BroadcastToClients(b)
			// small sleep to avoid totally blocking
			if interval > 0 {
				time.Sleep(time.Duration(interval) * time.Millisecond)
			}
		}
	}(req.Count, req.HostID, req.IntervalMs)

	return c.JSON(fiber.Map{"msg": fmt.Sprintf("dispatched %d websocket messages for host %d", req.Count, req.HostID)})
}

func mathRandFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
