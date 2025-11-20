package models

import "time"

// AgentToken represents an agent authentication token with rotation support
type AgentToken struct {
	ID            int64      `gorm:"primaryKey" db:"id" json:"id"`
	AgentID       int64      `gorm:"index" db:"agent_id" json:"agent_id"`
	Token         string     `gorm:"uniqueIndex;size:128" db:"token" json:"token"`
	PreviousToken string     `gorm:"size:128" db:"previous_token" json:"previous_token,omitempty"`
	Version       int        `db:"version" json:"version"`
	IsActive      bool       `gorm:"index" db:"is_active" json:"is_active"`
	ExpiresAt     *time.Time `gorm:"index" db:"expires_at" json:"expires_at,omitempty"`
	RotatedAt     *time.Time `db:"rotated_at" json:"rotated_at,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" db:"created_at" json:"created_at"`
	RevokedAt     *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
}

// AgentHealth represents agent health check status
type AgentHealth struct {
	ID              int64     `gorm:"primaryKey" db:"id" json:"id"`
	AgentID         int64     `gorm:"uniqueIndex" db:"agent_id" json:"agent_id"`
	HealthScore     int       `db:"health_score" json:"health_score"`    // 0-100
	Status          string    `gorm:"size:32" db:"status" json:"status"` // healthy, degraded, unhealthy, offline
	LastHeartbeat   time.Time `db:"last_heartbeat" json:"last_heartbeat"`
	HeartbeatMissed int       `db:"heartbeat_missed" json:"heartbeat_missed"`
	AvgLatency      float64   `db:"avg_latency" json:"avg_latency"`                     // ms
	ErrorRate       float64   `db:"error_rate" json:"error_rate"`                       // percentage
	DataQuality     float64   `db:"data_quality" json:"data_quality"`                   // 0-1
	Metrics         string    `gorm:"type:jsonb" db:"metrics" json:"metrics,omitempty"` // JSON health metrics
	UpdatedAt       time.Time `gorm:"autoUpdateTime" db:"updated_at" json:"updated_at"`
}

func (AgentHealth) TableName() string {
	return "agent_health"
}

// AgentVersion represents agent version information
type AgentVersion struct {
	Version       string    `gorm:"primaryKey;size:32" db:"version" json:"version"`
	ReleaseDate   time.Time `db:"release_date" json:"release_date"`
	IsLatest      bool      `db:"is_latest" json:"is_latest"`
	IsMandatory   bool      `db:"is_mandatory" json:"is_mandatory"` // Force upgrade
	MinCompatible string    `gorm:"size:32" db:"min_compatible" json:"min_compatible"`
	Changelog     string    `gorm:"type:text" db:"changelog" json:"changelog"`
	DownloadURL   string    `gorm:"size:512" db:"download_url" json:"download_url"`
	Checksum      string    `gorm:"size:128" db:"checksum" json:"checksum"`
	CreatedAt     time.Time `gorm:"autoCreateTime" db:"created_at" json:"created_at"`
}
