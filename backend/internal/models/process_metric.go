package models

import "time"

// ProcessMetric represents per-process metrics collected from agents
type ProcessMetric struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID      int64     `json:"tenant_id" gorm:"not null;index:idx_process_metrics_tenant_id"`
	HostID        int64     `json:"host_id" gorm:"not null;index:idx_process_metrics_host_id"`
	PID           int       `json:"pid" gorm:"column:pid;not null"`
	Name          string    `json:"name" gorm:"type:varchar(255);not null"`
	Username      string    `json:"username" gorm:"type:varchar(100)"`
	CPUPercent    float64   `json:"cpu_percent" gorm:"type:decimal(5,2)"`
	MemoryPercent float64   `json:"memory_percent" gorm:"type:decimal(5,2)"`
	MemoryRSS     int64     `json:"memory_rss"`
	Status        string    `json:"status" gorm:"type:varchar(50)"`
	NumThreads    int       `json:"num_threads"`
	CreateTime    int64     `json:"create_time"`
	Timestamp     time.Time `json:"timestamp" gorm:"not null;index:idx_process_metrics_timestamp;index:idx_process_metrics_host_timestamp,priority:2"`
}

// TableName specifies the table name for GORM
func (ProcessMetric) TableName() string {
	return "process_metrics"
}
