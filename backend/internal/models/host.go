package models

import "time"

type Host struct {
	ID          int64  `gorm:"primaryKey"`
	Hostname    string `gorm:"unique"`
	IP          string
	OS          string
	AgentStatus string
	TenantID    int64
	Tags        string // Comma-separated tags or use separate table for advanced tagging
	Group       string
	LastSeen    time.Time
	RegToken    string    // For agent registration
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}
