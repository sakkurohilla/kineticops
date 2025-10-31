package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	kafkaevents "github.com/sakkurohilla/kineticops/backend/internal/messaging/redpanda"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"github.com/sakkurohilla/kineticops/backend/internal/telemetry"
	ws "github.com/sakkurohilla/kineticops/backend/internal/websocket"
)

// HostMetric represents collected metrics from a host
type HostMetric struct {
	HostID      int64
	CPUUsage    float64
	MemoryUsage float64
	MemoryTotal float64
	MemoryUsed  float64
	DiskUsage   float64
	DiskTotal   float64
	DiskUsed    float64
	NetworkIn   float64
	NetworkOut  float64
	Uptime      int64
	LoadAverage string
}

// HostMetricDB represents the database model for host_metrics table
type HostMetricDB struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	HostID      int64     `gorm:"column:host_id;not null"`
	CPUUsage    float64   `gorm:"column:cpu_usage;type:decimal(5,2);default:0"`
	MemoryUsage float64   `gorm:"column:memory_usage;type:decimal(5,2);default:0"`
	MemoryTotal float64   `gorm:"column:memory_total;type:decimal(10,2);default:0"`
	MemoryUsed  float64   `gorm:"column:memory_used;type:decimal(10,2);default:0"`
	DiskUsage   float64   `gorm:"column:disk_usage;type:decimal(5,2);default:0"`
	DiskTotal   float64   `gorm:"column:disk_total;type:decimal(10,2);default:0"`
	DiskUsed    float64   `gorm:"column:disk_used;type:decimal(10,2);default:0"`
	NetworkIn   float64   `gorm:"column:network_in;type:decimal(10,2);default:0"`
	NetworkOut  float64   `gorm:"column:network_out;type:decimal(10,2);default:0"`
	Uptime      int64     `gorm:"column:uptime;default:0"`
	LoadAverage string    `gorm:"column:load_average;size:64;default:''"`
	Timestamp   time.Time `gorm:"column:timestamp;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for GORM
func (HostMetricDB) TableName() string {
	return "host_metrics"
}

// CollectHostMetrics collects all metrics from a single host
func CollectHostMetrics(host *models.Host) (*HostMetric, error) {
	// Create SSH client with password or key
	sshClient, err := NewSSHClientWithKey(host.IP, int(host.SSHPort), host.SSHUser, host.SSHPassword, host.SSHKey)
	if err != nil {
		return nil, fmt.Errorf("SSH connection failed: %w", err)
	}
	defer sshClient.Close()

	metric := &HostMetric{
		HostID: host.ID,
	}

	// Collect metrics in parallel to reduce overall collection time and avoid one
	// slow command blocking others. Each collector is tolerant and returns zero
	// values on parse/command errors.
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(6)

	go func() {
		defer wg.Done()
		if v, err := sshClient.CollectCPUUsage(); err != nil {
			log.Printf("[WARN] Failed to collect CPU for host %d: %v", host.ID, err)
		} else {
			mu.Lock()
			metric.CPUUsage = v
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if used, total, perc, err := sshClient.CollectMemoryUsage(); err != nil {
			log.Printf("[WARN] Failed to collect memory for host %d: %v", host.ID, err)
		} else {
			mu.Lock()
			metric.MemoryUsed = used
			metric.MemoryTotal = total
			metric.MemoryUsage = perc
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if used, total, perc, err := sshClient.CollectDiskUsage(); err != nil {
			log.Printf("[WARN] Failed to collect disk for host %d: %v", host.ID, err)
		} else {
			mu.Lock()
			metric.DiskUsed = used
			metric.DiskTotal = total
			metric.DiskUsage = perc
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if in, out, err := sshClient.CollectNetworkStats(); err != nil {
			log.Printf("[WARN] Failed to collect network for host %d: %v", host.ID, err)
		} else {
			mu.Lock()
			metric.NetworkIn = in
			metric.NetworkOut = out
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if up, err := sshClient.CollectUptime(); err != nil {
			log.Printf("[WARN] Failed to collect uptime for host %d: %v", host.ID, err)
		} else {
			mu.Lock()
			metric.Uptime = up
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if l, err := sshClient.CollectLoadAverage(); err != nil {
			log.Printf("[WARN] Failed to collect load average for host %d: %v", host.ID, err)
		} else {
			mu.Lock()
			metric.LoadAverage = l
			mu.Unlock()
		}
	}()

	wg.Wait()

	return metric, nil
}

// SaveHostMetrics saves metrics to database using GORM
func SaveHostMetrics(metric *HostMetric) error {
	dbm := &models.HostMetric{
		HostID:      metric.HostID,
		CPUUsage:    metric.CPUUsage,
		MemoryUsage: metric.MemoryUsage,
		MemoryTotal: metric.MemoryTotal,
		MemoryUsed:  metric.MemoryUsed,
		DiskUsage:   metric.DiskUsage,
		DiskTotal:   metric.DiskTotal,
		DiskUsed:    metric.DiskUsed,
		NetworkIn:   metric.NetworkIn,
		NetworkOut:  metric.NetworkOut,
		Uptime:      metric.Uptime,
		LoadAverage: metric.LoadAverage,
		Timestamp:   time.Now(),
	}

	if err := postgres.SaveHostMetric(postgres.DB, dbm); err != nil {
		// instrumentation
		telemetry.IncCollectionError(context.Background(), 1)
		return err
	}
	telemetry.IncCollectionSuccess(context.Background(), 1)

	// Mirror key metrics into the generic metrics table so the /api/v1/metrics
	// endpoints (which read from models.Metric) return persistent data and
	// dashboard/metrics pages don't get cleared when websocket-only events arrive.
	// Attempt to resolve tenant id from host record; if not available, tenantID=0.
	var tenantID int64 = 0
	if h, err := postgres.GetHost(postgres.DB, dbm.HostID); err == nil {
		tenantID = h.TenantID
	}

	metricsToSave := []models.Metric{
		{HostID: dbm.HostID, TenantID: tenantID, Name: "cpu_usage", Value: dbm.CPUUsage, Timestamp: dbm.Timestamp},
		{HostID: dbm.HostID, TenantID: tenantID, Name: "memory_usage", Value: dbm.MemoryUsage, Timestamp: dbm.Timestamp},
		{HostID: dbm.HostID, TenantID: tenantID, Name: "disk_usage", Value: dbm.DiskUsage, Timestamp: dbm.Timestamp},
		{HostID: dbm.HostID, TenantID: tenantID, Name: "network", Value: dbm.NetworkIn + dbm.NetworkOut, Timestamp: dbm.Timestamp},
	}
	for _, mm := range metricsToSave {
		if err := postgres.SaveMetric(postgres.DB, &mm); err != nil {
			// don't fail the whole flow for metric mirroring; just log and continue
			log.Printf("[WARN] failed to mirror metric %s for host %d: %v", mm.Name, dbm.HostID, err)
		}
	}

	// Publish metric event to Kafka/Redpanda for real-time websocket broadcast
	// add monotonic sequence id for ordering across websocket consumers
	seq := telemetry.NextSeq()
	payload := map[string]interface{}{
		"host_id":      dbm.HostID,
		"cpu_usage":    dbm.CPUUsage,
		"memory_usage": dbm.MemoryUsage,
		"disk_usage":   dbm.DiskUsage,
		"network_in":   dbm.NetworkIn,
		"network_out":  dbm.NetworkOut,
		"seq":          seq,
		"uptime":       dbm.Uptime,
		"load_average": dbm.LoadAverage,
		"timestamp":    dbm.Timestamp.Format(time.RFC3339),
	}
	if b, err := json.Marshal(payload); err == nil {
		// remember the last message for new clients (best-effort)
		if gh := ws.GetGlobalHub(); gh != nil {
			gh.RememberMessage(b)
		}
		if err := kafkaevents.PublishEvent(b); err != nil {
			log.Printf("[WARN] Failed to publish metric event: %v", err)
			// instrumentation
			telemetry.IncWSSendErrors(context.Background(), 1)
			// Fallback: broadcast directly to connected websocket clients so UI remains realtime
			ws.BroadcastToClients(b)
			telemetry.IncWSBroadcast(context.Background(), 1)
		} else {
			telemetry.IncKafkaPublish(context.Background(), 1)
		}
	} else {
		log.Printf("[WARN] Failed to marshal metric payload: %v", err)
	}

	return nil
}

// UpdateHostStatus updates host online/offline status
func UpdateHostStatus(hostID int64, status string) error {
	return postgres.UpdateHost(postgres.DB, hostID, map[string]interface{}{
		"agent_status": status,
		"last_seen":    time.Now(),
	})
}
