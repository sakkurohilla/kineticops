package metrics

import (
	"context"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

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

// collectMetrics gathers all system metrics
func (s *SystemModule) collectMetrics() error {
	timestamp := time.Now().UTC()
	hostname, _ := host.Info()
	s.logger.Info("Starting metrics collection cycle", "timestamp", timestamp.Format(time.RFC3339))

	// Collect CPU metrics
	if s.config.CPU.Enabled {
		if err := s.collectCPUMetrics(timestamp, hostname.Hostname); err != nil {
			s.logger.Error("Failed to collect CPU metrics", "error", err)
		}
	}

	// Collect memory metrics
	if s.config.Memory.Enabled {
		if err := s.collectMemoryMetrics(timestamp, hostname.Hostname); err != nil {
			s.logger.Error("Failed to collect memory metrics", "error", err)
		}
	}

	// Collect filesystem metrics
	if s.config.Filesystem.Enabled {
		if err := s.collectFilesystemMetrics(timestamp, hostname.Hostname); err != nil {
			s.logger.Error("Failed to collect filesystem metrics", "error", err)
		}
	}

	// Collect network metrics
	if s.config.Network.Enabled {
		if err := s.collectNetworkMetrics(timestamp, hostname.Hostname); err != nil {
			s.logger.Error("Failed to collect network metrics", "error", err)
		}
	}

	// Collect load metrics
	if err := s.collectLoadMetrics(timestamp, hostname.Hostname); err != nil {
		s.logger.Error("Failed to collect load metrics", "error", err)
	}

	s.logger.Info("Metrics collection cycle completed successfully")
	return nil
}

// collectCPUMetrics collects CPU usage metrics
func (s *SystemModule) collectCPUMetrics(timestamp time.Time, hostname string) error {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return err
	}

	if len(cpuPercent) > 0 {
		event := s.createBaseEvent(timestamp, hostname, "cpu")
		event["system"] = map[string]interface{}{
			"cpu": map[string]interface{}{
				"total": map[string]interface{}{
					"pct": cpuPercent[0] / 100.0,
				},
			},
		}
		event["metricset"] = map[string]interface{}{
			"name": "cpu",
		}

		if err := s.pipeline.Send(event); err != nil {
			return err
		}
	}

	// Per-CPU metrics if enabled
	if s.config.CPU.PerCPU {
		cpuPercentPerCPU, err := cpu.Percent(0, true)
		if err != nil {
			return err
		}

		for i, percent := range cpuPercentPerCPU {
			event := s.createBaseEvent(timestamp, hostname, "cpu")
			event["system"] = map[string]interface{}{
				"cpu": map[string]interface{}{
					"core": map[string]interface{}{
						"id":  i,
						"pct": percent / 100.0,
					},
				},
			}
			event["metricset"] = map[string]interface{}{
				"name": "cpu",
			}

			if err := s.pipeline.Send(event); err != nil {
				return err
			}
		}
	}

	return nil
}

// collectMemoryMetrics collects memory usage metrics
func (s *SystemModule) collectMemoryMetrics(timestamp time.Time, hostname string) error {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	swapInfo, err := mem.SwapMemory()
	if err != nil {
		return err
	}

	event := s.createBaseEvent(timestamp, hostname, "memory")
	event["system"] = map[string]interface{}{
		"memory": map[string]interface{}{
			"total": memInfo.Total,
			"used": map[string]interface{}{
				"bytes": memInfo.Used,
				"pct":   memInfo.UsedPercent / 100.0,
			},
			"free":      memInfo.Free,
			"available": memInfo.Available,
		},
	}
	event["metricset"] = map[string]interface{}{
		"name": "memory",
	}

	if swapInfo.Total > 0 {
		event["system"].(map[string]interface{})["swap"] = map[string]interface{}{
			"total": swapInfo.Total,
			"used": map[string]interface{}{
				"bytes": swapInfo.Used,
				"pct":   swapInfo.UsedPercent / 100.0,
			},
			"free": swapInfo.Free,
		}
	}

	return s.pipeline.Send(event)
}

// collectFilesystemMetrics collects filesystem usage metrics
func (s *SystemModule) collectFilesystemMetrics(timestamp time.Time, hostname string) error {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return err
	}

	for _, partition := range partitions {
		// Skip irrelevant filesystems
		if s.shouldSkipFilesystem(partition) {
			continue
		}

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		// Skip very small filesystems (< 100MB)
		if usage.Total < 100*1024*1024 {
			continue
		}

		event := s.createBaseEvent(timestamp, hostname, "filesystem")
		event["system"] = map[string]interface{}{
			"filesystem": map[string]interface{}{
				"device_name": partition.Device,
				"mount_point": partition.Mountpoint,
				"type":        partition.Fstype,
				"total":       usage.Total,
				"used": map[string]interface{}{
					"bytes": usage.Used,
					"pct":   usage.UsedPercent / 100.0,
				},
				"free":      usage.Free,
				"available": usage.Free,
			},
		}
		event["metricset"] = map[string]interface{}{
			"name": "filesystem",
		}

		if err := s.pipeline.Send(event); err != nil {
			return err
		}
	}

	return nil
}

// shouldSkipFilesystem determines if a filesystem should be skipped
func (s *SystemModule) shouldSkipFilesystem(partition disk.PartitionStat) bool {
	// Skip CD-ROM, DVD, and other optical drives
	if partition.Device == "/dev/sr0" || partition.Device == "/dev/cdrom" {
		return true
	}
	
	// Skip tmpfs, devtmpfs, and other virtual filesystems
	if partition.Fstype == "tmpfs" || partition.Fstype == "devtmpfs" || 
	   partition.Fstype == "sysfs" || partition.Fstype == "proc" ||
	   partition.Fstype == "devpts" || partition.Fstype == "cgroup" ||
	   partition.Fstype == "pstore" || partition.Fstype == "efivarfs" {
		return true
	}
	
	// Skip snap mounts
	if len(partition.Mountpoint) > 5 && partition.Mountpoint[:5] == "/snap" {
		return true
	}
	
	// Skip loop devices
	if len(partition.Device) > 9 && partition.Device[:9] == "/dev/loop" {
		return true
	}
	
	return false
}

// collectNetworkMetrics collects network interface metrics
func (s *SystemModule) collectNetworkMetrics(timestamp time.Time, hostname string) error {
	interfaces, err := net.IOCounters(true)
	if err != nil {
		return err
	}

	for _, iface := range interfaces {
		// Skip loopback interfaces
		if iface.Name == "lo" || iface.Name == "lo0" {
			continue
		}

		event := s.createBaseEvent(timestamp, hostname, "network")
		event["system"] = map[string]interface{}{
			"network": map[string]interface{}{
				"name": iface.Name,
				"in": map[string]interface{}{
					"bytes":   iface.BytesRecv,
					"packets": iface.PacketsRecv,
					"errors":  iface.Errin,
					"dropped": iface.Dropin,
				},
				"out": map[string]interface{}{
					"bytes":   iface.BytesSent,
					"packets": iface.PacketsSent,
					"errors":  iface.Errout,
					"dropped": iface.Dropout,
				},
			},
		}
		event["metricset"] = map[string]interface{}{
			"name": "network",
		}

		if err := s.pipeline.Send(event); err != nil {
			return err
		}
	}

	return nil
}

// collectLoadMetrics collects system load metrics
func (s *SystemModule) collectLoadMetrics(timestamp time.Time, hostname string) error {
	loadAvg, err := load.Avg()
	if err != nil {
		return err
	}

	event := s.createBaseEvent(timestamp, hostname, "load")
	event["system"] = map[string]interface{}{
		"load": map[string]interface{}{
			"1":  loadAvg.Load1,
			"5":  loadAvg.Load5,
			"15": loadAvg.Load15,
			"norm": map[string]interface{}{
				"1":  loadAvg.Load1 / float64(runtime.NumCPU()),
				"5":  loadAvg.Load5 / float64(runtime.NumCPU()),
				"15": loadAvg.Load15 / float64(runtime.NumCPU()),
			},
		},
	}
	event["metricset"] = map[string]interface{}{
		"name": "load",
	}

	return s.pipeline.Send(event)
}

// createBaseEvent creates a base event structure
func (s *SystemModule) createBaseEvent(timestamp time.Time, hostname, metricset string) map[string]interface{} {
	return map[string]interface{}{
		"@timestamp": timestamp.Format(time.RFC3339),
		"agent": map[string]interface{}{
			"name":    "kineticops-agent",
			"type":    "metricbeat",
			"version": "1.0.0",
		},
		"host": map[string]interface{}{
			"hostname": hostname,
		},
		"event": map[string]interface{}{
			"kind":     "metric",
			"category": "host",
			"type":     "info",
			"module":   "system",
		},
		"service": map[string]interface{}{
			"type": "system",
		},
	}
}