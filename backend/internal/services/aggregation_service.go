package services

import (
	"fmt"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

type AggregationService struct {
	db interface{}
}

type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type AggregationQuery struct {
	MetricName string
	HostID     int64
	StartTime  time.Time
	EndTime    time.Time
	Interval   string // "1m", "5m", "1h", "1d"
	Function   string // "avg", "max", "min", "sum"
}

func NewAggregationService() *AggregationService {
	return &AggregationService{
		db: postgres.DB,
	}
}

// QueryTimeSeries executes time series query with proper aggregation
func (a *AggregationService) QueryTimeSeries(query AggregationQuery) ([]TimeSeriesPoint, error) {
	// Validate inputs
	if query.StartTime.After(query.EndTime) {
		return nil, fmt.Errorf("start time cannot be after end time")
	}

	// Auto-select optimal interval based on time range
	if query.Interval == "" {
		query.Interval = a.selectOptimalInterval(query.StartTime, query.EndTime)
	}

	// Default aggregation function
	if query.Function == "" {
		query.Function = "avg"
	}

	// Build optimized query
	sqlQuery := a.buildTimeSeriesQuery(query)

	rows, err := postgres.DB.Raw(sqlQuery, query.HostID, query.MetricName, query.StartTime, query.EndTime).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var point TimeSeriesPoint
		if err := rows.Scan(&point.Timestamp, &point.Value); err != nil {
			continue
		}
		points = append(points, point)
	}

	return points, nil
}

// selectOptimalInterval chooses best interval based on time range
func (a *AggregationService) selectOptimalInterval(start, end time.Time) string {
	duration := end.Sub(start)

	switch {
	case duration <= 1*time.Hour:
		return "1m"
	case duration <= 6*time.Hour:
		return "5m"
	case duration <= 24*time.Hour:
		return "15m"
	case duration <= 7*24*time.Hour:
		return "1h"
	default:
		return "1d"
	}
}

// buildTimeSeriesQuery creates optimized SQL for time series aggregation
func (a *AggregationService) buildTimeSeriesQuery(query AggregationQuery) string {
	// Select optimal table based on time range and interval
	tableName := a.selectOptimalTable(query.StartTime, query.EndTime, query.Interval)
	intervalSQL := a.getIntervalSQL(query.Interval)
	functionSQL := a.getFunctionSQL(query.Function)

	if tableName == "metrics" {
		// Use raw data with aggregation
		return fmt.Sprintf(`
			SELECT 
				date_trunc('%s', timestamp) as bucket,
				%s(value) as aggregated_value
			FROM %s 
			WHERE host_id = $1 
				AND name = $2 
				AND timestamp >= $3 
				AND timestamp <= $4
			GROUP BY bucket 
			ORDER BY bucket ASC
		`, intervalSQL, functionSQL, tableName)
	} else {
		// Use pre-aggregated data
		return fmt.Sprintf(`
			SELECT 
				timestamp as bucket,
				%s(value) as aggregated_value
			FROM %s 
			WHERE host_id = $1 
				AND name = $2 
				AND timestamp >= $3 
				AND timestamp <= $4
			GROUP BY timestamp 
			ORDER BY timestamp ASC
		`, functionSQL, tableName)
	}
}

// selectOptimalTable chooses best table for query performance
func (a *AggregationService) selectOptimalTable(start, end time.Time, interval string) string {
	duration := end.Sub(start)

	// Use downsampled tables for longer ranges
	switch {
	case duration > 7*24*time.Hour && (interval == "1h" || interval == "1d"):
		return "metrics_1h"
	case duration > 30*24*time.Hour && interval == "1d":
		return "metrics_1d"
	case duration > 2*time.Hour && (interval == "5m" || interval == "15m"):
		return "metrics_5m"
	default:
		return "metrics"
	}
}

// getIntervalSQL converts interval string to PostgreSQL interval
func (a *AggregationService) getIntervalSQL(interval string) string {
	switch interval {
	case "1m":
		return "minute"
	case "5m":
		return "5 minutes"
	case "15m":
		return "15 minutes"
	case "1h":
		return "hour"
	case "1d":
		return "day"
	default:
		return "minute"
	}
}

// getFunctionSQL converts function string to SQL aggregation
func (a *AggregationService) getFunctionSQL(function string) string {
	switch function {
	case "avg":
		return "AVG"
	case "max":
		return "MAX"
	case "min":
		return "MIN"
	case "sum":
		return "SUM"
	case "count":
		return "COUNT"
	default:
		return "AVG"
	}
}

// GetMetricsForDashboard returns optimized metrics for dashboard display
func (a *AggregationService) GetMetricsForDashboard(hostID int64, timeRange string) (map[string][]TimeSeriesPoint, error) {
	end := time.Now()
	start := a.parseTimeRange(timeRange, end)

	metrics := []string{"cpu_usage", "memory_usage", "disk_usage"}
	result := make(map[string][]TimeSeriesPoint)

	for _, metric := range metrics {
		query := AggregationQuery{
			MetricName: metric,
			HostID:     hostID,
			StartTime:  start,
			EndTime:    end,
			Function:   "avg",
		}

		points, err := a.QueryTimeSeries(query)
		if err != nil {
			result[metric] = []TimeSeriesPoint{}
		} else {
			result[metric] = points
		}
	}

	return result, nil
}

// parseTimeRange converts time range string to start time
func (a *AggregationService) parseTimeRange(timeRange string, end time.Time) time.Time {
	switch timeRange {
	case "1h":
		return end.Add(-1 * time.Hour)
	case "6h":
		return end.Add(-6 * time.Hour)
	case "24h":
		return end.Add(-24 * time.Hour)
	case "7d":
		return end.Add(-7 * 24 * time.Hour)
	case "30d":
		return end.Add(-30 * 24 * time.Hour)
	default:
		return end.Add(-24 * time.Hour)
	}
}

// GetLatestMetrics returns most recent metric values for real-time display
func (a *AggregationService) GetLatestMetrics(hostID int64) (map[string]float64, error) {
	query := `
		SELECT DISTINCT ON (name) name, value 
		FROM metrics 
		WHERE host_id = $1 
		ORDER BY name, timestamp DESC
	`

	rows, err := postgres.DB.Raw(query, hostID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			continue
		}
		result[name] = value
	}

	return result, nil
}
