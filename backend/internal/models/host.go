package models

import "time"

type Host struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Hostname    string    `gorm:"not null" json:"hostname"`
	IP          string    `gorm:"not null" json:"ip"`
	OS          string    `json:"os"`
	AgentStatus string    `json:"agent_status"`
	Status      string    `json:"status"`
	TenantID    int64     `json:"tenant_id"`
	Tags        string    `json:"tags"`
	Group       string    `json:"group"`
	Description string    `json:"description"`
	LastSeen    time.Time `json:"last_seen"`
	LastSync    time.Time `json:"last_sync"`
	RegToken    string    `json:"reg_token"`

	// SSH Configuration
	SSHUser     string `gorm:"default:'root'" json:"ssh_user"`
	SSHPassword string `json:"ssh_password,omitempty"` // omitempty for security
	SSHPort     int64  `gorm:"default:22" json:"ssh_port"`
	SSHKey      string `json:"ssh_key,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
