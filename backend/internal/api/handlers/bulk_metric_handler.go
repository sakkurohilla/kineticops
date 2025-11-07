package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// BulkMetric represents a single metric in the bulk payload
type BulkMetric struct {
	HostID    int64             `json:"host_id"`
	TenantID  int64             `json:"tenant_id"`
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels"`
	Timestamp *time.Time        `json:"timestamp,omitempty"`
}

// BulkIngestMetrics accepts a JSON array of metrics and streams them into Postgres via COPY for high throughput.
func BulkIngestMetrics(c *fiber.Ctx) error {
	// Read body stream to avoid large allocations
	body := c.Body()
	if len(body) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "empty payload"})
	}

	var metrics []BulkMetric
	if err := json.Unmarshal(body, &metrics); err != nil {
		// Try streaming decode as fallback
		dec := json.NewDecoder(c.Context().RequestBodyStream())
		if dec == nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid JSON payload"})
		}
		metrics = []BulkMetric{}
		// Attempt to read array
		if err := dec.Decode(&metrics); err != nil && err != io.EOF {
			return c.Status(400).JSON(fiber.Map{"error": "invalid JSON payload"})
		}
	}

	if len(metrics) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "no metrics provided"})
	}

	// Use sqlx DB's underlying *sql.DB for COPY
	sqlDB := postgres.SqlxDB.DB
	var err error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server error"})
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server error"})
	}

	stmt, err := tx.Prepare(pq.CopyIn("metrics", "host_id", "tenant_id", "name", "value", "labels", "timestamp"))
	if err != nil {
		_ = tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "server error"})
	}

	for _, m := range metrics {
		labelsJSON, _ := json.Marshal(m.Labels)
		ts := time.Now().UTC()
		if m.Timestamp != nil {
			ts = m.Timestamp.UTC()
		}
		if m.TenantID == 0 {
			// fallback tenant 1 for ingested metrics when not provided
			m.TenantID = 1
		}
		if _, err := stmt.Exec(m.HostID, m.TenantID, m.Name, m.Value, string(labelsJSON), ts); err != nil {
			_ = stmt.Close()
			_ = tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("ingest failed: %v", err)})
		}
	}

	if _, err := stmt.Exec(); err != nil {
		_ = stmt.Close()
		_ = tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("ingest finalize failed: %v", err)})
	}
	if err := stmt.Close(); err != nil {
		_ = tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "server error"})
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "server error"})
	}

	return c.Status(201).JSON(fiber.Map{"ingested": len(metrics)})
}
