package models

import (
	"time"
)

// AuditLog represents a comprehensive security audit log entry
type AuditLog struct {
	ID           int64     `gorm:"primaryKey" db:"id" json:"id"`
	TenantID     int64     `gorm:"index" db:"tenant_id" json:"tenant_id"`
	UserID       int64     `gorm:"index" db:"user_id" json:"user_id"` // nullable for guest/system events
	Username     string    `db:"username" json:"username"`
	Event        string    `gorm:"index;size:64" db:"event" json:"event"`   // kept for backward compatibility
	Action       string    `gorm:"index;size:64" db:"action" json:"action"` // new detailed action
	Resource     string    `gorm:"size:128" db:"resource" json:"resource"`
	ResourceID   string    `gorm:"size:64" db:"resource_id" json:"resource_id"`
	IPAddress    string    `gorm:"size:45" db:"ip_address" json:"ip_address"`
	UserAgent    string    `gorm:"size:512" db:"user_agent" json:"user_agent"`
	Status       string    `gorm:"size:16" db:"status" json:"status"` // success, failed, denied
	ErrorMessage string    `gorm:"size:512" db:"error_message" json:"error_message,omitempty"`
	Details      string    `gorm:"type:text" db:"details" json:"details"` // JSON metadata or info
	Timestamp    time.Time `gorm:"index;autoCreateTime" db:"timestamp" json:"timestamp"`
	CreatedAt    time.Time `gorm:"index;autoCreateTime" db:"created_at" json:"created_at"`
}

// Common action types
const (
	AuditActionLogin            = "login"
	AuditActionLogout           = "logout"
	AuditActionLoginFailed      = "login_failed"
	AuditActionHostCreate       = "host_create"
	AuditActionHostUpdate       = "host_update"
	AuditActionHostDelete       = "host_delete"
	AuditActionAgentCreate      = "agent_create"
	AuditActionAgentRevoke      = "agent_revoke"
	AuditActionAgentUnrevoke    = "agent_unrevoke"
	AuditActionTokenRotate      = "token_rotate"
	AuditActionAlertCreate      = "alert_create"
	AuditActionAlertUpdate      = "alert_update"
	AuditActionAlertDelete      = "alert_delete"
	AuditActionConfigChange     = "config_change"
	AuditActionWorkflowExecute  = "workflow_execute"
	AuditActionUserCreate       = "user_create"
	AuditActionUserUpdate       = "user_update"
	AuditActionUserDelete       = "user_delete"
	AuditActionPasswordChange   = "password_change"
	AuditActionPermissionChange = "permission_change"
)
