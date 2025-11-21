package handlers

import (
	"context"
	"encoding/json"
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
	"github.com/sakkurohilla/kineticops/backend/internal/telemetry"
	ws "github.com/sakkurohilla/kineticops/backend/internal/websocket"
)

type AgentEvent struct {
	Timestamp string                 `json:"@timestamp"`
	Agent     map[string]interface{} `json:"agent"`
	Host      map[string]interface{} `json:"host"`
	Event     map[string]interface{} `json:"event"`
	Log       map[string]interface{} `json:"log"`
	Message   string                 `json:"message"`
	System    map[string]interface{} `json:"system"`
}

func ReceiveAgentData(c *fiber.Ctx) error {
	var payload struct {
		Events []AgentEvent `json:"events"`
	}

	if err := c.BodyParser(&payload); err != nil {
		// Log parse error for diagnostics (do not echo payload contents)
		logging.Warnf("agent data parse error: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	if len(payload.Events) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No events in payload"})
	}

	// Extract tenant info from headers or token
	tenantID := extractTenantID(c)
	if tenantID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid tenant authentication"})
	}

	processedCount := 0
	for _, event := range payload.Events {
		if processEvent(&event, tenantID) {
			processedCount++
		}
	}

	return c.JSON(fiber.Map{
		"message":   "Events processed",
		"processed": processedCount,
		"total":     len(payload.Events),
	})
}

func processEvent(event *AgentEvent, tenantID int64) bool {
	// Validate required fields
	hostData := event.Host
	if hostData == nil {
		return false
	}

	hostname, _ := hostData["hostname"].(string)
	if hostname == "" || len(hostname) < 2 {
		return false
	}

	// Extract all host information
	primaryIP, _ := hostData["primary_ip"].(string)
	os, _ := hostData["os"].(string)
	platform, _ := hostData["platform"].(string)
	platformFamily, _ := hostData["platform_family"].(string)
	platformVersion, _ := hostData["platform_version"].(string)
	arch, _ := hostData["arch"].(string)
	kernelVersion, _ := hostData["kernel_version"].(string)
	virtualization, _ := hostData["virtualization"].(string)

	host := findOrCreateHost(hostname, primaryIP, os, platform, platformFamily,
		platformVersion, arch, kernelVersion, virtualization, tenantID)
	if host == nil {
		return false
	}

	// Update host last seen and status
	now := time.Now().UTC()
	if res := postgres.DB.Model(&models.Host{}).Where("id = ?", host.ID).Updates(map[string]interface{}{
		"last_seen":    now,
		"agent_status": "online",
	}); res.Error != nil {
		logging.Warnf("failed to update last_seen/agent_status for host=%d: %v", host.ID, res.Error)
	}

	// Process system metrics with validation
	if event.System != nil {
		processSystemMetrics(host.ID, host.TenantID, event.System, now)
		// Also check for embedded log in the same event
		if event.Log != nil || event.Message != "" {
			processAgentLog(event, host.ID, tenantID)
		}
		return true
	}

	// If this event carried a log but no system block, try to process it
	if event.Log != nil || event.Message != "" {
		processAgentLog(event, host.ID, tenantID)
		return true
	}

	return false
}

// processAgentLog converts an AgentEvent carrying a log into models.Log and stores it.
func processAgentLog(event *AgentEvent, hostID, tenantID int64) {
	// Build log model from various possible fields
	l := &models.Log{
		TenantID: tenantID,
		HostID:   hostID,
	}

	// Timestamp
	if event.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, event.Timestamp); err == nil {
			l.Timestamp = t
		} else {
			l.Timestamp = time.Now().UTC()
		}
	} else {
		l.Timestamp = time.Now().UTC()
	}

	// Message
	if event.Message != "" {
		l.Message = event.Message
	} else if msg, ok := event.Log["message"].(string); ok {
		l.Message = msg
	}

	// Level
	if lvl, ok := event.Log["level"].(string); ok {
		l.Level = lvl
	} else if t, ok := event.Event["type"].(string); ok {
		l.Level = t
	}

	// Correlation id
	if cid, ok := event.Log["correlation_id"].(string); ok {
		l.CorrelID = cid
	} else if cid2, ok := event.Event["correlation_id"].(string); ok {
		l.CorrelID = cid2
	}

	// Meta / fields
	l.Meta = make(map[string]string)
	if fields, ok := event.Log["fields"].(map[string]interface{}); ok {
		for k, v := range fields {
			l.Meta[k] = fmt.Sprintf("%v", v)
		}
	}
	// fallback to top-level event fields
	if len(l.Meta) == 0 {
		if evfields, ok := event.Event["fields"].(map[string]interface{}); ok {
			for k, v := range evfields {
				l.Meta[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	// Persist via service (parses/enriches and inserts into MongoDB)
	// Read HostID field here to avoid staticcheck reporting unused write when
	// the field is consumed reflectively by downstream code (e.g. JSON/Mongo).
	_ = l.HostID
	if err := services.CollectLog(context.Background(), l); err != nil {
		// don't fail the whole pipeline for logging issues; record and continue
		logging.Warnf("failed to persist agent log for host=%d: %v", hostID, err)
		return
	}

	// Broadcast to websocket clients for real-time tailing
	payload := map[string]interface{}{"type": "log", "log": l}
	if b, err := json.Marshal(payload); err == nil {
		if gh := ws.GetGlobalHub(); gh != nil {
			gh.RememberMessage(b)
		}
		ws.BroadcastToClients(b)
		telemetry.IncWSBroadcast(context.Background(), 1)
	}
}

func findOrCreateHost(hostname, primaryIP, os, platform, platformFamily,
	platformVersion, arch, kernelVersion, virtualization string,
	tenantID int64) *models.Host {

	var host models.Host

	// Multi-layer host identification
	host = findHostByMultipleIdentifiers(hostname, primaryIP, tenantID)
	if host.ID != 0 {
		// Update host information if changed
		updates := make(map[string]interface{})
		if host.IP != primaryIP && primaryIP != "" {
			updates["ip"] = primaryIP
		}
		// Only overwrite hostname if host was auto-created (reg_token starts with "auto-")
		// or if current hostname is empty. This prevents agent auto-updates from
		// reverting user-edited hostnames.
		if hostname != "" {
			if host.Hostname == "" || strings.HasPrefix(host.RegToken, "auto-") {
				if host.Hostname != hostname {
					updates["hostname"] = hostname
				}
			}
		}
		if host.OS != os && os != "" {
			updates["os"] = os
		}
		if len(updates) > 0 {
			if res := postgres.DB.Model(&host).Updates(updates); res.Error != nil {
				logging.Warnf("failed to update host info for host=%d: %v", host.ID, res.Error)
			}
		}
		return &host
	}

	// Validate required fields for new host
	if primaryIP == "" {
		return nil
	}
	if os == "" {
		os = "linux"
	}

	// Create new host with complete information
	// Prevent re-creation if the host was recently deleted (tombstone exists)
	var tombstoneCount int
	// check deleted_hosts for entries in last 7 days
	if terr := postgres.DB.Raw("SELECT count(*) FROM deleted_hosts WHERE hostname = ? AND tenant_id = ? AND deleted_at > NOW() - INTERVAL '7 days'", hostname, tenantID).Scan(&tombstoneCount).Error; terr != nil {
		logging.Warnf("failed to check deleted_hosts tombstones for hostname=%s tenant=%d: %v", hostname, tenantID, terr)
	}
	if tombstoneCount > 0 {
		logging.Infof("Skipping auto-creation for recently deleted host %s (tenant=%d)", hostname, tenantID)
		return nil
	}

	now := time.Now().UTC()
	host = models.Host{
		Hostname:    hostname,
		IP:          primaryIP,
		OS:          os,
		AgentStatus: "online",
		Status:      "active",
		TenantID:    tenantID,
		Group:       "auto-discovered",
		// Include additional platform details to avoid unused-parameter linter warnings
		Description: fmt.Sprintf("Auto-discovered %s host on %s (ver=%s arch=%s kernel=%s virt=%s)", platform, platformFamily, platformVersion, arch, kernelVersion, virtualization),
		LastSeen:    now,
		RegToken:    fmt.Sprintf("auto-%d-%s-%d", tenantID, hostname, now.Unix()),
	}

	if err := postgres.DB.Create(&host).Error; err != nil {
		logging.Warnf("failed to create auto-discovered host %s tenant=%d: %v", hostname, tenantID, err)
		return nil
	}

	return &host
}

// findHostByMultipleIdentifiers implements multi-layer host identification
func findHostByMultipleIdentifiers(hostname, ip string, tenantID int64) models.Host {
	var host models.Host

	// 1. PRIMARY: Try by reg_token (most reliable)
	token := fmt.Sprintf("auto-%d-%s", tenantID, hostname)
	err := postgres.DB.Where("reg_token LIKE ? AND tenant_id = ?", token+"%", tenantID).First(&host).Error
	if err == nil {
		return host
	}

	// 2. SECONDARY: Try by IP + tenant (network stable)
	if ip != "" {
		err = postgres.DB.Where("ip = ? AND tenant_id = ?", ip, tenantID).First(&host).Error
		if err == nil {
			return host
		}
	}

	// 3. FALLBACK: Try by hostname + tenant (legacy support)
	if hostname != "" {
		err = postgres.DB.Where("hostname = ? AND tenant_id = ?", hostname, tenantID).First(&host).Error
		if err == nil {
			return host
		}
	}

	// 4. LAST RESORT: Try by hostname pattern matching (for renamed hosts)
	if hostname != "" {
		err = postgres.DB.Where("hostname LIKE ? AND tenant_id = ?", hostname+"%", tenantID).First(&host).Error
		if err == nil {
			return host
		}
	}

	return models.Host{} // Not found
}

// extractTenantID extracts tenant ID from request headers or token
func extractTenantID(c *fiber.Ctx) int64 {
	// Try to get from X-Tenant-ID header
	if tenantHeader := c.Get("X-Tenant-ID"); tenantHeader != "" {
		var tenantID int64
		if n, err := fmt.Sscanf(tenantHeader, "%d", &tenantID); err == nil && n == 1 {
			return tenantID
		}
	}

	// Try to get from Authorization token
	token := c.Get("Authorization")
	if token != "" && len(token) > 7 && token[:7] == "Bearer " {
		installToken, err := services.ResolveUserFromInstallationToken(token[7:])
		if err == nil && !time.Now().After(installToken.ExpiresAt) {
			return int64(installToken.TenantID)
		}
	}

	// Default tenant for development/testing
	return 2
}

func processSystemMetrics(hostID, tenantID int64, system map[string]interface{}, timestamp time.Time) {
	metric := &models.HostMetric{Timestamp: timestamp}

	// CPU metrics with validation
	if cpu, ok := system["cpu"].(map[string]interface{}); ok {
		if total, ok := cpu["total"].(map[string]interface{}); ok {
			if pct, ok := total["pct"].(float64); ok && pct >= 0 && pct <= 1 {
				metric.CPUUsage = pct * 100
				logging.Infof("[METRICS] CPU for host=%d: usage=%.2f%%", hostID, metric.CPUUsage)
				if err := services.CollectMetric(hostID, tenantID, "cpu_usage", metric.CPUUsage, nil); err != nil {
					logging.Errorf("CollectMetric(cpu_usage) failed host=%d: %v", hostID, err)
				}
			}
		}
	}

	// Memory metrics with validation
	if memory, ok := system["memory"].(map[string]interface{}); ok {
		// Prefer bytes-based calculation when possible. Accept both map{"bytes":...}
		// shapes (new agent) and plain float64 values (older agents). Normalize
		// into totalBytes and usedBytes and compute a single canonical percentage.
		var totalBytes float64 = 0
		if total, ok := memory["total"].(float64); ok && total > 0 {
			totalBytes = total
			metric.MemoryTotal = total / (1024 * 1024) // MB
		}

		var usedBytes float64 = 0

		// 1) Try available.bytes (preferred)
		if availRaw, ok := memory["available"]; ok {
			switch v := availRaw.(type) {
			case map[string]interface{}:
				if ab, ok := v["bytes"].(float64); ok && ab >= 0 && totalBytes > 0 {
					usedBytes = totalBytes - ab
				}
			case float64:
				if v >= 0 && totalBytes > 0 {
					usedBytes = totalBytes - v
				}
			}
		}

		// 2) Fallback to free (either map or float)
		if usedBytes == 0 {
			if freeRaw, ok := memory["free"]; ok {
				switch v := freeRaw.(type) {
				case map[string]interface{}:
					if fb, ok := v["bytes"].(float64); ok && fb >= 0 && totalBytes > 0 {
						usedBytes = totalBytes - fb
					}
				case float64:
					if v >= 0 && totalBytes > 0 {
						usedBytes = totalBytes - v
					}
				}
			}
		}

		// 3) Fallback to used.bytes if still zero
		if usedBytes == 0 {
			if usedRaw, ok := memory["used"]; ok {
				if usedMap, ok := usedRaw.(map[string]interface{}); ok {
					if ub, ok := usedMap["bytes"].(float64); ok && ub >= 0 {
						usedBytes = ub
					}
				}
			}
		}

		// 4) Finally, if used.pct provided and we don't have numeric usedBytes/total, use pct
		var pctFromAgent float64 = -1
		if usedRaw, ok := memory["used"]; ok {
			if usedMap, ok := usedRaw.(map[string]interface{}); ok {
				if pct, ok := usedMap["pct"].(float64); ok && pct >= 0 {
					pctFromAgent = pct
				}
			}
		}

		// If we have both totalBytes and usedBytes, compute canonical values
		if usedBytes > 0 && totalBytes > 0 {
			// clamp
			if usedBytes < 0 {
				usedBytes = 0
			}
			if usedBytes > totalBytes {
				usedBytes = totalBytes
			}
			metric.MemoryUsed = usedBytes / (1024 * 1024)   // MB
			metric.MemoryTotal = totalBytes / (1024 * 1024) // MB (ensure set)
			// compute percent
			metric.MemoryUsage = (usedBytes / totalBytes) * 100
		} else if pctFromAgent >= 0 {
			// Use agent percentage when bytes are not available
			metric.MemoryUsage = pctFromAgent * 100
		}

		// Extract memory_free from agent data
		if freeRaw, ok := memory["free"]; ok {
			switch v := freeRaw.(type) {
			case float64:
				if v >= 0 {
					metric.MemoryFree = v / (1024 * 1024) // Convert to MB
				}
			case map[string]interface{}:
				if fb, ok := v["bytes"].(float64); ok && fb >= 0 {
					metric.MemoryFree = fb / (1024 * 1024) // Convert to MB
				}
			}
		}

		// Persist memory metrics (usage percentage, total MB, used MB, free MB)
		if metric.MemoryUsage >= 0 {
			logging.Infof("[METRICS] Memory for host=%d: usage=%.2f%%, total=%.2fMB, used=%.2fMB, free=%.2fMB",
				hostID, metric.MemoryUsage, metric.MemoryTotal, metric.MemoryUsed, metric.MemoryFree)
			if err := services.CollectMetric(hostID, tenantID, "memory_usage", metric.MemoryUsage, nil); err != nil {
				logging.Errorf("CollectMetric(memory_usage) failed host=%d: %v", hostID, err)
			}
		}
		if metric.MemoryTotal > 0 {
			if err := services.CollectMetric(hostID, tenantID, "memory_total", metric.MemoryTotal, nil); err != nil {
				logging.Errorf("CollectMetric(memory_total) failed host=%d: %v", hostID, err)
			}
		}
		if metric.MemoryUsed > 0 {
			if err := services.CollectMetric(hostID, tenantID, "memory_used", metric.MemoryUsed, nil); err != nil {
				logging.Errorf("CollectMetric(memory_used) failed host=%d: %v", hostID, err)
			}
		}
		if metric.MemoryFree > 0 {
			if err := services.CollectMetric(hostID, tenantID, "memory_free", metric.MemoryFree, nil); err != nil {
				logging.Errorf("CollectMetric(memory_free) failed host=%d: %v", hostID, err)
			}
		}
	}

	// Disk metrics with validation - ONLY root filesystem
	if fs, ok := system["filesystem"].(map[string]interface{}); ok {
		// STRICT: Only process root filesystem
		mountPoint, _ := fs["mount_point"].(string)
		deviceName, _ := fs["device_name"].(string)

		// Skip if not root filesystem
		if mountPoint != "/" {
			logging.Infof("Skipping filesystem %s mounted at %s", deviceName, mountPoint)
			return
		}

		logging.Infof("Processing root filesystem %s mounted at %s", deviceName, mountPoint)

		if used, ok := fs["used"].(map[string]interface{}); ok {
			// agent may provide pct (0..1 or 0..100) and/or raw bytes
			var agentPctVal float64 = -1
			if pct, ok := used["pct"].(float64); ok {
				agentPctVal = pct
			}

			var usedBytes float64 = 0
			if bytes, ok := used["bytes"].(float64); ok && bytes >= 0 {
				usedBytes = bytes
				metric.DiskUsed = bytes / (1024 * 1024 * 1024) // GB
			}

			var totalBytes float64 = 0
			if total, ok := fs["total"].(float64); ok && total > 0 {
				totalBytes = total
				metric.DiskTotal = total / (1024 * 1024 * 1024) // GB
			}

			// Prefer bytes-based calculation when available
			if usedBytes > 0 && totalBytes > 0 {
				serverPct := (usedBytes / totalBytes) * 100.0
				metric.DiskUsage = serverPct
				if err := services.CollectMetric(hostID, tenantID, "disk_usage", metric.DiskUsage, nil); err != nil {
					logging.Errorf("CollectMetric(disk_usage) failed host=%d: %v", hostID, err)
				}
				// Store disk_total and disk_used in GB
				if metric.DiskTotal > 0 {
					if err := services.CollectMetric(hostID, tenantID, "disk_total", metric.DiskTotal, nil); err != nil {
						logging.Errorf("CollectMetric(disk_total) failed host=%d: %v", hostID, err)
					}
				}
				if metric.DiskUsed > 0 {
					if err := services.CollectMetric(hostID, tenantID, "disk_used", metric.DiskUsed, nil); err != nil {
						logging.Errorf("CollectMetric(disk_used) failed host=%d: %v", hostID, err)
					}
				}
				// If agent also reported pct, normalize and compare
				if agentPctVal >= 0 {
					var agentPctPercent float64
					if agentPctVal <= 1 {
						agentPctPercent = agentPctVal * 100
					} else {
						agentPctPercent = agentPctVal
					}
					if math.Abs(agentPctPercent-serverPct) > 15.0 {
						logging.Warnf("Disk pct mismatch host=%d device=%s agent_pct=%.2f server_pct=%.2f", hostID, deviceName, agentPctPercent, serverPct)
					}
				}
				logging.Infof("Root disk usage (bytes) for %s mounted at %s: %.2f%%", deviceName, mountPoint, metric.DiskUsage)
			} else {
				// Fallback to agent-provided pct when bytes not available
				if agentPctVal >= 0 {
					var agentPctPercent float64
					if agentPctVal <= 1 {
						agentPctPercent = agentPctVal * 100
					} else {
						agentPctPercent = agentPctVal
					}
					metric.DiskUsage = agentPctPercent
					if err := services.CollectMetric(hostID, tenantID, "disk_usage", metric.DiskUsage, nil); err != nil {
						logging.Errorf("CollectMetric(disk_usage) failed host=%d: %v", hostID, err)
					}
					logging.Infof("Root disk usage (agent pct) for %s mounted at %s: %.2f%%", deviceName, mountPoint, metric.DiskUsage)
				}
			}
		} else {
			// No used block provided - try total only (unlikely) or skip
			if total, ok := fs["total"].(float64); ok && total > 0 {
				metric.DiskTotal = total / (1024 * 1024 * 1024) // GB
			}
		}
	}

	// Network metrics with validation
	if net, ok := system["network"].(map[string]interface{}); ok {
		var networkInBytes, networkOutBytes float64
		if in, ok := net["in"].(map[string]interface{}); ok {
			if bytes, ok := in["bytes"].(float64); ok && bytes >= 0 {
				networkInBytes = bytes
				metric.NetworkIn = bytes / (1024 * 1024) // Convert to MB
			}
		}
		if out, ok := net["out"].(map[string]interface{}); ok {
			if bytes, ok := out["bytes"].(float64); ok && bytes >= 0 {
				networkOutBytes = bytes
				metric.NetworkOut = bytes / (1024 * 1024) // Convert to MB
			}
		}

		// Store network metrics as time-series for historical analysis
		if networkInBytes > 0 {
			if err := services.CollectMetric(hostID, tenantID, "network_in_bytes", networkInBytes, nil); err != nil {
				logging.Errorf("CollectMetric(network_in_bytes) failed host=%d: %v", hostID, err)
			}
		}
		if networkOutBytes > 0 {
			if err := services.CollectMetric(hostID, tenantID, "network_out_bytes", networkOutBytes, nil); err != nil {
				logging.Errorf("CollectMetric(network_out_bytes) failed host=%d: %v", hostID, err)
			}
		}
		// Combined network throughput metric
		totalNetworkBytes := networkInBytes + networkOutBytes
		if totalNetworkBytes > 0 {
			if err := services.CollectMetric(hostID, tenantID, "network_bytes", totalNetworkBytes, nil); err != nil {
				logging.Errorf("CollectMetric(network_bytes) failed host=%d: %v", hostID, err)
			}
		}
	}

	// Disk I/O metrics with validation
	if diskio, ok := system["diskio"].(map[string]interface{}); ok {
		device, _ := diskio["device"].(string)

		if readBytes, ok := diskio["read_bytes"].(float64); ok && readBytes >= 0 {
			metric.DiskReadBytes = readBytes
			if err := services.CollectMetric(hostID, tenantID, "disk_read_bytes", readBytes, nil); err != nil {
				logging.Errorf("CollectMetric(disk_read_bytes) failed host=%d: %v", hostID, err)
			}
		}

		if writeBytes, ok := diskio["write_bytes"].(float64); ok && writeBytes >= 0 {
			metric.DiskWriteBytes = writeBytes
			if err := services.CollectMetric(hostID, tenantID, "disk_write_bytes", writeBytes, nil); err != nil {
				logging.Errorf("CollectMetric(disk_write_bytes) failed host=%d: %v", hostID, err)
			}
		}

		if readSpeed, ok := diskio["read_speed"].(float64); ok && readSpeed >= 0 {
			metric.DiskReadSpeed = readSpeed
			if err := services.CollectMetric(hostID, tenantID, "disk_read_speed", readSpeed, nil); err != nil {
				logging.Errorf("CollectMetric(disk_read_speed) failed host=%d: %v", hostID, err)
			}
		}

		if writeSpeed, ok := diskio["write_speed"].(float64); ok && writeSpeed >= 0 {
			metric.DiskWriteSpeed = writeSpeed
			if err := services.CollectMetric(hostID, tenantID, "disk_write_speed", writeSpeed, nil); err != nil {
				logging.Errorf("CollectMetric(disk_write_speed) failed host=%d: %v", hostID, err)
			}
		}

		if device != "" {
			logging.Infof("Disk I/O metrics collected for device %s: read=%.2f MB/s, write=%.2f MB/s",
				device, metric.DiskReadSpeed, metric.DiskWriteSpeed)
		}
	}

	// Load average with validation
	if load, ok := system["load"].(map[string]interface{}); ok {
		if load1, ok := load["1"].(float64); ok && load1 >= 0 {
			if load5, ok := load["5"].(float64); ok && load5 >= 0 {
				if load15, ok := load["15"].(float64); ok && load15 >= 0 {
					metric.LoadAverage = fmt.Sprintf("%.2f %.2f %.2f", load1, load5, load15)

					// Store load average metrics for time-series analysis
					if err := services.CollectMetric(hostID, tenantID, "load_1min", load1, nil); err != nil {
						logging.Errorf("CollectMetric(load_1min) failed host=%d: %v", hostID, err)
					}
					if err := services.CollectMetric(hostID, tenantID, "load_5min", load5, nil); err != nil {
						logging.Errorf("CollectMetric(load_5min) failed host=%d: %v", hostID, err)
					}
					if err := services.CollectMetric(hostID, tenantID, "load_15min", load15, nil); err != nil {
						logging.Errorf("CollectMetric(load_15min) failed host=%d: %v", hostID, err)
					}
				}
			}
		}
	} else {
		metric.LoadAverage = "0.00 0.00 0.00" // Default load average
	}

	// CRITICAL: Real uptime validation
	if uptimeVal, ok := system["uptime"]; ok {
		switch v := uptimeVal.(type) {
		case float64:
			if v > 0 {
				metric.Uptime = int64(v)
			} else {
				metric.Uptime = 0
			}
		case string:
			// Accept numeric uptime sent as a string (common in shell-based agents)
			if v == "unavailable" || v == "" {
				metric.Uptime = 0
			} else {
				if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
					metric.Uptime = int64(f)
				} else {
					metric.Uptime = 0
				}
			}
		default:
			metric.Uptime = 0
		}

		// Store uptime as a metric for historical tracking
		if metric.Uptime > 0 {
			if err := services.CollectMetric(hostID, tenantID, "uptime_seconds", float64(metric.Uptime), nil); err != nil {
				logging.Errorf("CollectMetric(uptime_seconds) failed host=%d: %v", hostID, err)
			}
		}
	} else {
		metric.Uptime = 0
	}

	// Process metrics handling
	if processesData, ok := system["processes"].(map[string]interface{}); ok {
		logging.Infof("[PROCESSES] Processing process metrics for host=%d", hostID)

		// Handle top CPU processes
		if topCPU, ok := processesData["top_cpu"].([]interface{}); ok && len(topCPU) > 0 {
			logging.Infof("[PROCESSES] Storing %d top CPU processes", len(topCPU))
			for _, procData := range topCPU {
				if proc, ok := procData.(map[string]interface{}); ok {
					// Extract PID - handle both int and float64
					var pid int
					switch v := proc["pid"].(type) {
					case float64:
						pid = int(v)
					case int:
						pid = v
					case int32:
						pid = int(v)
					case int64:
						pid = int(v)
					default:
						logging.Errorf("[PROCESSES] Unexpected PID type: %T value: %v", proc["pid"], proc["pid"])
						continue
					}

					if pid == 0 {
						logging.Errorf("[PROCESSES] Got PID=0, skipping. Raw data: %+v", proc)
						continue
					}

					name, _ := proc["name"].(string)
					username, _ := proc["username"].(string)
					cpuPercent, _ := proc["cpu_percent"].(float64)
					memoryPercent, _ := proc["memory_percent"].(float64)
					memoryRSS, _ := proc["memory_rss"].(float64)
					status, _ := proc["status"].(string)
					numThreads, _ := proc["num_threads"].(float64)
					createTime, _ := proc["create_time"].(float64)

					// Insert process metric
					err := postgres.DB.Exec(`
						INSERT INTO process_metrics (host_id, tenant_id, pid, name, username, cpu_percent, 
							memory_percent, memory_rss, status, num_threads, create_time, timestamp)
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
					`, hostID, tenantID, pid, name, username, cpuPercent, memoryPercent,
						int64(memoryRSS), status, int(numThreads), int64(createTime)).Error

					if err != nil {
						logging.Errorf("[PROCESSES] Failed to insert process metric: %v", err)
					}
				}
			}
		}

		// Handle top memory processes (optional, can be same as CPU)
		if topMemory, ok := processesData["top_memory"].([]interface{}); ok && len(topMemory) > 0 {
			logging.Infof("[PROCESSES] Received %d top memory processes", len(topMemory))
		}
	}

	// Detect and skip placeholder frames (all zero/empty core metrics)
	isPlaceholder := metric.CPUUsage == 0 && metric.MemoryUsage == 0 && metric.DiskUsage == 0 &&
		metric.MemoryTotal == 0 && metric.MemoryUsed == 0 && metric.NetworkIn == 0 && metric.NetworkOut == 0 &&
		(metric.LoadAverage == "" || metric.LoadAverage == "0.00 0.00 0.00") && metric.Uptime == 0

	if isPlaceholder {
		// Fallback: retrieve last stored metrics for this host to avoid broadcasting zeros
		var prev models.HostMetric
		err := postgres.DB.Raw(`SELECT cpu_usage, memory_usage, memory_total, memory_used, memory_free,
			disk_usage, disk_total, disk_used, disk_read_bytes, disk_write_bytes, disk_read_speed, disk_write_speed,
			network_in, network_out, uptime, load_average, timestamp FROM host_metrics WHERE host_id = ?`, hostID).Scan(&prev).Error
		if err != nil {
			logging.Warnf("[METRICS] Placeholder frame host=%d and no previous metrics found (err=%v) - suppress broadcast", hostID, err)
			return
		}
		if prev.Timestamp.IsZero() {
			logging.Warnf("[METRICS] Placeholder frame host=%d and previous metrics empty - suppress broadcast", hostID)
			return
		}
		logging.Infof("[METRICS] Placeholder frame host=%d - broadcasting last known metrics instead", hostID)
		payload := map[string]interface{}{
			"type":             "metric",
			"host_id":          hostID,
			"seq":              uint64(time.Now().UnixNano()),
			"cpu_usage":        prev.CPUUsage,
			"memory_usage":     prev.MemoryUsage,
			"memory_total":     prev.MemoryTotal,
			"memory_used":      prev.MemoryUsed,
			"memory_free":      prev.MemoryFree,
			"disk_usage":       prev.DiskUsage,
			"disk_total":       prev.DiskTotal,
			"disk_used":        prev.DiskUsed,
			"disk_read_bytes":  prev.DiskReadBytes,
			"disk_write_bytes": prev.DiskWriteBytes,
			"disk_read_speed":  prev.DiskReadSpeed,
			"disk_write_speed": prev.DiskWriteSpeed,
			"network_in":       prev.NetworkIn,
			"network_out":      prev.NetworkOut,
			"uptime":           prev.Uptime,
			"load_average":     prev.LoadAverage,
			"timestamp":        prev.Timestamp.UTC().Format(time.RFC3339),
			"fallback":         true,
		}
		if b, jerr := json.Marshal(payload); jerr == nil {
			if gh := ws.GetGlobalHub(); gh != nil {
				gh.RememberMessage(b)
			}
			ws.BroadcastToClients(b)
			telemetry.IncWSBroadcast(context.Background(), 1)
		}
		return
	} else {
		// Store in host_metrics table with error handling
		err := postgres.DB.Exec(`
			INSERT INTO host_metrics (host_id, cpu_usage, memory_usage, memory_total, memory_used, memory_free,
				disk_usage, disk_total, disk_used, disk_read_bytes, disk_write_bytes, disk_read_speed, disk_write_speed,
				network_in, network_out, uptime, load_average, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
			ON CONFLICT (host_id) DO UPDATE SET
				cpu_usage = EXCLUDED.cpu_usage,
				memory_usage = EXCLUDED.memory_usage,
				memory_total = EXCLUDED.memory_total,
				memory_used = EXCLUDED.memory_used,
				memory_free = EXCLUDED.memory_free,
				disk_usage = EXCLUDED.disk_usage,
				disk_total = EXCLUDED.disk_total,
				disk_used = EXCLUDED.disk_used,
				disk_read_bytes = EXCLUDED.disk_read_bytes,
				disk_write_bytes = EXCLUDED.disk_write_bytes,
				disk_read_speed = EXCLUDED.disk_read_speed,
				disk_write_speed = EXCLUDED.disk_write_speed,
				network_in = EXCLUDED.network_in,
				network_out = EXCLUDED.network_out,
				uptime = EXCLUDED.uptime,
				load_average = EXCLUDED.load_average,
				timestamp = EXCLUDED.timestamp
		`, hostID, metric.CPUUsage, metric.MemoryUsage, metric.MemoryTotal, metric.MemoryUsed, metric.MemoryFree,
			metric.DiskUsage, metric.DiskTotal, metric.DiskUsed, metric.DiskReadBytes, metric.DiskWriteBytes,
			metric.DiskReadSpeed, metric.DiskWriteSpeed, metric.NetworkIn, metric.NetworkOut,
			metric.Uptime, metric.LoadAverage, metric.Timestamp).Error

		if err != nil {
			logging.Errorf("Failed to store host metrics for host %d: %v", hostID, err)
		}

		// Broadcast the updated metrics to websocket clients for realtime dashboard updates
		payload := map[string]interface{}{
			"type":             "metric",
			"host_id":          hostID,
			"seq":              uint64(time.Now().UnixNano()),
			"cpu_usage":        metric.CPUUsage,
			"memory_usage":     metric.MemoryUsage,
			"memory_total":     metric.MemoryTotal,
			"memory_used":      metric.MemoryUsed,
			"memory_free":      metric.MemoryFree,
			"disk_usage":       metric.DiskUsage,
			"disk_total":       metric.DiskTotal,
			"disk_used":        metric.DiskUsed,
			"disk_read_bytes":  metric.DiskReadBytes,
			"disk_write_bytes": metric.DiskWriteBytes,
			"disk_read_speed":  metric.DiskReadSpeed,
			"disk_write_speed": metric.DiskWriteSpeed,
			"network_in":       metric.NetworkIn,
			"network_out":      metric.NetworkOut,
			"uptime":           metric.Uptime,
			"load_average":     metric.LoadAverage,
			"timestamp":        metric.Timestamp.UTC().Format(time.RFC3339),
		}
		if b, jerr := json.Marshal(payload); jerr == nil {
			if gh := ws.GetGlobalHub(); gh != nil {
				gh.RememberMessage(b)
			}
			ws.BroadcastToClients(b)
			telemetry.IncWSBroadcast(context.Background(), 1)
		}

	} // end non-placeholder else block

	// Broadcast process metrics if available
	if processesData, ok := system["processes"].(map[string]interface{}); ok {
		processPayload := map[string]interface{}{
			"type":      "processes",
			"host_id":   hostID,
			"processes": processesData,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		if pb, perr := json.Marshal(processPayload); perr == nil {
			if gh := ws.GetGlobalHub(); gh != nil {
				gh.RememberMessage(pb)
			}
			ws.BroadcastToClients(pb)
			telemetry.IncWSBroadcast(context.Background(), 1)
		}
	}

	// Broadcast service metrics if available
	if servicesData, ok := system["services"].(map[string]interface{}); ok {
		logging.Infof("[SERVICES] Received services data from agent for host=%d", hostID)
		servicePayload := map[string]interface{}{
			"type":      "services",
			"host_id":   hostID,
			"services":  servicesData,
			"seq":       uint64(time.Now().UnixNano()),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		if sb, serr := json.Marshal(servicePayload); serr == nil {
			logging.Infof("[SERVICES] Broadcasting services data for host=%d", hostID)
			if gh := ws.GetGlobalHub(); gh != nil {
				gh.RememberMessage(sb)
			}
			ws.BroadcastToClients(sb)
			telemetry.IncWSBroadcast(context.Background(), 1)
		} else {
			logging.Errorf("[SERVICES] Failed to marshal services payload: %v", serr)
		}
	}
}
