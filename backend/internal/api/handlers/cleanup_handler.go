package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// CleanupAllData removes all data from system - for testing/reset
func CleanupAllData(c *fiber.Ctx) error {
	// Truncate all tables
	tables := []string{
		"host_metrics",
		"hosts",
		"installation_tokens",
		"alerts",
		"agents",
		"workflow_sessions",
		"service_control_logs",
		"alert_history",
		"alert_comments",
		"alert_assignments",
		"agent_services",
		"agent_installation_logs",
	}

	for _, table := range tables {
		postgres.DB.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE")
	}

	return c.JSON(fiber.Map{
		"message":        "All system data cleared successfully",
		"tables_cleared": len(tables),
	})
}
