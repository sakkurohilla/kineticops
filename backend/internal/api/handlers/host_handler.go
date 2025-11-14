package handlers

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

// CreateHost creates a new host
func CreateHost(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	var req models.Host
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	req.TenantID = tid.(int64)
	req.RegToken = "host-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	req.LastSeen = time.Now()

	if err := postgres.CreateHost(postgres.DB, &req); err != nil {
		// Handle unique constraint (duplicate hostname for tenant) gracefully
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return c.Status(409).JSON(fiber.Map{"error": "Host with this hostname already exists"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Cannot create host"})
	}

	return c.JSON(req)
}

// ListHosts returns all hosts for the authenticated tenant
func ListHosts(c *fiber.Ctx) error {
	// Allow public listing: if tenant_id is not provided (unauthenticated),
	// pass tenantID=0 to services.ListHosts which will return all hosts.
	var tenantID int64 = 0
	if tid := c.Locals("tenant_id"); tid != nil {
		tenantID = tid.(int64)
	}

	limitStr := c.Query("limit", "10")
	offsetStr := c.Query("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	hosts, err := services.ListHosts(tenantID, limit, offset)
	if err != nil {
		return c.JSON([]models.Host{})
	}

	return c.JSON(hosts)
}

// GetHost retrieves a single host by ID
func GetHost(c *fiber.Ctx) error {
	// If unauthenticated, allow fetching host publicly.
	var tenantID int64 = 0
	if tid := c.Locals("tenant_id"); tid != nil {
		tenantID = tid.(int64)
	}
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	host, err := postgres.GetHost(postgres.DB, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}
	if tenantID != 0 && host.TenantID != tenantID {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	return c.JSON(host)
}

// UpdateHost updates host fields
func UpdateHost(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	tenantID := tid.(int64)
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	host, err := postgres.GetHost(postgres.DB, id)
	if err != nil || host.TenantID != tenantID {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	var fields map[string]interface{}
	if err := c.BodyParser(&fields); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	err = postgres.UpdateHost(postgres.DB, id, fields)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot update host"})
	}

	return c.JSON(fiber.Map{"msg": "host updated"})
}

// DeleteHost deletes a host - PUBLIC ACCESS for auto-discovered hosts
func DeleteHost(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	// Get host info before deletion
	host, err := postgres.GetHost(postgres.DB, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Host not found"})
	}

	// EXPIRE installation tokens for this tenant to prevent reuse of install tokens
	if res := postgres.DB.Model(&models.InstallationToken{}).Where("tenant_id = ?", host.TenantID).Updates(map[string]interface{}{
		"used":       true,
		"expires_at": time.Now().Add(-1 * time.Hour), // Expire 1 hour ago
	}); res.Error != nil {
		logging.Warnf("failed to expire installation tokens for tenant=%d: %v", host.TenantID, res.Error)
	}

	// Revoke and clear any agent tokens associated with this host to ensure
	// agents cannot reconnect using previously issued credentials.
	// Set revoked=true, revoked_at and clear the token for auditability.
	if res := postgres.DB.Exec("UPDATE agents SET revoked = TRUE, revoked_at = CURRENT_TIMESTAMP, token = NULL WHERE host_id = ?", id); res.Error != nil {
		logging.Warnf("failed to revoke agents for host_id=%d: %v", id, res.Error)
	}

	// Delete agent-related child records
	if res := postgres.DB.Exec("DELETE FROM agent_services WHERE agent_id IN (SELECT id FROM agents WHERE host_id = ?)", id); res.Error != nil {
		logging.Warnf("failed to delete agent_services for host_id=%d: %v", id, res.Error)
	}
	if res := postgres.DB.Exec("DELETE FROM agent_installation_logs WHERE agent_id IN (SELECT id FROM agents WHERE host_id = ?)", id); res.Error != nil {
		logging.Warnf("failed to delete agent_installation_logs for host_id=%d: %v", id, res.Error)
	}

	// Delete timeseries and snapshot data for the host
	if res := postgres.DB.Exec("DELETE FROM host_metrics WHERE host_id = ?", id); res.Error != nil {
		logging.Warnf("failed to delete host_metrics for host_id=%d: %v", id, res.Error)
	}
	if res := postgres.DB.Exec("DELETE FROM metrics WHERE host_id = ?", id); res.Error != nil {
		logging.Warnf("failed to delete metrics for host_id=%d: %v", id, res.Error)
	}
	if res := postgres.DB.Exec("DELETE FROM timeseries_metrics WHERE host_id = ?", id); res.Error != nil {
		logging.Warnf("failed to delete timeseries_metrics for host_id=%d: %v", id, res.Error)
	}
	if res := postgres.DB.Exec("DELETE FROM logs WHERE host_id = ?", id); res.Error != nil {
		logging.Warnf("failed to delete logs for host_id=%d: %v", id, res.Error)
	}

	// Finally delete any agent rows (optional - tokens already cleared)
	if res := postgres.DB.Exec("DELETE FROM agents WHERE host_id = ?", id); res.Error != nil {
		logging.Warnf("failed to delete agents for host_id=%d: %v", id, res.Error)
	}

	// Delete the host
	err = postgres.DeleteHost(postgres.DB, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot delete host"})
	}

	// Record tombstone to prevent immediate auto-recreation by agents
	if res := postgres.DB.Exec("INSERT INTO deleted_hosts (hostname, tenant_id, deleted_at) VALUES (?, ?, CURRENT_TIMESTAMP)", host.Hostname, host.TenantID); res.Error != nil {
		logging.Warnf("failed to insert deleted_host tombstone for hostname=%s tenant=%d: %v", host.Hostname, host.TenantID, res.Error)
	}

	return c.JSON(fiber.Map{
		"msg":  "Host deleted and all tokens expired",
		"note": "Agent will stop working - generate new token to reconnect",
	})
}

// HostHeartbeat updates host heartbeat/status
func HostHeartbeat(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	tenantID := tid.(int64)
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	host, err := postgres.GetHost(postgres.DB, id)
	if err != nil || host.TenantID != tenantID {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	err = postgres.UpdateHost(postgres.DB, id, map[string]interface{}{
		"agent_status": "online",
		"last_seen":    time.Now(),
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot update status"})
	}

	return c.JSON(fiber.Map{"msg": "host heartbeat updated"})
}

// TestSSHConnection tests SSH connection before saving
func TestSSHConnection(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	var req struct {
		IP         string `json:"ip"`
		Port       int    `json:"port"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		PrivateKey string `json:"private_key"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if req.Port == 0 {
		req.Port = 22
	}

	err := services.TestSSHConnectionWithKey(req.IP, req.Port, req.Username, req.Password, req.PrivateKey)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "SSH connection successful",
	})
}

// GetHostMetrics returns recent metrics for a host
func GetHostMetrics(c *fiber.Ctx) error {
	hostID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	rangeParam := c.Query("range", "24h")

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

	// Use aggregation service for proper time-series data
	aggService := services.NewMetricsAggregationService()
	metricNames := []string{"cpu_usage", "memory_usage", "disk_usage", "network_bytes"}
	result, err := aggService.GetMultipleMetricsAggregated(hostID, metricNames, rangeParam)

	if err != nil {
		return c.JSON(map[string][]interface{}{
			"cpu_usage":     {},
			"memory_usage":  {},
			"disk_usage":    {},
			"network_bytes": {},
		})
	}

	return c.JSON(result)
}

// GetHostLatestMetrics returns only the most recent metric
func GetHostLatestMetrics(c *fiber.Ctx) error {
	hostID, _ := strconv.ParseInt(c.Params("id"), 10, 64)

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

	// Get latest metrics from the metrics table
	var metrics []struct {
		Name      string    `json:"name"`
		Value     float64   `json:"value"`
		Timestamp time.Time `json:"timestamp"`
	}

	// Get latest metrics with special handling for disk_usage
	var allMetrics []struct {
		Name      string    `json:"name"`
		Value     float64   `json:"value"`
		Timestamp time.Time `json:"timestamp"`
	}

	// Get all recent metrics
	err = postgres.DB.Raw(`
		SELECT name, value, timestamp
		FROM metrics 
		WHERE host_id = ? AND tenant_id = ? AND timestamp > NOW() - INTERVAL '5 minutes'
		ORDER BY timestamp DESC
	`, hostID, host.TenantID).Scan(&allMetrics).Error

	// `err` was already checked after the DB scan above; no additional check needed here.

	// Process metrics with special disk_usage handling
	metricMap := make(map[string]struct {
		Value     float64
		Timestamp time.Time
	})

	for _, m := range allMetrics {
		if m.Name == "disk_usage" {
			// For disk_usage, always use the minimum value (root filesystem)
			if existing, exists := metricMap[m.Name]; !exists || m.Value < existing.Value {
				metricMap[m.Name] = struct {
					Value     float64
					Timestamp time.Time
				}{m.Value, m.Timestamp}
			}
		} else {
			// For other metrics, use the latest value
			if existing, exists := metricMap[m.Name]; !exists || m.Timestamp.After(existing.Timestamp) {
				metricMap[m.Name] = struct {
					Value     float64
					Timestamp time.Time
				}{m.Value, m.Timestamp}
			}
		}
	}

	// Convert back to original format
	metrics = nil
	for name, data := range metricMap {
		metrics = append(metrics, struct {
			Name      string    `json:"name"`
			Value     float64   `json:"value"`
			Timestamp time.Time `json:"timestamp"`
		}{name, data.Value, data.Timestamp})
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot fetch metrics"})
	}

	// Convert to expected format with proper field names
	result := make(map[string]interface{})
	var latestTimestamp time.Time
	for _, m := range metrics {
		switch m.Name {
		case "cpu_usage":
			result["cpu_usage"] = m.Value
		case "memory_usage":
			result["memory_usage"] = m.Value
		case "disk_usage":
			result["disk_usage"] = m.Value // Query already returns MIN value
		case "network_bytes":
			result["network_in"] = m.Value
			result["network_out"] = m.Value
		default:
			result[m.Name] = m.Value
		}
		if m.Timestamp.After(latestTimestamp) {
			latestTimestamp = m.Timestamp
		}
	}
	// Convert UTC to IST for display
	// Also include latest totals/used values from host_metrics table (if available)
	var hm struct {
		MemoryTotal float64   `json:"memory_total"`
		MemoryUsed  float64   `json:"memory_used"`
		DiskTotal   float64   `json:"disk_total"`
		DiskUsed    float64   `json:"disk_used"`
		Uptime      int64     `json:"uptime"`
		CPUUsage    float64   `json:"cpu_usage"`
		Timestamp   time.Time `json:"timestamp"`
	}
	// best-effort fetch - non-fatal
	_ = postgres.DB.Raw(`
		SELECT memory_total, memory_used, disk_total, disk_used, timestamp
		FROM host_metrics
		WHERE host_id = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`, hostID).Scan(&hm).Error
	if hm.MemoryTotal != 0 {
		result["memory_total"] = hm.MemoryTotal
	}
	if hm.MemoryUsed != 0 {
		result["memory_used"] = hm.MemoryUsed
	}
	if hm.DiskTotal != 0 {
		result["disk_total"] = hm.DiskTotal
	}
	if hm.DiskUsed != 0 {
		result["disk_used"] = hm.DiskUsed
	}
	if hm.Uptime != 0 {
		result["uptime"] = hm.Uptime
	}

	// If timeseries metrics did not supply a recent cpu_usage or it is zero,
	// fall back to the latest snapshot stored in host_metrics (if present).
	if hm.CPUUsage > 0 {
		if existing, ok := result["cpu_usage"].(float64); !ok || existing == 0 {
			result["cpu_usage"] = hm.CPUUsage
		}
	}
	if hm.Timestamp.After(latestTimestamp) {
		latestTimestamp = hm.Timestamp
	}

	// If no metrics were found in the metrics table, still return a result
	// derived from the latest host_metrics snapshot so the UI has a recent
	// value to display after a refresh.
	if len(metrics) == 0 {
		// If we have only host_metrics data, ensure timestamp is set
		if latestTimestamp.IsZero() && !hm.Timestamp.IsZero() {
			latestTimestamp = hm.Timestamp
		}
		// If still zero, return empty payload
		if latestTimestamp.IsZero() {
			return c.JSON(nil)
		}
	}

	// Canonicalize memory_usage and disk_usage using totals when available.
	// Prefer server-side computation so frontend and API consumers have a single source of truth.
	// If host_metrics contains totals, compute percentages and overwrite any existing metric values.
	if hm.MemoryTotal > 0 && hm.MemoryUsed >= 0 {
		computedMemUsage := (hm.MemoryUsed / hm.MemoryTotal) * 100.0
		// clamp to [0,100]
		if computedMemUsage < 0 {
			computedMemUsage = 0
		}
		if computedMemUsage > 100 {
			computedMemUsage = 100
		}
		// record computed value and flag
		result["memory_usage"] = computedMemUsage
		result["memory_usage_computed"] = true
		// if there was a prior memory_usage coming from metrics table, record the delta
		if prior, ok := result["memory_usage"].(float64); ok {
			diff := math.Abs(prior - computedMemUsage)
			result["memory_usage_diff"] = diff
			if diff >= 15.0 {
				result["memory_usage_warning"] = "Large discrepancy between snapshot and metric (>15%)"
			}
		}
	}
	if hm.DiskTotal > 0 && hm.DiskUsed >= 0 {
		computedDiskUsage := (hm.DiskUsed / hm.DiskTotal) * 100.0
		if computedDiskUsage < 0 {
			computedDiskUsage = 0
		}
		if computedDiskUsage > 100 {
			computedDiskUsage = 100
		}
		result["disk_usage"] = computedDiskUsage
		result["disk_usage_computed"] = true
		if prior, ok := result["disk_usage"].(float64); ok {
			diff := math.Abs(prior - computedDiskUsage)
			result["disk_usage_diff"] = diff
			if diff >= 15.0 {
				result["disk_usage_warning"] = "Large discrepancy between snapshot and metric (>15%)"
			}
		}
	}
	ist, _ := time.LoadLocation("Asia/Kolkata")
	result["timestamp"] = latestTimestamp.In(ist).Format(time.RFC3339)

	return c.JSON(result)
}

// GetHostMetricsTimeRange returns metrics for a host between start and end timestamps (RFC3339)
func GetHostMetricsTimeRange(c *fiber.Ctx) error {
	// Public endpoint - no auth required

	hostID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	startStr := c.Query("start")
	endStr := c.Query("end")

	var start, end time.Time
	var err error
	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid start timestamp"})
		}
	} else {
		// default to 24 hours ago
		start = time.Now().Add(-24 * time.Hour)
	}
	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid end timestamp"})
		}
	} else {
		end = time.Now()
	}

	var metrics []map[string]interface{}
	err = postgres.DB.Raw(`
		SELECT * FROM host_metrics WHERE host_id = ? AND timestamp BETWEEN ? AND ? ORDER BY timestamp ASC
	`, hostID, start, end).Scan(&metrics).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot fetch metrics"})
	}

	// Return empty array when no data
	if len(metrics) == 0 {
		return c.JSON([]map[string]interface{}{})
	}

	return c.JSON(metrics)
}

// GetHostDashboardMetrics returns summarized metrics for host dashboard visualization
func GetHostDashboardMetrics(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	hostID, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	var metrics []struct {
		CPUUsage    float64   `json:"cpu_usage"`
		MemoryUsage float64   `json:"memory_usage"`
		DiskUsage   float64   `json:"disk_usage"`
		NetworkIn   float64   `json:"network_in"`
		NetworkOut  float64   `json:"network_out"`
		Uptime      int64     `json:"uptime"`
		LoadAverage string    `json:"load_average"`
		Timestamp   time.Time `json:"timestamp"`
	}

	err := postgres.DB.Table("host_metrics").
		Where("host_id = ?", hostID).
		Order("timestamp DESC").
		Limit(100).
		Find(&metrics).Error

	if err != nil {
		// Return empty array on error - NO MOCK DATA
		return c.JSON([]map[string]interface{}{})
	}

	// Return empty array if no metrics found - NO MOCK DATA
	if len(metrics) == 0 {
		return c.JSON([]map[string]interface{}{})
	}

	return c.JSON(metrics)
}

// GetAllHostsLatestMetrics returns the latest metrics for all hosts (tenant-scoped when authenticated).
// Returns an array of objects: { host_id, cpu_usage, memory_usage, disk_usage, network_in, network_out, memory_total, memory_used, disk_total, disk_used, uptime, timestamp }
func GetAllHostsLatestMetrics(c *fiber.Ctx) error {
	// Determine tenant scope (0 = all)
	var tenantID int64 = 0
	if tid := c.Locals("tenant_id"); tid != nil {
		tenantID = tid.(int64)
	}

	// Determine host scope: if tenant-scoped, list hosts for the tenant and limit queries
	var hostIDs []int64
	if tenantID != 0 {
		hostsList, err := services.ListHosts(tenantID, 1000, 0)
		if err == nil && len(hostsList) > 0 {
			hostIDs = make([]int64, 0, len(hostsList))
			for _, h := range hostsList {
				hostIDs = append(hostIDs, h.ID)
			}
		}
	}

	// 1) load latest host_metrics snapshot per host
	var snaps []struct {
		HostID      int64     `json:"host_id"`
		MemoryTotal float64   `json:"memory_total"`
		MemoryUsed  float64   `json:"memory_used"`
		DiskTotal   float64   `json:"disk_total"`
		DiskUsed    float64   `json:"disk_used"`
		Uptime      int64     `json:"uptime"`
		Timestamp   time.Time `json:"timestamp"`
	}

	snapQuery := `SELECT DISTINCT ON (host_id) host_id, memory_total, memory_used, disk_total, disk_used, uptime, timestamp
		FROM host_metrics`
	if len(hostIDs) > 0 {
		snapQuery += ` WHERE host_id IN (?)`
	}
	snapQuery += ` ORDER BY host_id, timestamp DESC`

	if len(hostIDs) > 0 {
		if err := postgres.DB.Raw(snapQuery, hostIDs).Scan(&snaps).Error; err != nil {
			// non-fatal - continue with empty snapshots
			snaps = []struct {
				HostID      int64     `json:"host_id"`
				MemoryTotal float64   `json:"memory_total"`
				MemoryUsed  float64   `json:"memory_used"`
				DiskTotal   float64   `json:"disk_total"`
				DiskUsed    float64   `json:"disk_used"`
				Uptime      int64     `json:"uptime"`
				Timestamp   time.Time `json:"timestamp"`
			}{}
		}
	} else {
		if err := postgres.DB.Raw(snapQuery).Scan(&snaps).Error; err != nil {
			snaps = []struct {
				HostID      int64     `json:"host_id"`
				MemoryTotal float64   `json:"memory_total"`
				MemoryUsed  float64   `json:"memory_used"`
				DiskTotal   float64   `json:"disk_total"`
				DiskUsed    float64   `json:"disk_used"`
				Uptime      int64     `json:"uptime"`
				Timestamp   time.Time `json:"timestamp"`
			}{}
		}
	}

	// map host_id -> snapshot
	snapMap := make(map[int64]map[string]interface{})
	for _, s := range snaps {
		snapMap[s.HostID] = map[string]interface{}{
			"memory_total": s.MemoryTotal,
			"memory_used":  s.MemoryUsed,
			"disk_total":   s.DiskTotal,
			"disk_used":    s.DiskUsed,
			"uptime":       s.Uptime,
			"timestamp":    s.Timestamp,
		}
	}

	// 2) load latest metrics (timeseries) per host for last 5 minutes
	var rows []struct {
		HostID    int64     `json:"host_id"`
		Name      string    `json:"name"`
		Value     float64   `json:"value"`
		Timestamp time.Time `json:"timestamp"`
	}

	// Use a window function to pick the latest row per host+name within recent window
	metricsQuery := `WITH recent AS (
		SELECT host_id, name, value, timestamp,
		  ROW_NUMBER() OVER (PARTITION BY host_id, name ORDER BY timestamp DESC) AS rn
		FROM metrics WHERE timestamp > NOW() - INTERVAL '5 minutes'`
	if len(hostIDs) > 0 {
		metricsQuery += ` AND host_id IN (?)`
	}
	metricsQuery += ` ) SELECT host_id, name, value, timestamp FROM recent WHERE rn = 1`

	if len(hostIDs) > 0 {
		if err := postgres.DB.Raw(metricsQuery, hostIDs).Scan(&rows).Error; err != nil {
			rows = []struct {
				HostID    int64     `json:"host_id"`
				Name      string    `json:"name"`
				Value     float64   `json:"value"`
				Timestamp time.Time `json:"timestamp"`
			}{}
		}
	} else {
		if err := postgres.DB.Raw(metricsQuery).Scan(&rows).Error; err != nil {
			rows = []struct {
				HostID    int64     `json:"host_id"`
				Name      string    `json:"name"`
				Value     float64   `json:"value"`
				Timestamp time.Time `json:"timestamp"`
			}{}
		}
	}

	// assemble per-host map
	resultMap := make(map[int64]map[string]interface{})

	// seed with snapshots
	for hid, m := range snapMap {
		resultMap[hid] = map[string]interface{}{}
		for k, v := range m {
			resultMap[hid][k] = v
		}
	}

	// merge timeseries metrics
	for _, r := range rows {
		m, ok := resultMap[r.HostID]
		if !ok {
			m = map[string]interface{}{}
			resultMap[r.HostID] = m
		}
		switch r.Name {
		case "cpu_usage":
			m["cpu_usage"] = r.Value
		case "memory_usage":
			m["memory_usage"] = r.Value
		case "disk_usage":
			m["disk_usage"] = r.Value
		case "network_bytes":
			// approximate split for display - clients may interpret differently
			m["network_in"] = r.Value
			m["network_out"] = r.Value
		default:
			m[r.Name] = r.Value
		}
		// prefer latest timestamp
		if existingTs, ok := m["timestamp"].(time.Time); !ok || r.Timestamp.After(existingTs) {
			m["timestamp"] = r.Timestamp
		}
	}

	// If cpu_usage is missing or zero after merging recent timeseries rows,
	// try one more fallback per-host: query the latest non-zero cpu_usage from the
	// metrics table (last 10 minutes) and use it.
	for hid, m := range resultMap {
		if v, exists := m["cpu_usage"]; !exists || v == 0 {
			var lastCPU struct {
				Value     float64   `db:"value"`
				Timestamp time.Time `db:"timestamp"`
			}
			err := postgres.DB.Raw(`
				SELECT value, timestamp FROM metrics
				WHERE host_id = ? AND name = 'cpu_usage' AND value > 0
				AND timestamp > NOW() - INTERVAL '10 minutes'
				ORDER BY timestamp DESC LIMIT 1
			`, hid).Scan(&lastCPU).Error
			if err == nil && lastCPU.Value > 0 {
				m["cpu_usage"] = lastCPU.Value
				if ts, ok := m["timestamp"].(time.Time); !ok || lastCPU.Timestamp.After(ts) {
					m["timestamp"] = lastCPU.Timestamp
				}
			}
		}
		resultMap[hid] = m
	}

	// Convert to slice of host metrics
	out := make([]map[string]interface{}, 0, len(resultMap))
	for hid, m := range resultMap {
		m["host_id"] = hid
		// Ensure timestamp formatted as RFC3339 for frontend
		if ts, ok := m["timestamp"].(time.Time); ok {
			m["timestamp"] = ts.Format(time.RFC3339)
		}
		out = append(out, m)
	}

	return c.JSON(out)
}

// CollectHostNow triggers an immediate collection for the given host id (debug/testing).
func CollectHostNow(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}
	tenantID := tid.(int64)

	hostID, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	host, err := postgres.GetHost(postgres.DB, hostID)
	if err != nil || host.TenantID != tenantID {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	metric, err := services.CollectHostMetrics(host)
	if err != nil {
		// mark host offline and return error
		if uerr := services.UpdateHostStatus(host.ID, "offline"); uerr != nil {
			logging.Warnf("failed to mark host %d offline after collect error: %v", host.ID, uerr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "collect failed", "detail": err.Error()})
	}

	if err := services.SaveHostMetrics(metric); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "save failed", "detail": err.Error()})
	}

	if uerr := services.UpdateHostStatus(host.ID, "online"); uerr != nil {
		logging.Warnf("failed to mark host %d online after collect: %v", host.ID, uerr)
	}

	return c.JSON(fiber.Map{"success": true, "metric": metric})
}

// CreateHostWithAgent creates a host with simple agent setup
func CreateHostWithAgent(c *fiber.Ctx) error {
	var req models.Host
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	req.TenantID = tid.(int64)
	req.AgentStatus = "pending"
	req.LastSeen = time.Now()

	if err := postgres.CreateHost(postgres.DB, &req); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.Status(409).JSON(fiber.Map{"error": "Host already exists"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Cannot create host"})
	}

	return c.JSON(fiber.Map{
		"host_id":         req.ID,
		"message":         "Host created. Install agent using the provided instructions.",
		"install_command": fmt.Sprintf("curl -sSL https://install.kineticops.com/agent.sh | sudo bash -s -- --host=%s --token=<your-token>", req.IP),
	})
}

// CleanupFailedHost removes host if creation failed
func CleanupFailedHost(hostname string) {
	hosts, err := postgres.ListHosts(postgres.DB, 0, 100, 0)
	if err == nil {
		for _, h := range hosts {
			if h.Hostname == hostname && (h.AgentStatus == "installing" || h.AgentStatus == "failed") {
				if derr := postgres.DeleteHost(postgres.DB, h.ID); derr != nil {
					logging.Warnf("failed to cleanup failed host %s (id=%d): %v", h.Hostname, h.ID, derr)
				}
				break
			}
		}
	}
}
