package models

import "time"

type WorkflowSession struct {
	ID              int       `json:"id" db:"id"`
	HostID          int       `json:"host_id" db:"host_id"`
	UserID          int       `json:"user_id" db:"user_id"`
	AgentID         *int      `json:"agent_id" db:"agent_id"`
	SessionToken    string    `json:"session_token" db:"session_token"`
	Status          string    `json:"status" db:"status"`
	AuthenticatedAt time.Time `json:"authenticated_at" db:"authenticated_at"`
	ExpiresAt       time.Time `json:"expires_at" db:"expires_at"`
	LastActivity    time.Time `json:"last_activity" db:"last_activity"`
	// Credentials stored encrypted in session (for the duration of the session only)
	Username string `json:"-" db:"username"`
	Password string `json:"-" db:"password_encrypted"`
	SSHKey   string `json:"-" db:"ssh_key_encrypted"`
}

type ServiceControlLog struct {
	ID           int       `json:"id" db:"id"`
	ServiceID    *int      `json:"service_id" db:"service_id"`
	ServiceName  string    `json:"service_name" db:"service_name"`
	HostID       int       `json:"host_id" db:"host_id"`
	Action       string    `json:"action" db:"action"`
	Status       string    `json:"status" db:"status"`
	Output       string    `json:"output" db:"output"`
	ErrorMessage string    `json:"error_message" db:"error_message"`
	ExecutedBy   int       `json:"executed_by" db:"executed_by"`
	ExecutedAt   time.Time `json:"executed_at" db:"executed_at"`
}

type ControlAction string

const (
	ActionStart   ControlAction = "start"
	ActionStop    ControlAction = "stop"
	ActionRestart ControlAction = "restart"
	ActionEnable  ControlAction = "enable"
	ActionDisable ControlAction = "disable"
)

type WorkflowSessionRequest struct {
	HostID   int    `json:"host_id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	SSHKey   string `json:"ssh_key,omitempty"`
}

type WorkflowSessionResponse struct {
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	HostID       int       `json:"host_id"`
	Status       string    `json:"status"`
}

type ServiceControlRequest struct {
	Action ControlAction `json:"action"`
}

type ServiceControlResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}
