package models

import "time"

type Host struct {
	ID          int64     `gorm:"primaryKey" db:"id" json:"id"`
	Hostname    string    `gorm:"not null" db:"hostname" json:"hostname"`
	IP          string    `gorm:"not null" db:"ip" json:"ip"`
	OS          string    `db:"os" json:"os"`
	AgentStatus string    `db:"agent_status" json:"agent_status"`
	Status      string    `db:"status" json:"status"`
	TenantID    int64     `db:"tenant_id" json:"tenant_id"`
	Tags        string    `db:"tags" json:"tags"`
	Group       string    `db:"group" json:"group"`
	Description string    `db:"description" json:"description"`
	LastSeen    time.Time `db:"last_seen" json:"last_seen"`
	LastSync    time.Time `db:"last_sync" json:"last_sync"`
	RegToken    string    `db:"reg_token" json:"reg_token"`

	// SSH Configuration
	SSHUser     string `db:"ssh_user" json:"ssh_user"`
	SSHPassword string `db:"ssh_password" json:"ssh_password,omitempty"` // omitempty for security
	SSHPort     int64  `gorm:"default:22" db:"ssh_port" json:"ssh_port"`
	SSHKey      string `db:"ssh_key" json:"ssh_key,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime" db:"created_at" json:"created_at"`
}
