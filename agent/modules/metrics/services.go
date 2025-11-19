package metrics

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sakkurohilla/kineticops/agent/utils"
	"github.com/shirou/gopsutil/v3/process"
)

// ServiceInfo holds information about a systemd service
type ServiceInfo struct {
	Name          string  `json:"name"`
	DisplayName   string  `json:"display_name"`
	Status        string  `json:"status"`          // active, inactive, failed
	SubStatus     string  `json:"sub_status"`      // running, exited, dead
	CPUPercent    float64 `json:"cpu_percent"`     // Current CPU usage
	MemoryPercent float64 `json:"memory_percent"`  // Current memory usage
	MemoryMB      float64 `json:"memory_mb"`       // Memory in MB
	PID           int32   `json:"pid"`             // Main PID
	RestartCount  int     `json:"restart_count"`   // Number of restarts
	Enabled       bool    `json:"enabled"`         // Auto-start enabled
	FailureReason string  `json:"failure_reason"`  // Reason for failure if failed
	IsUserService bool    `json:"is_user_service"` // User-installed vs system service
}

// GetTopServices returns top N services sorted by CPU or memory usage
func GetTopServices(topN int, sortBy string, logger *utils.Logger) ([]ServiceInfo, error) {
	// Get list of all services
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager", "--output=json")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		logger.Error("Failed to list systemd services", "error", err)
		return nil, err
	}

	var services []map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &services); err != nil {
		logger.Error("Failed to parse systemctl output", "error", err)
		return nil, err
	}

	var serviceList []ServiceInfo

	// Process each service
	for _, svc := range services {
		unit, _ := svc["unit"].(string)
		if unit == "" || !strings.HasSuffix(unit, ".service") {
			continue
		}

		name := strings.TrimSuffix(unit, ".service")

		// Get service properties
		props := getServiceProperties(name, logger)
		if props == nil {
			continue
		}

		// Get resource usage
		cpuPercent, memPercent, memMB := getServiceResources(props.PID, logger)

		// Determine if user-installed service
		isUserService := isUserInstalledService(name)

		serviceList = append(serviceList, ServiceInfo{
			Name:          name,
			DisplayName:   props.Description,
			Status:        props.ActiveState,
			SubStatus:     props.SubState,
			CPUPercent:    cpuPercent,
			MemoryPercent: memPercent,
			MemoryMB:      memMB,
			PID:           props.PID,
			RestartCount:  props.NRestarts,
			Enabled:       props.UnitFileState == "enabled",
			FailureReason: props.FailureReason,
			IsUserService: isUserService,
		})
	}

	// Filter: Show ALL user services regardless of state
	// Only exclude system services entirely
	var activeServices []ServiceInfo
	for _, svc := range serviceList {
		// Show ONLY user-installed services (any state: active, inactive, failed, etc.)
		if svc.IsUserService {
			activeServices = append(activeServices, svc)
		}
		// System services are NEVER shown
	}

	// Sort by requested metric
	if sortBy == "cpu" {
		// Simple bubble sort by CPU
		for i := 0; i < len(activeServices); i++ {
			for j := i + 1; j < len(activeServices); j++ {
				if activeServices[i].CPUPercent < activeServices[j].CPUPercent {
					activeServices[i], activeServices[j] = activeServices[j], activeServices[i]
				}
			}
		}
	} else {
		// Sort by memory
		for i := 0; i < len(activeServices); i++ {
			for j := i + 1; j < len(activeServices); j++ {
				if activeServices[i].MemoryPercent < activeServices[j].MemoryPercent {
					activeServices[i], activeServices[j] = activeServices[j], activeServices[i]
				}
			}
		}
	}

	// Return top N
	if len(activeServices) > topN {
		activeServices = activeServices[:topN]
	}

	return activeServices, nil
}

// ServiceProperties holds systemd service properties
type ServiceProperties struct {
	Description   string
	ActiveState   string
	SubState      string
	PID           int32
	NRestarts     int
	UnitFileState string
	FailureReason string
}

// getServiceProperties retrieves properties for a service
func getServiceProperties(name string, _ *utils.Logger) *ServiceProperties {
	cmd := exec.Command("systemctl", "show", name, "--no-pager")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil
	}

	props := &ServiceProperties{}
	lines := strings.Split(out.String(), "\n")

	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "Description":
			props.Description = value
		case "ActiveState":
			props.ActiveState = value
		case "SubState":
			props.SubState = value
			// Normalize SubState: "dead" -> "stopped" for better UX
			if value == "dead" {
				props.SubState = "stopped"
			}
		case "MainPID":
			if pid, err := strconv.ParseInt(value, 10, 32); err == nil {
				props.PID = int32(pid)
			}
		case "NRestarts":
			// NRestarts only counts automatic restarts after failures
			// For manual restarts, we'd need to parse journal timestamps
			if restarts, err := strconv.Atoi(value); err == nil {
				props.NRestarts = restarts
			}
		case "UnitFileState":
			props.UnitFileState = value
		case "Result":
			if value != "success" {
				props.FailureReason = value
			}
		case "StatusText":
			if props.FailureReason == "" && value != "" {
				props.FailureReason = value
			}
		}
	}

	return props
}

// isUserInstalledService determines if a service is user-installed (not system default)
func isUserInstalledService(name string) bool {
	// STRICT WHITELIST APPROACH - Only show services matching these patterns
	// This ensures we ONLY show real user-installed server software

	userServicePatterns := []string{
		"docker", "dockerd",
		"mysql", "mysqld", "mariadb",
		"postgres", "postgresql",
		"mongodb", "mongod", "mongo",
		"nginx",
		"apache", "apache2", "httpd",
		"tomcat", "tomcat7", "tomcat8", "tomcat9",
		"redis", "redis-server",
		"memcached",
		"elasticsearch",
		"kibana",
		"logstash",
		"jenkins",
		"gitlab",
		"prometheus",
		"grafana",
		"grafana-server",
		"node",
		"php-fpm", "php7", "php8",
		"rabbitmq", "rabbitmq-server",
		"kafka",
		"zookeeper",
		"cassandra",
		"kineticops-agent", "kineticops",
		"haproxy",
		"varnish",
		"squid",
		"bind9", "named",
		"postfix", "dovecot",
		"vsftpd", "proftpd",
	}

	nameLower := strings.ToLower(name)

	// Check if service name contains any user service pattern
	for _, pattern := range userServicePatterns {
		if strings.Contains(nameLower, pattern) {
			return true
		}
	}

	// NOT in whitelist = EXCLUDE IT
	return false
}

// getServiceResources gets CPU and memory usage for a service PID
func getServiceResources(pid int32, _ *utils.Logger) (float64, float64, float64) {
	if pid <= 0 {
		return 0, 0, 0
	}

	proc, err := process.NewProcess(pid)
	if err != nil {
		return 0, 0, 0
	}

	cpuPercent, _ := proc.CPUPercent()
	memPercent, _ := proc.MemoryPercent()

	memInfo, err := proc.MemoryInfo()
	var memMB float64 = 0
	if err == nil && memInfo != nil {
		memMB = float64(memInfo.RSS) / (1024 * 1024)
	}

	return cpuPercent, float64(memPercent), memMB
}

// CollectServiceMetrics collects service data for sending to backend
func CollectServiceMetrics(logger *utils.Logger) map[string]interface{} {
	logger.Info("CollectServiceMetrics called - starting service collection")

	// Get top 10 by CPU
	topCPU, err := GetTopServices(10, "cpu", logger)
	if err != nil {
		logger.Error("Failed to get top CPU services", "error", err)
		topCPU = []ServiceInfo{}
	}

	// Get top 10 by memory
	topMemory, err := GetTopServices(10, "memory", logger)
	if err != nil {
		logger.Error("Failed to get top memory services", "error", err)
		topMemory = []ServiceInfo{}
	}

	return map[string]interface{}{
		"top_cpu":    topCPU,
		"top_memory": topMemory,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}
}
