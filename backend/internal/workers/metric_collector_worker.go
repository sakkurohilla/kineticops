package workers

import (
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
)

// StartMetricCollector starts the background metric collection worker
func StartMetricCollector() {
	// DISABLED: Metric collector was auto-creating hosts
	// Only collect metrics for hosts with active agents
	logging.Infof("[WORKER] Metric collector disabled - only agent heartbeats will update metrics")
}
