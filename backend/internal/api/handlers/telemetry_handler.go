package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"github.com/sakkurohilla/kineticops/backend/internal/telemetry"
	ws "github.com/sakkurohilla/kineticops/backend/internal/websocket"
)

// Telemetry returns a snapshot of in-memory telemetry counters and
// server analytics (DB-backed counts). This endpoint is protected by
// AuthRequired in the routes registration.
func Telemetry(c *fiber.Ctx) error {
	// Start with in-memory counters
	data := telemetry.GetCounters()

	// Add DB-backed analytics. If the DB is unavailable we'll set counts to -1
	var hostsCount int64 = -1
	var agentsCount int64 = -1
	var metricsLastHour int64 = -1

	if postgres.DB != nil {
		// hosts count
		if err := postgres.DB.Raw("SELECT COUNT(*) FROM hosts").Scan(&hostsCount).Error; err != nil {
			hostsCount = -1
		}
		// agents count (if agents table exists)
		if err := postgres.DB.Raw("SELECT COUNT(*) FROM agents").Scan(&agentsCount).Error; err != nil {
			agentsCount = -1
		}
		// metrics in the last hour
		if err := postgres.DB.Raw("SELECT COUNT(*) FROM metrics WHERE timestamp >= NOW() - interval '1 hour'").Scan(&metricsLastHour).Error; err != nil {
			metricsLastHour = -1
		}
	}

	// Active websocket clients
	activeWS := ws.GetGlobalClientCount()

	// Merge analytics into response
	resp := fiber.Map{
		"telemetry":         data,
		"hosts_count":       hostsCount,
		"agents_count":      agentsCount,
		"metrics_last_hour": metricsLastHour,
		"active_ws_clients": activeWS,
		"timestamp":         time.Now().UTC().Format(time.RFC3339),
	}

	return c.JSON(resp)
}
