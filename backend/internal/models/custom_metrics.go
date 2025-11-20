package models

import "time"

// MetricType represents different types of metrics
type MetricType string

const (
	MetricTypeGauge     MetricType = "gauge"     // Point-in-time value
	MetricTypeCounter   MetricType = "counter"   // Ever-increasing counter
	MetricTypeHistogram MetricType = "histogram" // Distribution of values
	MetricTypeSummary   MetricType = "summary"   // Summary statistics
)

// CustomMetric represents an extended metric with type support
type CustomMetric struct {
	ID        int64      `gorm:"primaryKey" db:"id" json:"id"`
	HostID    int64      `gorm:"index" db:"host_id" json:"host_id"`
	TenantID  int64      `gorm:"index" db:"tenant_id" json:"tenant_id"`
	Name      string     `gorm:"index;size:128" db:"name" json:"name"`
	Type      MetricType `gorm:"size:32" db:"type" json:"type"`
	Value     float64    `db:"value" json:"value"`
	Count     int64      `db:"count" json:"count,omitempty"`           // For histograms/summaries
	Sum       float64    `db:"sum" json:"sum,omitempty"`               // For histograms/summaries
	Min       float64    `db:"min" json:"min,omitempty"`               // For aggregations
	Max       float64    `db:"max" json:"max,omitempty"`               // For aggregations
	P50       float64    `db:"p50" json:"p50,omitempty"`               // Median
	P95       float64    `db:"p95" json:"p95,omitempty"`               // 95th percentile
	P99       float64    `db:"p99" json:"p99,omitempty"`               // 99th percentile
	Labels    string     `gorm:"type:jsonb" db:"labels" json:"labels"` // JSON labels
	Timestamp time.Time  `gorm:"index" db:"timestamp" json:"timestamp"`
	CreatedAt time.Time  `gorm:"autoCreateTime" db:"created_at" json:"created_at"`
}

// MetricAggregation stores pre-aggregated metrics for faster queries
type MetricAggregation struct {
	ID           int64     `gorm:"primaryKey" db:"id" json:"id"`
	HostID       int64     `gorm:"index" db:"host_id" json:"host_id"`
	MetricName   string    `gorm:"index;size:128" db:"metric_name" json:"metric_name"`
	Interval     string    `gorm:"size:16" db:"interval" json:"interval"` // 1m, 5m, 1h, 24h
	IntervalTime time.Time `gorm:"index" db:"interval_time" json:"interval_time"`
	Avg          float64   `db:"avg" json:"avg"`
	Min          float64   `db:"min" json:"min"`
	Max          float64   `db:"max" json:"max"`
	Sum          float64   `db:"sum" json:"sum"`
	Count        int64     `db:"count" json:"count"`
	P50          float64   `db:"p50" json:"p50"`
	P95          float64   `db:"p95" json:"p95"`
	P99          float64   `db:"p99" json:"p99"`
	CreatedAt    time.Time `gorm:"autoCreateTime" db:"created_at" json:"created_at"`
}

// MetricTrend stores trend analysis data
type MetricTrend struct {
	ID             int64     `gorm:"primaryKey" db:"id" json:"id"`
	HostID         int64     `gorm:"index" db:"host_id" json:"host_id"`
	MetricName     string    `gorm:"index;size:128" db:"metric_name" json:"metric_name"`
	TrendType      string    `gorm:"size:32" db:"trend_type" json:"trend_type"` // increasing, decreasing, stable, anomaly
	Confidence     float64   `db:"confidence" json:"confidence"`                // 0-1
	MovingAvg      float64   `db:"moving_avg" json:"moving_avg"`
	StdDev         float64   `db:"std_dev" json:"std_dev"`
	Slope          float64   `db:"slope" json:"slope"` // Trend slope
	IsAnomaly      bool      `db:"is_anomaly" json:"is_anomaly"`
	AnomalyScore   float64   `db:"anomaly_score" json:"anomaly_score"`
	PredictedValue float64   `db:"predicted_value" json:"predicted_value,omitempty"`
	AnalyzedAt     time.Time `gorm:"index" db:"analyzed_at" json:"analyzed_at"`
	CreatedAt      time.Time `gorm:"autoCreateTime" db:"created_at" json:"created_at"`
}
