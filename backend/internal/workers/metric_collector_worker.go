package workers

import (
	"log"
)

// StartMetricCollector starts the background metric collection worker
func StartMetricCollector() {
	// DISABLED: Metric collector was auto-creating hosts
	// Only collect metrics for hosts with active agents
	log.Println("[WORKER] Metric collector disabled - only agent heartbeats will update metrics")
}
