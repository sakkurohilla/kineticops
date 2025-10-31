package models

import (
	"encoding/json"
	"time"
)

type Agent struct {
	ID             int             `json:"id" db:"id"`
	HostID         int             `json:"host_id" db:"host_id"`
	AgentToken     string          `json:"agent_token" db:"agent_token"`
	Status         string          `json:"status" db:"status"`
	Version        string          `json:"version" db:"version"`
	SetupMethod    string          `json:"setup_method" db:"setup_method"`
	InstalledAt    *time.Time      `json:"installed_at" db:"installed_at"`
	LastHeartbeat  *time.Time      `json:"last_heartbeat" db:"last_heartbeat"`
	CPUUsage       float64         `json:"cpu_usage" db:"cpu_usage"`
	MemoryUsage    float64         `json:"memory_usage" db:"memory_usage"`
	DiskUsage      float64         `json:"disk_usage" db:"disk_usage"`
	ServicesCount  int             `json:"services_count" db:"services_count"`
	Metadata       json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

type AgentMetadata struct {
	OS       string `json:"os"`
	Hostname string `json:"hostname"`
	Arch     string `json:"arch"`
	Kernel   string `json:"kernel"`
	Memory   int64  `json:"memory"`
	Disk     int64  `json:"disk"`
	Cores    int    `json:"cores"`
	BootTime int64  `json:"boot_time"`
}

type AgentService struct {
	ID          int       `json:"id" db:"id"`
	AgentID     int       `json:"agent_id" db:"agent_id"`
	ServiceName string    `json:"service_name" db:"service_name"`
	Status      string    `json:"status" db:"status"`
	Port        int       `json:"port" db:"port"`
	ProcessID   int       `json:"process_id" db:"process_id"`
	CPUUsage    float64   `json:"cpu_usage" db:"cpu_usage"`
	MemoryUsage int64     `json:"memory_usage" db:"memory_usage"`
	Uptime      int64     `json:"uptime" db:"uptime"`
	LastCheck   time.Time `json:"last_check" db:"last_check"`
}

type AgentInstallationLog struct {
	ID           int        `json:"id" db:"id"`
	AgentID      int        `json:"agent_id" db:"agent_id"`
	SetupMethod  string     `json:"setup_method" db:"setup_method"`
	Status       string     `json:"status" db:"status"`
	Logs         string     `json:"logs" db:"logs"`
	ErrorMessage string     `json:"error_message" db:"error_message"`
	StartedAt    time.Time  `json:"started_at" db:"started_at"`
	CompletedAt  *time.Time `json:"completed_at" db:"completed_at"`
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
	Metadata    AgentMetadata         `json:"metadata"`
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