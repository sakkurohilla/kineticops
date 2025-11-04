package metrics

import (
	"context"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"

	"github.com/sakkurohilla/kineticops/agent/config"
	"github.com/sakkurohilla/kineticops/agent/pipelines"
	"github.com/sakkurohilla/kineticops/agent/utils"
)

// SystemModule collects system metrics
type SystemModule struct {
	config   *config.SystemModule
	pipeline *pipelines.PipelineManager
	logger   *utils.Logger
	stopChan chan struct{}
}

// NewSystemModule creates a new system metrics module
func NewSystemModule(cfg *config.SystemModule, pipeline *pipelines.PipelineManager, logger *utils.Logger) (*SystemModule, error) {
	return &SystemModule{
		config:   cfg,
		pipeline: pipeline,
		logger:   logger,
		stopChan: make(chan struct{}),
	}, nil
}

// Name returns the module name
func (s *SystemModule) Name() string {
	return "system"
}

// IsEnabled returns whether the module is enabled
func (s *SystemModule) IsEnabled() bool {
	return s.config.Enabled
}

// Start begins collecting system metrics
func (s *SystemModule) Start(ctx context.Context) error {
	s.logger.Info("Starting system metrics collection", "period", s.config.Period)

	ticker := time.NewTicker(s.config.Period)
	defer ticker.Stop()

	// Collect initial metrics
	s.logger.Info("Collecting initial metrics")
	if err := s.collectMetrics(); err != nil {
		s.logger.Error("Failed to collect initial metrics", "error", err)
	} else {
		s.logger.Info("Initial metrics collected successfully")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-s.stopChan:
			return nil
		case <-ticker.C:
			s.logger.Info("Collecting metrics on timer")
			if err := s.collectMetrics(); err != nil {
				s.logger.Error("Failed to collect metrics", "error", err)
			} else {
				s.logger.Info("Metrics collected successfully")
			}
		}
	}
}

// Stop stops the metrics collection
func (s *SystemModule) Stop() error {
	close(s.stopChan)
	return nil
}

// collectMetrics gathers all system metrics with proper validation
func (s *SystemModule) collectMetrics() error {
	timestamp := time.Now().UTC()
	hostInfo, err := host.Info()
	if err != nil {
		s.logger.Error("Failed to get host info", "error", err)
		return err
	}

	// Get all network interfaces and IPs
	networkIPs := s.getAllNetworkIPs()
	if len(networkIPs) == 0 {
		s.logger.Warn("No network interfaces found")
		networkIPs = []string{"127.0.0.1"}
	}

	// Create comprehensive event with validation
	event := map[string]interface{}{
		"@timestamp": timestamp.Format(time.RFC3339),
		"agent": map[string]interface{}{
			"name":    "kineticops-agent",
			"type":    "metricbeat",
			"version": "1.0.0",
		},
		"host": map[string]interface{}{
			"hostname":        hostInfo.Hostname,
			"ips":             networkIPs,
			"primary_ip":      networkIPs[0],
			"os":              hostInfo.OS,
			"platform":        hostInfo.Platform,
			"platform_family": hostInfo.PlatformFamily,
			"platform_version": hostInfo.PlatformVersion,
			"arch":            hostInfo.KernelArch,
			"kernel_version":  hostInfo.KernelVersion,
			"virtualization":  hostInfo.VirtualizationSystem,
		},
		"event": map[string]interface{}{
			"kind":     "metric",
			"category": "host",
			"type":     "info",
			"module":   "system",
		},
		"system": map[string]interface{}{},
	}

	systemData := event["system"].(map[string]interface{})

	// CRITICAL: Real uptime from system (seconds)
	if hostInfo.Uptime > 0 {
		systemData["uptime"] = hostInfo.Uptime
	} else {
		s.logger.Warn("Uptime unavailable")
		systemData["uptime"] = "unavailable"
	}

	// CRITICAL: Real boot time (unix timestamp)
	if hostInfo.BootTime > 0 {
		systemData["boot_time"] = hostInfo.BootTime
	} else {
		s.logger.Warn("Boot time unavailable")
		systemData["boot_time"] = "unavailable"
	}

	// Collect all metrics with error handling
	if cpuData := s.getCPUMetrics(); cpuData != nil {
		systemData["cpu"] = cpuData
	}

	if memData := s.getMemoryMetrics(); memData != nil {
		systemData["memory"] = memData
	}

	// Only collect root filesystem metrics
	if diskData := s.getDiskMetrics(); diskData != nil {
		systemData["filesystem"] = diskData
	}

	if netData := s.getNetworkMetrics(); netData != nil {
		systemData["network"] = netData
	}

	if loadData := s.getLoadMetrics(); loadData != nil {
		systemData["load"] = loadData
	}

	return s.pipeline.Send(event)
}

// getCPUMetrics returns CPU usage data
func (s *SystemModule) getCPUMetrics() map[string]interface{} {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil || len(cpuPercent) == 0 {
		return nil
	}
	return map[string]interface{}{
		"total": map[string]interface{}{
			"pct": cpuPercent[0] / 100.0,
		},
	}
}

// getMemoryMetrics returns memory usage data
func (s *SystemModule) getMemoryMetrics() map[string]interface{} {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil
	}
	return map[string]interface{}{
		"total": memInfo.Total,
		"used": map[string]interface{}{
			"bytes": memInfo.Used,
			"pct":   memInfo.UsedPercent / 100.0,
		},
		"free":      memInfo.Free,
		"available": memInfo.Available,
	}
}

// getDiskMetrics returns root filesystem usage data only
func (s *SystemModule) getDiskMetrics() map[string]interface{} {
	usage, err := disk.Usage("/")
	if err != nil {
		s.logger.Error("Failed to get root filesystem usage", "error", err)
		return nil
	}
	
	// Only return root filesystem data
	return map[string]interface{}{
		"device_name": "/dev/root",
		"mount_point": "/",
		"total":       usage.Total,
		"used": map[string]interface{}{
			"bytes": usage.Used,
			"pct":   usage.UsedPercent / 100.0,
		},
		"free": usage.Free,
	}
}



// getNetworkMetrics returns primary network interface data
func (s *SystemModule) getNetworkMetrics() map[string]interface{} {
	interfaces, err := psnet.IOCounters(true)
	if err != nil {
		return nil
	}

	// Find primary interface (not loopback, has traffic)
	for _, iface := range interfaces {
		if iface.Name != "lo" && iface.Name != "lo0" && (iface.BytesRecv > 0 || iface.BytesSent > 0) {
			return map[string]interface{}{
				"name": iface.Name,
				"in": map[string]interface{}{
					"bytes": iface.BytesRecv,
				},
				"out": map[string]interface{}{
					"bytes": iface.BytesSent,
				},
			}
		}
	}
	return nil
}

// getLoadMetrics returns system load data
func (s *SystemModule) getLoadMetrics() map[string]interface{} {
	loadAvg, err := load.Avg()
	if err != nil {
		return nil
	}
	return map[string]interface{}{
		"1":  loadAvg.Load1,
		"5":  loadAvg.Load5,
		"15": loadAvg.Load15,
	}
}

// getAllNetworkIPs returns all network interface IP addresses
func (s *SystemModule) getAllNetworkIPs() []string {
	var ips []string
	
	// Get primary IP via route to external
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		primaryIP := strings.Split(localAddr.IP.String(), ":")[0]
		ips = append(ips, primaryIP)
		conn.Close()
	}

	// Get all interface IPs using gopsutil
	interfaces, err := psnet.Interfaces()
	if err != nil {
		s.logger.Error("Failed to get network interfaces", "error", err)
		return ips
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if strings.Contains(iface.Name, "lo") || len(iface.Addrs) == 0 {
			continue
		}

		for _, addr := range iface.Addrs {
			// Parse IP address
			if ipAddr := net.ParseIP(addr.Addr); ipAddr != nil {
				// Skip IPv6 for simplicity, add IPv4 only
				if ipAddr.To4() != nil {
					ipStr := ipAddr.String()
					// Avoid duplicates
					found := false
					for _, existing := range ips {
						if existing == ipStr {
							found = true
							break
						}
					}
					if !found {
						ips = append(ips, ipStr)
					}
				}
			}
		}
	}

	return ips
}

