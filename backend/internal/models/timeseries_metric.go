package models

import "time"

// TimeseriesMetric optimized for time-series storage and queries
type TimeseriesMetric struct {
	Time       time.Time `gorm:"primaryKey;index:idx_time_host_metric" json:"time"`
	HostID     int64     `gorm:"primaryKey;index:idx_time_host_metric" json:"host_id"`
	MetricName string    `gorm:"primaryKey;index:idx_time_host_metric;size:50" json:"metric_name"`
	Value      float64   `json:"value"`
	TenantID   int64     `gorm:"index" json:"tenant_id"`
	Labels     string    `gorm:"type:jsonb" json:"labels,omitempty"` // JSON for additional metadata
}

// TableName ensures GORM uses the correct table name
func (TimeseriesMetric) TableName() string {
	return "timeseries_metrics"
}

// MetricBatch for bulk inserts (enterprise performance)
type MetricBatch struct {
	Metrics []TimeseriesMetric `json:"metrics"`
}

// Common metric names as constants
const (
	MetricCPUUsage     = "cpu_usage"
	MetricMemoryUsage  = "memory_usage"
	MetricDiskUsage    = "disk_usage"
	MetricNetworkIn    = "network_in"
	MetricNetworkOut   = "network_out"
	MetricLoadAverage  = "load_average"
	MetricUptime       = "uptime"
)