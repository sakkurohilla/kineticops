package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// MetricAggregationService handles metric aggregation
type MetricAggregationService struct{}

func NewMetricAggregationService() *MetricAggregationService {
	return &MetricAggregationService{}
}

// AggregateMetrics performs aggregation for a given interval
func (s *MetricAggregationService) AggregateMetrics(interval string, duration time.Duration) error {
	cutoff := time.Now().Add(-duration)

	var metrics []struct {
		HostID     int64
		MetricName string
		Avg        float64
		Min        float64
		Max        float64
		Sum        float64
		Count      int64
	}

	// Aggregate from metrics table
	err := postgres.DB.Raw(`
		SELECT 
			host_id,
			name as metric_name,
			AVG(value) as avg,
			MIN(value) as min,
			MAX(value) as max,
			SUM(value) as sum,
			COUNT(*) as count
		FROM metrics
		WHERE timestamp > ?
		GROUP BY host_id, name
	`, cutoff).Scan(&metrics).Error

	if err != nil {
		return fmt.Errorf("failed to aggregate metrics: %w", err)
	}

	// Calculate percentiles (simplified - for production use proper percentile calculation)
	for _, m := range metrics {
		agg := models.MetricAggregation{
			HostID:       m.HostID,
			MetricName:   m.MetricName,
			Interval:     interval,
			IntervalTime: time.Now().Truncate(duration),
			Avg:          m.Avg,
			Min:          m.Min,
			Max:          m.Max,
			Sum:          m.Sum,
			Count:        m.Count,
		}

		// Calculate percentiles
		var values []float64
		postgres.DB.Raw(`
			SELECT value FROM metrics 
			WHERE host_id = ? AND name = ? AND timestamp > ?
			ORDER BY value
		`, m.HostID, m.MetricName, cutoff).Pluck("value", &values)

		if len(values) > 0 {
			agg.P50 = percentile(values, 0.50)
			agg.P95 = percentile(values, 0.95)
			agg.P99 = percentile(values, 0.99)
		}

		// Insert or update aggregation
		postgres.DB.Create(&agg)
	}

	return nil
}

// StartAggregationWorker runs periodic aggregations
func (s *MetricAggregationService) StartAggregationWorker(ctx context.Context) {
	// 1-minute aggregation
	go s.runAggregation(ctx, "1m", 1*time.Minute, 1*time.Minute)

	// 5-minute aggregation
	go s.runAggregation(ctx, "5m", 5*time.Minute, 5*time.Minute)

	// 1-hour aggregation
	go s.runAggregation(ctx, "1h", 1*time.Hour, 1*time.Hour)

	// 24-hour aggregation
	go s.runAggregation(ctx, "24h", 24*time.Hour, 24*time.Hour)
}

func (s *MetricAggregationService) runAggregation(ctx context.Context, interval string, duration, tickDuration time.Duration) {
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.AggregateMetrics(interval, duration); err != nil {
				logging.Errorf("Aggregation failed for interval %s: %v", interval, err)
			} else {
				logging.Infof("Metrics aggregated for interval: %s", interval)
			}
		}
	}
}

// GetAggregatedMetrics retrieves aggregated metrics
func (s *MetricAggregationService) GetAggregatedMetrics(hostID int64, metricName, interval string, start, end time.Time) ([]models.MetricAggregation, error) {
	var aggregations []models.MetricAggregation

	query := postgres.DB.Where("host_id = ? AND metric_name = ? AND interval = ?", hostID, metricName, interval)

	if !start.IsZero() {
		query = query.Where("interval_time >= ?", start)
	}
	if !end.IsZero() {
		query = query.Where("interval_time <= ?", end)
	}

	err := query.Order("interval_time ASC").Find(&aggregations).Error
	return aggregations, err
}

// percentile calculates the nth percentile of a sorted slice
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	index := p * float64(len(values)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return values[lower]
	}

	fraction := index - float64(lower)
	return values[lower]*(1-fraction) + values[upper]*fraction
}

// Global aggregation service
var MetricAggregationSvc = NewMetricAggregationService()
