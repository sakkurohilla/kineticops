package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/golang-lru"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
	"github.com/sakkurohilla/kineticops/backend/internal/websocket"
	kafkaevents "github.com/sakkurohilla/kineticops/backend/internal/messaging/redpanda"
)

// WebSocket hub for real-time updates
var wsHub *websocket.Hub

// Kafka producer for streaming metrics
var kafkaProducer *kafkaevents.Producer

// AgentDataEvent represents data from the ELK-style agent
type AgentDataEvent struct {
	Timestamp string                 `json:"@timestamp"`
	Agent     AgentInfo              `json:"agent"`
	Host      HostInfo               `json:"host"`
	Event     EventInfo              `json:"event"`
	System    map[string]interface{} `json:"system,omitempty"`
	Log       map[string]interface{} `json:"log,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Metricset MetricsetInfo          `json:"metricset,omitempty"`
}

type AgentInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

type HostInfo struct {
	Hostname string `json:"hostname"`
}

type EventInfo struct {
	Kind     string `json:"kind"`
	Category string `json:"category"`
	Type     string `json:"type"`
	Module   string `json:"module,omitempty"`
}

type MetricsetInfo struct {
	Name string `json:"name"`
}

// SetWebSocketHub sets the WebSocket hub for broadcasting
func SetWebSocketHub(hub *websocket.Hub) {
	wsHub = hub
}

// SetKafkaProducer sets the Kafka producer for streaming
func SetKafkaProducer(producer *kafkaevents.Producer) {
	kafkaProducer = producer
}

// ReceiveAgentData handles data from ELK-style agents
func ReceiveAgentData(c *fiber.Ctx) error {
	var payload struct {
		Events []AgentDataEvent `json:"events"`
	}

	if err := c.BodyParser(&payload); err != nil {
		fmt.Printf("[DEBUG] Failed to parse agent data: %v\n", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	fmt.Printf("[DEBUG] Received %d events from agent\n", len(payload.Events))
	if len(payload.Events) > 0 {
		fmt.Printf("[DEBUG] First event: %+v\n", payload.Events[0])
	}

	if len(payload.Events) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No events provided"})
	}

	// Extract installation token from headers
	token := c.Get("X-Installation-Token")
	if token == "" {
		// Try to get from Authorization header as fallback
		auth := c.Get("Authorization")
		if auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
			token = auth[7:]
		}
	}
	fmt.Printf("[DEBUG] Agent token: %s\n", token)

	// Send events to Kafka for processing instead of direct processing
	for _, event := range payload.Events {
		if jsonData, err := json.Marshal(event); err == nil {
			if kafkaProducer != nil {
				kafkaProducer.SendMessage(jsonData)
				fmt.Printf("[DEBUG] Sent event to Kafka pipeline\n")
			} else {
				// Fallback to direct processing
				processAgentEvent(&event, token)
			}
		}
	}

	return c.JSON(fiber.Map{
		"message": "Events processed",
		"count":   len(payload.Events),
	})
}

// processAgentEvent processes a single agent event
func processAgentEvent(event *AgentDataEvent, token string) error {
	// Find or create host with token-based user association
	host, err := findOrCreateHost(event.Host.Hostname, event.Agent.Name, token)
	if err != nil {
		return err
	}

	// Update host last seen with UTC timestamp
	postgres.DB.Model(&models.Host{}).Where("id = ?", host.ID).Updates(map[string]interface{}{
		"last_seen":    time.Now().UTC(),
		"agent_status": "online",
	})
	fmt.Printf("[DEBUG] Updated host %d last_seen to %s\n", host.ID, time.Now().UTC().Format(time.RFC3339))

	// Process based on event kind
	switch event.Event.Kind {
	case "metric":
		return processMetricEvent(event, host)
	case "event":
		return processLogEvent(event, host)
	default:
		// Unknown event type, treat as heartbeat
		return nil
	}
}

// findOrCreateHost finds existing host or creates new one with proper user association
func findOrCreateHost(hostname, agentName, token string) (*models.Host, error) {
	var host models.Host
	
	// Try to find existing host by hostname first
	err := postgres.DB.Where("hostname = ?", hostname).First(&host).Error
	if err == nil {
		fmt.Printf("[DEBUG] Found existing host by hostname: %s (ID: %d)\n", hostname, host.ID)
		return &host, nil
	}
	
	// Also check by IP if hostname lookup failed
	if hostname != "" {
		err = postgres.DB.Where("ip = ?", hostname).First(&host).Error
		if err == nil {
			fmt.Printf("[DEBUG] Found existing host by IP: %s (ID: %d)\n", hostname, host.ID)
			return &host, nil
		}
	}

	// Resolve user from installation token if provided
	var tenantID uint = 1 // Default fallback
	
	if token != "" {
		if installToken, err := services.ResolveUserFromInstallationToken(token); err == nil {
			tenantID = uint(installToken.TenantID)
			// Don't mark token as used - agents need to keep using it
		}
	}

	// Create new host with proper user association
	now := time.Now().UTC()
	host = models.Host{
		Hostname:    hostname,
		IP:          "auto-discovered",
		OS:          "linux",
		AgentStatus: "online",
		Status:      "active",
		TenantID:    int64(tenantID),
		Group:       "auto-discovered",
		Description: fmt.Sprintf("Auto-discovered by %s", agentName),
		LastSeen:    now,
	}
	fmt.Printf("[DEBUG] Creating new host %s with last_seen: %s\n", hostname, now.Format(time.RFC3339))

	if err := postgres.DB.Create(&host).Error; err != nil {
		return nil, err
	}

	return &host, nil
}

// processMetricEvent processes system metrics
func processMetricEvent(event *AgentDataEvent, host *models.Host) error {
	if event.System == nil {
		return nil
	}

	timestamp, _ := time.Parse(time.RFC3339, event.Timestamp)
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	// Process different metric types
	if cpu, ok := event.System["cpu"].(map[string]interface{}); ok {
		if err := processCPUMetrics(cpu, host, timestamp); err != nil {
			return err
		}
	}

	if memory, ok := event.System["memory"].(map[string]interface{}); ok {
		if err := processMemoryMetrics(memory, host, timestamp); err != nil {
			return err
		}
	}

	if filesystem, ok := event.System["filesystem"].(map[string]interface{}); ok {
		if err := processFilesystemMetrics(filesystem, host, timestamp); err != nil {
			return err
		}
	}

	if network, ok := event.System["network"].(map[string]interface{}); ok {
		if err := processNetworkMetrics(network, host, timestamp); err != nil {
			return err
		}
	}

	if load, ok := event.System["load"].(map[string]interface{}); ok {
		if err := processLoadMetrics(load, host, timestamp); err != nil {
			return err
		}
	}

	return nil
}

// processCPUMetrics processes CPU metrics
func processCPUMetrics(cpu map[string]interface{}, host *models.Host, timestamp time.Time) error {
	fmt.Printf("[DEBUG] Processing CPU metrics: %+v\n", cpu)
	if total, ok := cpu["total"].(map[string]interface{}); ok {
		if pct, ok := total["pct"].(float64); ok {
			cpuUsage := pct * 100
			fmt.Printf("[DEBUG] CPU usage: %.2f%% for host %d\n", cpuUsage, host.ID)
			services.CollectMetric(host.ID, host.TenantID, "cpu_usage", cpuUsage, map[string]string{"type": "total"})
			// Also store in host_metrics table for dashboard
			storeHostMetric(host.ID, "cpu_usage", cpuUsage, timestamp)
		}
	}
	return nil
}

// processMemoryMetrics processes memory metrics
func processMemoryMetrics(memory map[string]interface{}, host *models.Host, timestamp time.Time) error {
	if used, ok := memory["used"].(map[string]interface{}); ok {
		if pct, ok := used["pct"].(float64); ok {
			memUsage := pct * 100
			services.CollectMetric(host.ID, host.TenantID, "memory_usage", memUsage, map[string]string{"type": "used"})
			storeHostMetric(host.ID, "memory_usage", memUsage, timestamp)
			fmt.Printf("[DEBUG] Memory usage: %.2f%% for host %d\n", memUsage, host.ID)
		}
		if bytes, ok := used["bytes"].(float64); ok {
			// Convert bytes to MB
			memUsedMB := bytes / (1024 * 1024)
			storeHostMetricField(host.ID, "memory_used", memUsedMB, timestamp)
		}
	}
	if total, ok := memory["total"].(float64); ok {
		// Convert bytes to MB
		memTotalMB := total / (1024 * 1024)
		storeHostMetricField(host.ID, "memory_total", memTotalMB, timestamp)
	}
	return nil
}

// LRU cache for enterprise scale (max 10,000 hosts)
var hostMetricsCache *lru.Cache

func init() {
	var err error
	hostMetricsCache, err = lru.New(10000)
	if err != nil {
		panic("Failed to create LRU cache")
	}
}

// storeHostMetric stores metrics in host_metrics table for dashboard
func storeHostMetric(hostID int64, metricType string, value float64, timestamp time.Time) {
	// Get current metrics for this host from cache or database
	var hostMetric *models.HostMetric
	if cached, exists := hostMetricsCache.Get(hostID); exists {
		hostMetric = cached.(*models.HostMetric)
	} else {
		// Try to get latest record from database
		var dbMetric models.HostMetric
		err := postgres.DB.Where("host_id = ?", hostID).Order("timestamp DESC").First(&dbMetric).Error
		if err != nil {
			// Create new record
			hostMetric = &models.HostMetric{
				HostID:    hostID,
				Timestamp: timestamp,
			}
		} else {
			hostMetric = &dbMetric
		}
		hostMetricsCache.Add(hostID, hostMetric)
	}

	// Update the specific metric
	switch metricType {
	case "cpu_usage":
		hostMetric.CPUUsage = value
	case "memory_usage":
		hostMetric.MemoryUsage = value
	case "disk_usage":
		hostMetric.DiskUsage = value
	case "network_in":
		hostMetric.NetworkIn = value
	case "network_out":
		hostMetric.NetworkOut = value
	}

	// Always update timestamp to latest
	hostMetric.Timestamp = timestamp
	
	// Calculate and store uptime
	var host models.Host
	if err := postgres.DB.Where("id = ?", hostID).First(&host).Error; err == nil {
		hostMetric.Uptime = calculateUptime(&host)
	}

	// Queue for bulk insert instead of individual inserts (enterprise performance)
	queueMetricForBulkInsert(hostID, metricType, value, timestamp)
	
	// Send to RedPanda/Kafka for real-time streaming
	sendToKafka(hostID, hostMetric)
	
	// Broadcast real-time update via WebSocket (fallback)
	broadcastMetricUpdate(hostID, hostMetric)
	
	fmt.Printf("[DEBUG] Queued %s=%.2f for host %d at %s\n", metricType, value, hostID, timestamp.Format(time.RFC3339))
}

// storeHostMetricField stores additional metric fields in the cache
func storeHostMetricField(hostID int64, fieldName string, value float64, timestamp time.Time) {
	if cached, exists := hostMetricsCache.Get(hostID); exists {
		hostMetric := cached.(*models.HostMetric)
		switch fieldName {
		case "memory_total":
			hostMetric.MemoryTotal = value
		case "memory_used":
			hostMetric.MemoryUsed = value
		case "disk_total":
			hostMetric.DiskTotal = value
		case "disk_used":
			hostMetric.DiskUsed = value
		}
		hostMetric.Timestamp = timestamp
	}
}

// storeHostMetricLoadAverage stores load average string
func storeHostMetricLoadAverage(hostID int64, loadAvg string, timestamp time.Time) {
	if cached, exists := hostMetricsCache.Get(hostID); exists {
		hostMetric := cached.(*models.HostMetric)
		hostMetric.LoadAverage = loadAvg
		hostMetric.Timestamp = timestamp
	}
}

// calculateUptime calculates uptime based on host creation time
func calculateUptime(host *models.Host) int64 {
	// Calculate uptime in seconds since host was first seen
	if host.CreatedAt.IsZero() {
		return 0
	}
	return int64(time.Since(host.CreatedAt).Seconds())
}

// processFilesystemMetrics processes filesystem metrics
func processFilesystemMetrics(filesystem map[string]interface{}, host *models.Host, timestamp time.Time) error {
	// Get mount point to filter out non-main filesystems
	mountPoint, _ := filesystem["mount_point"].(string)
	deviceName, _ := filesystem["device_name"].(string)
	
	// Only process root filesystem
	if mountPoint != "/" {
		fmt.Printf("[DEBUG] Skipping filesystem %s mounted at %s\n", deviceName, mountPoint)
		return nil
	}
	
	if used, ok := filesystem["used"].(map[string]interface{}); ok {
		if pct, ok := used["pct"].(float64); ok {
			diskUsage := pct * 100
			services.CollectMetric(host.ID, host.TenantID, "disk_usage", diskUsage, map[string]string{"type": "used", "mount": mountPoint})
			storeHostMetric(host.ID, "disk_usage", diskUsage, timestamp)
			fmt.Printf("[DEBUG] Disk usage: %.2f%% for host %d (mount: %s)\n", diskUsage, host.ID, mountPoint)
		}
		if bytes, ok := used["bytes"].(float64); ok {
			// Convert bytes to GB
			diskUsedGB := bytes / (1024 * 1024 * 1024)
			storeHostMetricField(host.ID, "disk_used", diskUsedGB, timestamp)
		}
	}
	if total, ok := filesystem["total"].(float64); ok {
		// Convert bytes to GB
		diskTotalGB := total / (1024 * 1024 * 1024)
		storeHostMetricField(host.ID, "disk_total", diskTotalGB, timestamp)
	}
	return nil
}

// isMainPartition checks if device is a main storage partition
func isMainPartition(deviceName string) bool {
	// Skip CD-ROM, DVD drives
	if deviceName == "/dev/sr0" || deviceName == "/dev/cdrom" {
		return false
	}
	// Accept main disk partitions
	if len(deviceName) >= 8 && (deviceName[:8] == "/dev/sda" || deviceName[:8] == "/dev/nvme") {
		return true
	}
	return false
}

// processNetworkMetrics processes network metrics
func processNetworkMetrics(network map[string]interface{}, host *models.Host, timestamp time.Time) error {
	name, _ := network["name"].(string)
	
	if in, ok := network["in"].(map[string]interface{}); ok {
		if bytes, ok := in["bytes"].(float64); ok {
			services.CollectMetric(host.ID, host.TenantID, "network_bytes", bytes, map[string]string{"direction": "in", "interface": name})
		}
	}

	if out, ok := network["out"].(map[string]interface{}); ok {
		if bytes, ok := out["bytes"].(float64); ok {
			services.CollectMetric(host.ID, host.TenantID, "network_bytes", bytes, map[string]string{"direction": "out", "interface": name})
		}
	}

	return nil
}

// processLoadMetrics processes load average metrics
func processLoadMetrics(load map[string]interface{}, host *models.Host, timestamp time.Time) error {
	if load1, ok := load["1"].(float64); ok {
		services.CollectMetric(host.ID, host.TenantID, "load_average", load1, map[string]string{"period": "1m"})
	}
	if load5, ok := load["5"].(float64); ok {
		services.CollectMetric(host.ID, host.TenantID, "load_average", load5, map[string]string{"period": "5m"})
	}
	if load15, ok := load["15"].(float64); ok {
		services.CollectMetric(host.ID, host.TenantID, "load_average", load15, map[string]string{"period": "15m"})
	}

	// Store load average string for dashboard
	if load1, ok := load["1"].(float64); ok {
		if load5, ok := load["5"].(float64); ok {
			if load15, ok := load["15"].(float64); ok {
				loadStr := fmt.Sprintf("%.2f %.2f %.2f", load1, load5, load15)
				storeHostMetricLoadAverage(host.ID, loadStr, timestamp)
			}
		}
	}

	return nil
}

// processLogEvent processes log events
func processLogEvent(event *AgentDataEvent, host *models.Host) error {
	if event.Message == "" {
		return nil
	}

	timestamp, _ := time.Parse(time.RFC3339, event.Timestamp)
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	// Extract log level
	level := "info"
	if event.Log != nil {
		if logLevel, ok := event.Log["level"].(string); ok {
			level = logLevel
		}
	}

	// Create log entry
	logEntry := &models.Log{
		TenantID:  host.TenantID,
		HostID:    host.ID,
		Timestamp: timestamp,
		Level:     level,
		Message:   event.Message,
		Meta:      make(map[string]string),
	}

	// Add metadata
	if event.Log != nil {
		if file, ok := event.Log["file"].(map[string]interface{}); ok {
			if path, ok := file["path"].(string); ok {
				logEntry.Meta["file_path"] = path
			}
		}
	}

	// Store log (this would typically go to MongoDB)
	// For now, we'll just log it
	return nil
}

// broadcastMetricUpdate sends real-time metric updates via WebSocket
func broadcastMetricUpdate(hostID int64, metric *models.HostMetric) {
	// Create WebSocket payload
	payload := map[string]interface{}{
		"type":          "metric_update",
		"host_id":       hostID,
		"cpu_usage":     metric.CPUUsage,
		"memory_usage":  metric.MemoryUsage,
		"memory_total":  metric.MemoryTotal,
		"memory_used":   metric.MemoryUsed,
		"disk_usage":    metric.DiskUsage,
		"disk_total":    metric.DiskTotal,
		"disk_used":     metric.DiskUsed,
		"network_in":    metric.NetworkIn,
		"network_out":   metric.NetworkOut,
		"uptime":        metric.Uptime,
		"load_average":  metric.LoadAverage,
		"timestamp":     metric.Timestamp.Format(time.RFC3339),
	}
	
	// Broadcast to WebSocket hub (if available)
	if wsHub != nil {
		if jsonData, err := json.Marshal(payload); err == nil {
			wsHub.Broadcast(jsonData)
			fmt.Printf("[DEBUG] Broadcasted metric update for host %d via WebSocket\n", hostID)
		}
	}
}

// sendToKafka sends metrics to RedPanda/Kafka for streaming
func sendToKafka(hostID int64, metric *models.HostMetric) {
	payload := map[string]interface{}{
		"type":          "metric_update",
		"host_id":       hostID,
		"cpu_usage":     metric.CPUUsage,
		"memory_usage":  metric.MemoryUsage,
		"memory_total":  metric.MemoryTotal,
		"memory_used":   metric.MemoryUsed,
		"disk_usage":    metric.DiskUsage,
		"disk_total":    metric.DiskTotal,
		"disk_used":     metric.DiskUsed,
		"network_in":    metric.NetworkIn,
		"network_out":   metric.NetworkOut,
		"uptime":        metric.Uptime,
		"load_average":  metric.LoadAverage,
		"timestamp":     metric.Timestamp.Format(time.RFC3339),
		"seq":           time.Now().UnixNano(), // Sequence for ordering
	}
	
	if jsonData, err := json.Marshal(payload); err == nil {
		// Send to Kafka producer (if available)
		if kafkaProducer != nil {
			kafkaProducer.SendMessage(jsonData)
			fmt.Printf("[DEBUG] Sent metric to Kafka for host %d\n", hostID)
		}
	}
}

// ProcessKafkaEvent processes agent events received from Kafka
func ProcessKafkaEvent(event *AgentDataEvent) {
	// Process the event (same logic as direct HTTP)
	if err := processAgentEvent(event, ""); err != nil {
		fmt.Printf("[DEBUG] Failed to process Kafka event: %v\n", err)
	}
}

// Bulk insert system for enterprise performance
var (
	metricQueue = make(chan models.TimeseriesMetric, 10000)
	bulkSize    = 100
)

func init() {
	// Start bulk insert worker
	go bulkInsertWorker()
}

// queueMetricForBulkInsert adds metric to bulk insert queue
func queueMetricForBulkInsert(hostID int64, metricName string, value float64, timestamp time.Time) {
	metric := models.TimeseriesMetric{
		Time:       timestamp,
		HostID:     hostID,
		MetricName: metricName,
		Value:      value,
		TenantID:   1, // TODO: Get from context
	}
	
	select {
	case metricQueue <- metric:
	default:
		// Queue full, drop metric (enterprise resilience)
		fmt.Printf("[WARN] Metric queue full, dropping metric for host %d\n", hostID)
	}
}

// bulkInsertWorker processes metrics in batches
func bulkInsertWorker() {
	batch := make([]models.TimeseriesMetric, 0, bulkSize)
	ticker := time.NewTicker(5 * time.Second) // Flush every 5 seconds
	
	for {
		select {
		case metric := <-metricQueue:
			batch = append(batch, metric)
			if len(batch) >= bulkSize {
				flushBatch(batch)
				batch = batch[:0] // Reset slice
			}
		case <-ticker.C:
			if len(batch) > 0 {
				flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch performs bulk insert
func flushBatch(batch []models.TimeseriesMetric) {
	if len(batch) == 0 {
		return
	}
	
	if err := postgres.DB.CreateInBatches(batch, len(batch)).Error; err != nil {
		fmt.Printf("[ERROR] Bulk insert failed: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] Bulk inserted %d metrics\n", len(batch))
	}
}