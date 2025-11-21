package metrics

import (
	"sort"
	"time"

	"github.com/sakkurohilla/kineticops/agent/utils"
	"github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo holds information about a single process
type ProcessInfo struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	Username      string  `json:"username"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
	MemoryRSS     uint64  `json:"memory_rss"` // bytes
	Status        string  `json:"status"`
	NumThreads    int32   `json:"num_threads"`
	CreateTime    int64   `json:"create_time"`
}

// GetTopProcesses returns top N processes sorted by CPU or memory usage
func GetTopProcesses(topN int, sortBy string, logger *utils.Logger) ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		logger.Error("Failed to get process list", "error", err)
		return nil, err
	}

	var processList []ProcessInfo

	for _, p := range procs {
		// Get process info with error handling for each field
		pid := p.Pid

		name, _ := p.Name()
		if name == "" {
			name = "unknown"
		}

		username, _ := p.Username()
		if username == "" {
			username = "unknown"
		}

		// Get CPU percent - gopsutil returns per-core % by default
		// We need total system %, not per-core, so this value is already correct
		cpuPercent, err := p.CPUPercent()
		if err != nil {
			cpuPercent = 0.0
		}

		// Get memory percent
		memInfo, err := p.MemoryInfo()
		var memoryRSS uint64 = 0
		if err == nil && memInfo != nil {
			memoryRSS = memInfo.RSS
		}

		memPercent, err := p.MemoryPercent()
		if err != nil {
			memPercent = 0.0
		}

		status, _ := p.Status()
		if len(status) == 0 {
			status = []string{"unknown"}
		}

		numThreads, _ := p.NumThreads()
		createTime, _ := p.CreateTime()

		processList = append(processList, ProcessInfo{
			PID:           pid,
			Name:          name,
			Username:      username,
			CPUPercent:    cpuPercent,
			MemoryPercent: memPercent,
			MemoryRSS:     memoryRSS,
			Status:        status[0],
			NumThreads:    numThreads,
			CreateTime:    createTime,
		})
	}

	// Sort by requested metric
	if sortBy == "cpu" {
		sort.Slice(processList, func(i, j int) bool {
			return processList[i].CPUPercent > processList[j].CPUPercent
		})
	} else {
		// Default to memory
		sort.Slice(processList, func(i, j int) bool {
			return processList[i].MemoryPercent > processList[j].MemoryPercent
		})
	}

	// Return top N
	if len(processList) > topN {
		processList = processList[:topN]
	}

	return processList, nil
}

// CollectProcessMetrics collects process data for sending to backend
func CollectProcessMetrics(logger *utils.Logger) map[string]interface{} {
	logger.Info("CollectProcessMetrics called - starting process collection")

	// Get top 10 by CPU
	topCPU, err := GetTopProcesses(10, "cpu", logger)
	if err != nil {
		logger.Error("Failed to get top CPU processes", "error", err)
		topCPU = []ProcessInfo{}
	}

	// Get top 10 by memory
	topMemory, err := GetTopProcesses(10, "memory", logger)
	if err != nil {
		logger.Error("Failed to get top memory processes", "error", err)
		topMemory = []ProcessInfo{}
	}

	// Wait a bit to get accurate CPU measurements
	time.Sleep(100 * time.Millisecond)

	return map[string]interface{}{
		"top_cpu":    topCPU,
		"top_memory": topMemory,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}
}

// ApplicationInfo holds information about detected applications
type ApplicationInfo struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"` // node, python, java, etc.
	PID        int32   `json:"pid"`
	CPUPercent float64 `json:"cpu_percent"`
	MemoryMB   float64 `json:"memory_mb"`
	Port       int     `json:"port,omitempty"` // if detectable
	CmdLine    string  `json:"cmdline,omitempty"`
}

// DetectApplications identifies running applications from processes
func DetectApplications(logger *utils.Logger) []ApplicationInfo {
	procs, err := process.Processes()
	if err != nil {
		logger.Error("Failed to get process list for application detection", "error", err)
		return []ApplicationInfo{}
	}

	var applications []ApplicationInfo
	appTypes := map[string]string{
		"node":         "Node.js",
		"npm":          "Node.js",
		"python":       "Python",
		"python3":      "Python",
		"java":         "Java",
		"php-fpm":      "PHP",
		"php":          "PHP",
		"ruby":         "Ruby",
		"dotnet":       ".NET",
		"mongod":       "MongoDB",
		"redis-server": "Redis",
		"postgres":     "PostgreSQL",
		"mysql":        "MySQL",
		"nginx":        "Nginx",
		"apache2":      "Apache",
		"httpd":        "Apache",
	}

	for _, p := range procs {
		name, _ := p.Name()
		if name == "" {
			continue
		}

		// Check if this is a known application type
		appType, isApp := appTypes[name]
		if !isApp {
			continue
		}

		// Get process details
		cpuPercent, _ := p.CPUPercent()
		memInfo, _ := p.MemoryInfo()
		var memoryMB float64
		if memInfo != nil {
			memoryMB = float64(memInfo.RSS) / (1024 * 1024)
		}

		cmdline, _ := p.Cmdline()

		applications = append(applications, ApplicationInfo{
			Name:       name,
			Type:       appType,
			PID:        p.Pid,
			CPUPercent: cpuPercent,
			MemoryMB:   memoryMB,
			CmdLine:    cmdline,
		})
	}

	logger.Info("Detected applications", "count", len(applications))
	return applications
}
