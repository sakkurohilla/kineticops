package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/messaging/redpanda"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	redisrepo "github.com/sakkurohilla/kineticops/backend/internal/repository/redis"
	"github.com/sakkurohilla/kineticops/backend/internal/telemetry"
)

// Optional schema validation for metric types
func ValidateMetricSchema(name string, val float64, labels string) error {
	if val < 0 && (name == "cpu_usage" || name == "ram_usage") {
		return fmt.Errorf("invalid metric value")
	}
	return nil
}

func CollectMetric(hostID, tenantID int64, name string, value float64, labels map[string]string) error {
	lb, _ := json.Marshal(labels)
	metric := &models.Metric{
		HostID:    hostID,
		TenantID:  tenantID,
		Name:      name,
		Value:     value,
		Timestamp: time.Now(),
		Labels:    string(lb),
	}
	// Try to enqueue metric into the batcher for bulk insert. If the batcher
	// channel is full or not running, fall back to direct insert for reliability.
	if metricBatcher != nil {
		select {
		case metricBatcher.metricsChan <- metric:
			return nil
		default:
			// Channel full - fall back to direct save
		}
	}

	start := time.Now()
	err := postgres.SaveMetric(postgres.DB, metric)
	if err != nil {
		telemetry.IncCollectionError(context.Background(), 1)
		return err
	}
	telemetry.IncCollectionSuccess(context.Background(), 1)
	// Optionally record ingestion latency in logs for lightweight monitoring
	fmt.Printf("[METRICS] direct insert latency=%s host=%d metric=%s\n", time.Since(start), hostID, name)
	return nil
}

// metricBatcherSingleton provides a background goroutine that batches metrics.
var metricBatcher *metricBatcherSingleton

type metricBatcherSingleton struct {
	metricsChan   chan *models.Metric
	batchSize     int
	flushInterval time.Duration
	stopCh        chan struct{}
	lastFlushMu   sync.Mutex
	lastFlushAt   time.Time
}

// StartMetricBatcher initializes and runs the background metric batcher. Call once at startup.
func StartMetricBatcher(batchSize int, flushInterval time.Duration) {
	if metricBatcher != nil {
		return
	}
	mb := &metricBatcherSingleton{
		metricsChan:   make(chan *models.Metric, batchSize*4),
		batchSize:     batchSize,
		flushInterval: flushInterval,
		stopCh:        make(chan struct{}),
	}
	metricBatcher = mb

	go mb.run()
	fmt.Printf("[METRICS] started batcher batchSize=%d flushInterval=%s\n", batchSize, flushInterval)
}

// StopMetricBatcher stops the batcher gracefully.
func StopMetricBatcher() {
	if metricBatcher == nil {
		return
	}
	close(metricBatcher.stopCh)
	metricBatcher = nil
}

// GetMetricBatcherQueueLength returns the number of metrics currently buffered in the batcher.
func GetMetricBatcherQueueLength() int {
	if metricBatcher == nil {
		return 0
	}
	return len(metricBatcher.metricsChan)
}

// GetMetricBatcherStatus returns queue length and last flush timestamp (UTC) for health checks.
func GetMetricBatcherStatus() (int, string) {
	if metricBatcher == nil {
		return 0, ""
	}
	metricBatcher.lastFlushMu.Lock()
	lf := metricBatcher.lastFlushAt
	metricBatcher.lastFlushMu.Unlock()
	var lastFlushStr string
	if !lf.IsZero() {
		lastFlushStr = lf.Format(time.RFC3339)
	}
	return len(metricBatcher.metricsChan), lastFlushStr
}

func (m *metricBatcherSingleton) run() {
	buffer := make([]*models.Metric, 0, m.batchSize)
	flushTimer := time.NewTimer(m.flushInterval)
	defer flushTimer.Stop()

	flush := func() {
		if len(buffer) == 0 {
			return
		}
		start := time.Now()
		err := postgres.SaveMetricsBatch(postgres.DB, buffer)
		if err != nil {
			telemetry.IncCollectionError(context.Background(), int64(len(buffer)))
			fmt.Printf("[METRICS] batch save error: %v\n", err)
			// Publish the failed batch to Redpanda for durable buffering
			if b, jerr := json.Marshal(buffer); jerr == nil {
				if pubErr := redpanda.PublishEvent(b); pubErr != nil {
					fmt.Printf("[METRICS] failed to publish failed-batch to redpanda: %v\n", pubErr)
				} else {
					fmt.Printf("[METRICS] published failed batch of %d metrics to redpanda\n", len(buffer))
				}
			}
		} else {
			telemetry.IncCollectionSuccess(context.Background(), int64(len(buffer)))
			fmt.Printf("[METRICS] flushed %d metrics in %s\n", len(buffer), time.Since(start))
			// record last successful flush
			m.lastFlushMu.Lock()
			m.lastFlushAt = time.Now().UTC()
			m.lastFlushMu.Unlock()
			// Trigger alert evaluation asynchronously per metric
			go func(batch []*models.Metric) {
				for _, m := range batch {
					// best-effort alert checking; errors are ignored here
					_ = CheckAndTriggerAlerts(m.TenantID, m.Name, m.HostID, m.Value)
				}
			}(append([]*models.Metric(nil), buffer...))
		}
		buffer = buffer[:0]
	}

	for {
		select {
		case <-m.stopCh:
			flush()
			return
		case <-flushTimer.C:
			flush()
			flushTimer.Reset(m.flushInterval)
		case mt := <-m.metricsChan:
			buffer = append(buffer, mt)
			// update Prometheus gauge for ingestion queue length
			telemetry.SetMetricIngestionQueueLength(len(m.metricsChan))
			if len(buffer) >= m.batchSize {
				flush()
				flushTimer.Reset(m.flushInterval)
			}
		}
	}
}

func ListMetrics(tenantID, hostID int64, name string, start, end time.Time, limit int) ([]models.Metric, error) {
	cacheKey := fmt.Sprintf("metrics:%d:%d:%s:%s:%s:%d", tenantID, hostID, name, start.Format(time.RFC3339), end.Format(time.RFC3339), limit)
	cached, err := redisrepo.GetMetricsCache(cacheKey)
	if err == nil && cached != nil {
		return cached, nil // Return cached result
	}

	data, err := postgres.ListMetrics(postgres.DB, tenantID, hostID, name, start, end, limit)
	if err == nil && data != nil {
		_ = redisrepo.SetMetricsCache(cacheKey, data)
	}
	return data, err
}

func LatestMetric(hostID int64, name string) (*models.Metric, error) {
	return postgres.LatestMetric(postgres.DB, hostID, name)
}

func EnforceRetentionPolicy(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return postgres.DeleteOldMetrics(postgres.DB, cutoff)
}
