package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// ProcessMetric represents a process metric record
type ProcessMetric struct {
	ID            int64     `json:"id" gorm:"column:id"`
	HostID        int64     `json:"host_id" gorm:"column:host_id"`
	TenantID      int64     `json:"tenant_id" gorm:"column:tenant_id"`
	PID           int       `json:"pid" gorm:"column:pid"`
	Name          string    `json:"name" gorm:"column:name"`
	Username      string    `json:"username" gorm:"column:username"`
	CPUPercent    float64   `json:"cpu_percent" gorm:"column:cpu_percent"`
	MemoryPercent float64   `json:"memory_percent" gorm:"column:memory_percent"`
	MemoryRSS     int64     `json:"memory_rss" gorm:"column:memory_rss"`
	Status        string    `json:"status" gorm:"column:status"`
	NumThreads    int       `json:"num_threads" gorm:"column:num_threads"`
	CreateTime    int64     `json:"create_time" gorm:"column:create_time"`
	Timestamp     time.Time `json:"timestamp" gorm:"column:timestamp"`
}

// GetHostProcesses returns latest process metrics for a host
func GetHostProcesses(c *fiber.Ctx) error {
	hostID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	sortBy := c.Query("sort", "cpu") // cpu or memory
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	if limit > 100 {
		limit = 100
	}

	// Check if host belongs to authenticated tenant
	var tenantID int64 = 0
	if tid := c.Locals("tenant_id"); tid != nil {
		tenantID = tid.(int64)
	}

	// Verify host access
	host, err := postgres.GetHost(postgres.DB, hostID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Host not found"})
	}
	if tenantID != 0 && host.TenantID != tenantID {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	var processes []ProcessMetric

	// Get latest process metrics (within last 2 minutes)
	err = postgres.DB.Raw(`
		SELECT DISTINCT ON (pid) 
			id, host_id, tenant_id, pid, name, username, cpu_percent, memory_percent, 
			memory_rss, status, num_threads, create_time, timestamp
		FROM process_metrics
		WHERE host_id = ? AND timestamp > NOW() - INTERVAL '2 minutes'
		ORDER BY pid, timestamp DESC
	`, hostID).Scan(&processes).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch process metrics"})
	}

	// Sort in Go after fetching
	if sortBy == "cpu" {
		// Already sorted by timestamp, now need to sort by CPU
		var sortedProcs []ProcessMetric
		err = postgres.DB.Raw(`
			SELECT id, host_id, tenant_id, pid, name, username, cpu_percent, memory_percent, 
				memory_rss, status, num_threads, create_time, timestamp
			FROM (
				SELECT DISTINCT ON (pid) 
					id, host_id, tenant_id, pid, name, username, cpu_percent, memory_percent, 
					memory_rss, status, num_threads, create_time, timestamp
				FROM process_metrics
				WHERE host_id = ? AND timestamp > NOW() - INTERVAL '2 minutes'
				ORDER BY pid, timestamp DESC
			) AS latest
			ORDER BY cpu_percent DESC
			LIMIT ?
		`, hostID, limit).Scan(&sortedProcs).Error

		if err == nil {
			processes = sortedProcs
		}
	} else {
		// Sort by memory
		var sortedProcs []ProcessMetric
		err = postgres.DB.Raw(`
			SELECT id, host_id, tenant_id, pid, name, username, cpu_percent, memory_percent, 
				memory_rss, status, num_threads, create_time, timestamp
			FROM (
				SELECT DISTINCT ON (pid) 
					id, host_id, tenant_id, pid, name, username, cpu_percent, memory_percent, 
					memory_rss, status, num_threads, create_time, timestamp
				FROM process_metrics
				WHERE host_id = ? AND timestamp > NOW() - INTERVAL '2 minutes'
				ORDER BY pid, timestamp DESC
			) AS latest
			ORDER BY memory_percent DESC
			LIMIT ?
		`, hostID, limit).Scan(&sortedProcs).Error

		if err == nil {
			processes = sortedProcs
		}
	}

	return c.JSON(fiber.Map{
		"processes": processes,
		"count":     len(processes),
		"sort_by":   sortBy,
	})
}

// GetAllHostsProcessStats returns process count and top process for all hosts
func GetAllHostsProcessStats(c *fiber.Ctx) error {
	var tenantID int64 = 0
	if tid := c.Locals("tenant_id"); tid != nil {
		tenantID = tid.(int64)
	}

	type HostProcessStats struct {
		HostID        int64   `json:"host_id"`
		ProcessCount  int     `json:"process_count"`
		TopCPUProcess string  `json:"top_cpu_process"`
		TopCPUPercent float64 `json:"top_cpu_percent"`
	}

	var stats []HostProcessStats

	query := `
		SELECT 
			host_id,
			COUNT(DISTINCT pid) as process_count,
			(SELECT name FROM process_metrics p2 
			 WHERE p2.host_id = p1.host_id 
			 AND p2.timestamp > NOW() - INTERVAL '2 minutes'
			 ORDER BY p2.cpu_percent DESC LIMIT 1) as top_cpu_process,
			(SELECT cpu_percent FROM process_metrics p2 
			 WHERE p2.host_id = p1.host_id 
			 AND p2.timestamp > NOW() - INTERVAL '2 minutes'
			 ORDER BY p2.cpu_percent DESC LIMIT 1) as top_cpu_percent
		FROM process_metrics p1
		WHERE timestamp > NOW() - INTERVAL '2 minutes'
	`

	if tenantID != 0 {
		query += ` AND tenant_id = ` + strconv.FormatInt(tenantID, 10)
	}

	query += ` GROUP BY host_id`

	err := postgres.DB.Raw(query).Scan(&stats).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch process stats"})
	}

	return c.JSON(stats)
}
