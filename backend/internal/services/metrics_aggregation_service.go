package services

import (
	"fmt"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// MetricsAggregationService handles time-series data aggregation
type MetricsAggregationService struct{}

// NewMetricsAggregationService creates a new aggregation service
func NewMetricsAggregationService() *MetricsAggregationService {
	return &MetricsAggregationService{}
}

// AggregatedMetric represents aggregated metric data
type AggregatedMetric struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	HostID    int64     `json:"host_id"`
	Name      string    `json:"name"`
}

// GetAggregatedMetrics returns properly aggregated metrics for time ranges
func (s *MetricsAggregationService) GetAggregatedMetrics(hostID int64, metricName string, timeRange string) ([]AggregatedMetric, error) {
	var startTime time.Time
	var interval string
	now := time.Now().UTC()

	// Determine time range and aggregation interval
	switch timeRange {
	case "1h":
		startTime = now.Add(-1 * time.Hour)
		interval = "1 minute"
	case "6h":
		startTime = now.Add(-6 * time.Hour)
		interval = "5 minutes"
	case "24h":
		startTime = now.Add(-24 * time.Hour)
		interval = "15 minutes"
	case "7d":
		startTime = now.Add(-7 * 24 * time.Hour)
		interval = "1 hour"
	case "30d":
		startTime = now.Add(-30 * 24 * time.Hour)
		interval = "4 hours"
	default:
		startTime = now.Add(-1 * time.Hour)
		interval = "1 minute"
	}

	// SQL query for time-series aggregation
	query := `
		SELECT 
			date_trunc($1, timestamp) as time_bucket,
			AVG(value) as avg_value,
			host_id,
			name
		FROM metrics 
		WHERE host_id = $2 
			AND name = $3 
			AND timestamp >= $4 
			AND timestamp <= $5
		GROUP BY time_bucket, host_id, name
		ORDER BY time_bucket ASC
	`

	var results []AggregatedMetric
	rows, err := postgres.DB.Raw(query, interval, hostID, metricName, startTime, now).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate metrics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result AggregatedMetric
		if err := rows.Scan(&result.Timestamp, &result.Value, &result.HostID, &result.Name); err != nil {
			continue
		}
		results = append(results, result)
	}

	// Fill gaps in time series with interpolated values
	return s.fillTimeGaps(results, startTime, now, interval), nil
}

// fillTimeGaps fills missing time points with interpolated values
func (s *MetricsAggregationService) fillTimeGaps(data []AggregatedMetric, start, end time.Time, interval string) []AggregatedMetric {
	if len(data) == 0 {
		return data
	}

	var intervalDuration time.Duration
	switch interval {
	case "1 minute":
		intervalDuration = time.Minute
	case "5 minutes":
		intervalDuration = 5 * time.Minute
	case "15 minutes":
		intervalDuration = 15 * time.Minute
	case "1 hour":
		intervalDuration = time.Hour
	case "4 hours":
		intervalDuration = 4 * time.Hour
	default:
		intervalDuration = time.Minute
	}

	var filled []AggregatedMetric
	dataIndex := 0
	
	for current := start.Truncate(intervalDuration); current.Before(end); current = current.Add(intervalDuration) {
		if dataIndex < len(data) && data[dataIndex].Timestamp.Equal(current) {
			// Exact match - use real data
			filled = append(filled, data[dataIndex])
			dataIndex++
		} else {
			// Gap - interpolate or use last known value
			var value float64
			if len(filled) > 0 {
				value = filled[len(filled)-1].Value // Use last known value
			}
			
			filled = append(filled, AggregatedMetric{
				Timestamp: current,
				Value:     value,
				HostID:    data[0].HostID,
				Name:      data[0].Name,
			})
		}
	}

	return filled
}

// GetMultipleMetricsAggregated returns multiple metrics aggregated
func (s *MetricsAggregationService) GetMultipleMetricsAggregated(hostID int64, metricNames []string, timeRange string) (map[string][]AggregatedMetric, error) {
	result := make(map[string][]AggregatedMetric)
	
	for _, metricName := range metricNames {
		data, err := s.GetAggregatedMetrics(hostID, metricName, timeRange)
		if err != nil {
			// Continue with other metrics even if one fails
			result[metricName] = []AggregatedMetric{}
			continue
		}
		result[metricName] = data
	}
	
	return result, nil
}