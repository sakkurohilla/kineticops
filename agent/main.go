package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

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

func main() {
	token := os.Getenv("KINETICOPS_TOKEN")
	if token == "" {
		log.Fatal("KINETICOPS_TOKEN environment variable is required")
	}

	serverURL := os.Getenv("KINETICOPS_SERVER")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	log.Printf("KineticOps Agent starting with token: %s...", token[:8])
	log.Printf("Server URL: %s", serverURL)

	for {
		if err := sendHeartbeat(serverURL, token); err != nil {
			log.Printf("Heartbeat failed: %v", err)
		}
		time.Sleep(30 * time.Second)
	}
}

func sendHeartbeat(serverURL, token string) error {
	heartbeat := AgentHeartbeat{
		Token:       token,
		CPUUsage:    getCPUUsage(),
		MemoryUsage: getMemoryUsage(),
		DiskUsage:   getDiskUsage(),
		Services:    getServices(),
		SystemInfo:  getSystemInfo(),
	}

	data, err := json.Marshal(heartbeat)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat: %v", err)
	}

	resp, err := http.Post(serverURL+"/api/v1/agents/heartbeat", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	log.Printf("Heartbeat sent successfully - CPU: %.1f%%, Memory: %.1f%%, Disk: %.1f%%", 
		heartbeat.CPUUsage, heartbeat.MemoryUsage, heartbeat.DiskUsage)
	return nil
}

func getCPUUsage() float64 {
	cmd := exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | awk '{print $2}' | sed 's/%us,//'")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	
	usage, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0
	}
	return usage
}

func getMemoryUsage() float64 {
	cmd := exec.Command("sh", "-c", "free | grep Mem | awk '{printf(\"%.1f\", $3/$2 * 100.0)}'")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	
	usage, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0
	}
	return usage
}

func getDiskUsage() float64 {
	cmd := exec.Command("sh", "-c", "df -h / | awk 'NR==2{printf(\"%.1f\", $5)}' | sed 's/%//'")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	
	usage, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0
	}
	return usage
}

func getServices() []ServiceInfo {
	services := []ServiceInfo{}
	
	// Check common services
	serviceNames := []string{"nginx", "apache2", "mysql", "postgresql", "redis-server", "docker"}
	
	for _, name := range serviceNames {
		cmd := exec.Command("pgrep", "-f", name)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			pids := strings.Fields(strings.TrimSpace(string(output)))
			if len(pids) > 0 {
				pid, _ := strconv.Atoi(pids[0])
				services = append(services, ServiceInfo{
					Name:   name,
					Status: "running",
					PID:    pid,
				})
			}
		}
	}
	
	return services
}

func getSystemInfo() map[string]interface{} {
	hostname, _ := os.Hostname()
	
	uptimeCmd := exec.Command("uptime", "-p")
	uptimeOutput, _ := uptimeCmd.Output()
	
	return map[string]interface{}{
		"hostname": hostname,
		"os":       "linux",
		"uptime":   strings.TrimSpace(string(uptimeOutput)),
	}
}