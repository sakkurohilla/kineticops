package handlers

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

type AgentEvent struct {
	Timestamp string                 `json:"@timestamp"`
	Agent     map[string]interface{} `json:"agent"`
	Host      map[string]interface{} `json:"host"`
	Event     map[string]interface{} `json:"event"`
	System    map[string]interface{} `json:"system"`
}

func ReceiveAgentData(c *fiber.Ctx) error {
	var payload struct {
		Events []AgentEvent `json:"events"`
	}

	if err := c.BodyParser(&payload); err != nil {
		// Log parse error for diagnostics (do not echo payload contents)
		fmt.Printf("[ERROR] agent data parse error: %v\n", err)
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
	allIPs, _ := hostData["ips"].([]interface{})
	os, _ := hostData["os"].(string)
	platform, _ := hostData["platform"].(string)
	platformFamily, _ := hostData["platform_family"].(string)
	platformVersion, _ := hostData["platform_version"].(string)
	arch, _ := hostData["arch"].(string)
	kernelVersion, _ := hostData["kernel_version"].(string)
	virtualization, _ := hostData["virtualization"].(string)

	// Convert IPs to string slice
	var ipStrings []string
	for _, ip := range allIPs {
		if ipStr, ok := ip.(string); ok {
			ipStrings = append(ipStrings, ipStr)
		}
	}

	host := findOrCreateHost(hostname, primaryIP, os, platform, platformFamily,
		platformVersion, arch, kernelVersion, virtualization, ipStrings, tenantID)
	if host == nil {
		return false
	}

	// Update host last seen and status
	now := time.Now().UTC()
	postgres.DB.Model(&models.Host{}).Where("id = ?", host.ID).Updates(map[string]interface{}{
		"last_seen":    now,
		"agent_status": "online",
	})

	// Process system metrics with validation
	if event.System != nil {
		processSystemMetrics(host.ID, host.TenantID, event.System, now)
		return true
	}

	return false
}

func findOrCreateHost(hostname, primaryIP, os, platform, platformFamily,
	platformVersion, arch, kernelVersion, virtualization string,
	allIPs []string, tenantID int64) *models.Host {

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
			postgres.DB.Model(&host).Updates(updates)
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
	now := time.Now().UTC()
	host = models.Host{
		Hostname:    hostname,
		IP:          primaryIP,
		OS:          os,
		AgentStatus: "online",
		Status:      "active",
		TenantID:    tenantID,
		Group:       "auto-discovered",
		Description: fmt.Sprintf("Auto-discovered %s host on %s", platform, platformFamily),
		LastSeen:    now,
		RegToken:    fmt.Sprintf("auto-%d-%s-%d", tenantID, hostname, now.Unix()),
	}

	if err := postgres.DB.Create(&host).Error; err != nil {
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
	metric := &models.HostMetric{HostID: hostID, Timestamp: timestamp}

	// CPU metrics with validation
	if cpu, ok := system["cpu"].(map[string]interface{}); ok {
		if total, ok := cpu["total"].(map[string]interface{}); ok {
			if pct, ok := total["pct"].(float64); ok && pct >= 0 && pct <= 1 {
				metric.CPUUsage = pct * 100
				services.CollectMetric(hostID, tenantID, "cpu_usage", metric.CPUUsage, nil)
			}
		}
	}

	// Memory metrics with validation
	if memory, ok := system["memory"].(map[string]interface{}); ok {
		// Prefer bytes-based calculation when possible. Some agents report `available.bytes` instead
		// of `used.bytes`. To be accurate, compute used = total - available when available is present.
		var totalBytes float64 = 0
		if total, ok := memory["total"].(float64); ok && total > 0 {
			totalBytes = total
			metric.MemoryTotal = total / (1024 * 1024) // MB
		}

		var usedBytes float64 = 0
		// If available bytes provided, compute used = total - available
		if avail, ok := memory["available"].(map[string]interface{}); ok {
			if ab, ok := avail["bytes"].(float64); ok && ab >= 0 && totalBytes > 0 {
				usedBytes = totalBytes - ab
			}
		}

		// If agent reports free bytes (older agents), use total - free to compute used
		if usedBytes == 0 {
			if freeVal, ok := memory["free"].(float64); ok && freeVal >= 0 && totalBytes > 0 {
				usedBytes = totalBytes - freeVal
			}
		}

		// Fallback to used.bytes if available
		if used, ok := memory["used"].(map[string]interface{}); ok {
			if bytes, ok := used["bytes"].(float64); ok && bytes >= 0 {
				// if we haven't computed usedBytes from available, use used.bytes
				if usedBytes == 0 {
					usedBytes = bytes
				}
			}
			if pct, ok := used["pct"].(float64); ok && pct >= 0 && pct <= 1 {
				metric.MemoryUsage = pct * 100
				services.CollectMetric(hostID, tenantID, "memory_usage", metric.MemoryUsage, nil)
			}
		}

		// If we have totalBytes and computed usedBytes, normalize and store
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
			services.CollectMetric(hostID, tenantID, "memory_usage", metric.MemoryUsage, nil)
		}
	}

	// Disk metrics with validation - ONLY root filesystem
	if fs, ok := system["filesystem"].(map[string]interface{}); ok {
		// STRICT: Only process root filesystem
		mountPoint, _ := fs["mount_point"].(string)
		deviceName, _ := fs["device_name"].(string)

		// Skip if not root filesystem
		if mountPoint != "/" {
			fmt.Printf("[DEBUG] Skipping filesystem %s mounted at %s\n", deviceName, mountPoint)
			return
		}

		fmt.Printf("[DEBUG] Processing root filesystem %s mounted at %s\n", deviceName, mountPoint)

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
				services.CollectMetric(hostID, tenantID, "disk_usage", metric.DiskUsage, nil)
				// If agent also reported pct, normalize and compare
				if agentPctVal >= 0 {
					var agentPctPercent float64
					if agentPctVal <= 1 {
						agentPctPercent = agentPctVal * 100
					} else {
						agentPctPercent = agentPctVal
					}
					if math.Abs(agentPctPercent-serverPct) > 15.0 {
						fmt.Printf("[WARN] Disk pct mismatch host=%d device=%s agent_pct=%.2f server_pct=%.2f\n", hostID, deviceName, agentPctPercent, serverPct)
					}
				}
				fmt.Printf("[DEBUG] Root disk usage (bytes) for %s mounted at %s: %.2f%%\n", deviceName, mountPoint, metric.DiskUsage)
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
					services.CollectMetric(hostID, tenantID, "disk_usage", metric.DiskUsage, nil)
					fmt.Printf("[DEBUG] Root disk usage (agent pct) for %s mounted at %s: %.2f%%\n", deviceName, mountPoint, metric.DiskUsage)
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
		if in, ok := net["in"].(map[string]interface{}); ok {
			if bytes, ok := in["bytes"].(float64); ok && bytes >= 0 {
				metric.NetworkIn = bytes / (1024 * 1024) // Convert to MB
			}
		}
		if out, ok := net["out"].(map[string]interface{}); ok {
			if bytes, ok := out["bytes"].(float64); ok && bytes >= 0 {
				metric.NetworkOut = bytes / (1024 * 1024) // Convert to MB
			}
		}
	}

	// Load average with validation
	if load, ok := system["load"].(map[string]interface{}); ok {
		if load1, ok := load["1"].(float64); ok && load1 >= 0 {
			if load5, ok := load["5"].(float64); ok && load5 >= 0 {
				if load15, ok := load["15"].(float64); ok && load15 >= 0 {
					metric.LoadAverage = fmt.Sprintf("%.2f %.2f %.2f", load1, load5, load15)

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
			if v == "unavailable" {
				metric.Uptime = 0
			}
		default:
			metric.Uptime = 0
		}
	} else {
		metric.Uptime = 0
	}

	// Store in host_metrics table with error handling
	err := postgres.DB.Exec(`
		INSERT INTO host_metrics (host_id, cpu_usage, memory_usage, memory_total, memory_used, 
			disk_usage, disk_total, disk_used, network_in, network_out, uptime, load_average, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (host_id) DO UPDATE SET
			cpu_usage = EXCLUDED.cpu_usage,
			memory_usage = EXCLUDED.memory_usage,
			memory_total = EXCLUDED.memory_total,
			memory_used = EXCLUDED.memory_used,
			disk_usage = EXCLUDED.disk_usage,
			disk_total = EXCLUDED.disk_total,
			disk_used = EXCLUDED.disk_used,
			network_in = EXCLUDED.network_in,
			network_out = EXCLUDED.network_out,
			uptime = EXCLUDED.uptime,
			load_average = EXCLUDED.load_average,
			timestamp = EXCLUDED.timestamp
	`, hostID, metric.CPUUsage, metric.MemoryUsage, metric.MemoryTotal, metric.MemoryUsed,
		metric.DiskUsage, metric.DiskTotal, metric.DiskUsed, metric.NetworkIn, metric.NetworkOut,
		metric.Uptime, metric.LoadAverage, metric.Timestamp).Error

	if err != nil {
		fmt.Printf("[ERROR] Failed to store host metrics for host %d: %v\n", hostID, err)
	}
}
