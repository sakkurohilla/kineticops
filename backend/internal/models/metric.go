package models

import "time"

type Metric struct {
	ID        int64     `gorm:"primaryKey"`
	HostID    int64     `gorm:"index"`    // Foreign key to hosts table
	TenantID  int64     `gorm:"index"`    // For multi-tenancy
	Name      string    `gorm:"size:128"` // Metric name, e.g. cpu_usage, ram_usage
	Value     float64   // Metric value
	Timestamp time.Time `gorm:"index"` // When recorded
	Labels    string    // Serialized labels/tags (JSON or comma sep)
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
