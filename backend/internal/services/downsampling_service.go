package services

import (
	"fmt"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

type DownsamplingService struct {
	db interface{}
}

func NewDownsamplingService() *DownsamplingService {
	return &DownsamplingService{
		db: postgres.DB,
	}
}

// CreateDownsampledData creates pre-aggregated data for faster queries
func (d *DownsamplingService) CreateDownsampledData() error {
	intervals := []struct {
		name      string
		duration  string
		table     string
		truncUnit string
	}{
		{"5m", "5 minutes", "metrics_5m", "minute"},
		{"1h", "1 hour", "metrics_1h", "hour"},
		{"1d", "1 day", "metrics_1d", "day"},
	}

	for _, interval := range intervals {
		if err := d.createDownsampledTable(interval.table); err != nil {
			return err
		}

		if err := d.downsampleData(interval.duration, interval.table, interval.truncUnit); err != nil {
			return err
		}
	}

	return nil
}

// createDownsampledTable creates table for downsampled data
func (d *DownsamplingService) createDownsampledTable(tableName string) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			host_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			name VARCHAR(128) NOT NULL,
			value DECIMAL(10,4) NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			UNIQUE(host_id, name, timestamp)
		)
	`, tableName)

	return postgres.DB.Exec(query).Error
}

// downsampleData aggregates raw data into downsampled tables
func (d *DownsamplingService) downsampleData(duration, tableName, truncUnit string) error {
	// truncUnit must be a valid date_trunc unit like 'minute', 'hour', 'day'
	query := fmt.Sprintf(`
		INSERT INTO %s (host_id, tenant_id, name, value, timestamp)
		SELECT 
			host_id,
			tenant_id,
			name,
			AVG(value) as avg_value,
			date_trunc('%s', timestamp) as bucket
		FROM metrics 
		WHERE timestamp >= NOW() - INTERVAL '%s'
		GROUP BY host_id, tenant_id, name, bucket
		ON CONFLICT (host_id, name, timestamp) DO UPDATE SET
			value = EXCLUDED.value
	`, tableName, truncUnit, duration)

	return postgres.DB.Exec(query).Error
}

// StartDownsamplingWorker runs periodic downsampling
func (d *DownsamplingService) StartDownsamplingWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			if err := d.CreateDownsampledData(); err != nil {
				fmt.Printf("[ERROR] Downsampling failed: %v\n", err)
			}
		}
	}()
}
