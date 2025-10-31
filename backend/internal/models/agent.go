package models

import (
	"encoding/json"
	"time"
)

type Agent struct {
	ID              int             `json:"id" db:"id"`
	HostID          int             `json:"host_id" db:"host_id"`
	Token           string          `json:"token" db:"token"`
	Status          string          `json:"status" db:"status"`
	Version         string          `json:"version" db:"version"`
	OSInfo          json.RawMessage `json:"os_info" db:"os_info"`
	SystemInfo      json.RawMessage `json:"system_info" db:"system_info"`
	LastHeartbeat   *time.Time      `json:"last_heartbeat" db:"last_heartbeat"`
	InstallationLog string          `json:"installation_log" db:"installation_log"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

type HostService struct {
	ID           int       `json:"id" db:"id"`
	HostID       int       `json:"host_id" db:"host_id"`
	Name         string    `json:"name" db:"name"`
	Status       string    `json:"status" db:"status"`
	PID          int       `json:"pid" db:"pid"`
	MemoryUsage  int64     `json:"memory_usage" db:"memory_usage"`
	CPUUsage     float64   `json:"cpu_usage" db:"cpu_usage"`
	DiscoveredAt time.Time `json:"discovered_at" db:"discovered_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type AgentHeartbeat struct {
	Token       string                 `json:"token"`
	CPUUsage    float64               `json:"cpu_usage"`
	MemoryUsage float64               `json:"memory_usage"`
	DiskUsage   float64               `json:"disk_usage"`
	Services    []ServiceInfo         `json:"services"`
	SystemInfo  map[string]interface{} `json:"system_info"`
}

type ServiceInfo struct {
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	PID         int     `json:"pid"`
	MemoryUsage int64   `json:"memory_usage"`
	CPUUsage    float64 `json:"cpu_usage"`
}

type AgentSetupRequest struct {
	SetupMethod string `json:"setup_method"` // "automatic" or "manual"
	Hostname    string `json:"hostname"`
	IP          string `json:"ip"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	SSHKey      string `json:"ssh_key,omitempty"`
	Port        int    `json:"port,omitempty"`
}

type AgentSetupResponse struct {
	HostID        int    `json:"host_id"`
	AgentID       int    `json:"agent_id"`
	Token         string `json:"token"`
	SetupMethod   string `json:"setup_method"`
	InstallScript string `json:"install_script,omitempty"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}