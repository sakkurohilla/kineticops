package workers

import (
	"log"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
	"github.com/spf13/viper"
)

// StartMetricCollector starts the background metric collection worker
func StartMetricCollector() {
	c := cron.New()

	// Collect metrics every 30 seconds for better real-time data
	c.AddFunc("@every 30s", func() {
		collectMetricsForAllHosts()
	})

	c.Start()
	log.Println("[WORKER] Metric collector started (runs every 30s)")
}

func collectMetricsForAllHosts() {
	var hosts []models.Host

	// Get all active hosts with SSH configured (password or key)
	result := postgres.DB.Where("ssh_user != '' AND (ssh_password != '' OR ssh_key != '')").Find(&hosts)
	if result.Error != nil {
		log.Printf("[ERROR] Failed to fetch hosts: %v", result.Error)
		return
	}

	log.Printf("[COLLECTOR] Processing %d hosts", len(hosts))

	// Concurrency limiting to avoid overwhelming the controller
	max := viper.GetInt("MAX_CONCURRENT_COLLECTIONS")
	if max <= 0 {
		max = 10
	}
	sem := make(chan struct{}, max)
	var wg sync.WaitGroup

	for _, h := range hosts {
		wg.Add(1)
		sem <- struct{}{}
		// capture value
		go func(host models.Host) {
			defer wg.Done()
			defer func() { <-sem }()
			collectMetricsForHost(&host)
		}(h)
	}

	// Wait for all collections to finish before returning
	wg.Wait()
}

func collectMetricsForHost(host *models.Host) {
	log.Printf("[COLLECTOR] Collecting metrics for host %s (%s)", host.Hostname, host.IP)

	// Collect metrics
	metrics, err := services.CollectHostMetrics(host)
	if err != nil {
		log.Printf("[ERROR] Failed to collect metrics for host %d: %v", host.ID, err)
		services.UpdateHostStatus(host.ID, "offline")
		return
	}

	// Save metrics
	if err := services.SaveHostMetrics(metrics); err != nil {
		log.Printf("[ERROR] Failed to save metrics for host %d: %v", host.ID, err)
		return
	}

	// Update host status to online
	services.UpdateHostStatus(host.ID, "online")

	log.Printf("[SUCCESS] Metrics collected for host %s", host.Hostname)
}
