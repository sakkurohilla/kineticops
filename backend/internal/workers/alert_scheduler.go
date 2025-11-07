package workers

import (
	"fmt"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

// StartAlertScheduler periodically evaluates alert rules and triggers alerts when conditions are met.
func StartAlertScheduler() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			// For simplicity, evaluate rules for tenant 0 (all tenants) and per-tenant in DB listing.
			rules, err := postgres.ListAlertRules(postgres.DB, 0)
			if err != nil {
				fmt.Printf("[ALERT SCHEDULER] failed to list rules: %v\n", err)
				continue
			}
			for _, r := range rules {
				// For each rule, evaluate per-host using AggregationService
				agg := services.NewAggregationService()
				// Build query over the rule window (minutes)
				end := time.Now()
				start := end.Add(-time.Duration(r.Window) * time.Minute)
				query := services.AggregationQuery{
					MetricName: r.MetricName,
					StartTime:  start,
					EndTime:    end,
					Function:   "avg",
					Interval:   "1m",
				}

				// For each host in tenant scope, fetch host list
				hosts, _ := services.ListHosts(r.TenantID, 1000, 0)
				for _, h := range hosts {
					// reuse query per host
					query.HostID = int64(h.ID)
					pts, err := agg.QueryTimeSeries(query)
					if err != nil || len(pts) == 0 {
						continue
					}
					// take last value
					last := pts[len(pts)-1]
					// evaluate condition
					_ = services.CheckAndTriggerAlerts(r.TenantID, r.MetricName, int64(h.ID), last.Value)
				}
			}
		}
	}()
	fmt.Println("[ALERT SCHEDULER] started (30s interval)")
}
