package models

import "time"

// HostMetric maps to the host_metrics table used for dashboard and collector storage
type HostMetric struct {
	ID             int64     `gorm:"primaryKey" json:"id"`
	HostID         int64     `gorm:"index" json:"host_id"`
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    float64   `json:"memory_usage"`
	MemoryTotal    float64   `json:"memory_total"`
	MemoryUsed     float64   `json:"memory_used"`
	MemoryFree     float64   `json:"memory_free"`
	DiskUsage      float64   `json:"disk_usage"`
	DiskTotal      float64   `json:"disk_total"`
	DiskUsed       float64   `json:"disk_used"`
	DiskReadBytes  float64   `json:"disk_read_bytes"`
	DiskWriteBytes float64   `json:"disk_write_bytes"`
	DiskReadSpeed  float64   `json:"disk_read_speed"`
	DiskWriteSpeed float64   `json:"disk_write_speed"`
	NetworkIn      float64   `json:"network_in"`
	NetworkOut     float64   `json:"network_out"`
	Uptime         int64     `json:"uptime"`
	LoadAverage    string    `json:"load_average"`
	Timestamp      time.Time `json:"timestamp"`
}

// TableName ensures GORM uses the existing host_metrics table name
func (HostMetric) TableName() string {
	return "host_metrics"
}
