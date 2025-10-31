package postgres

import (
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SaveHostMetric inserts a HostMetric into the database.
// Uses ON CONFLICT DO NOTHING to tolerate duplicate timestamps for the same host.
func SaveHostMetric(db *gorm.DB, m *models.HostMetric) error {
	// Use upsert-ignore on conflict to avoid duplicate timestamp insertion errors.
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(m).Error
}

// GetLatestHostMetric returns the most recent HostMetric for a host.
// Returns (nil, nil) when no record exists.
func GetLatestHostMetric(db *gorm.DB, hostID int64) (*models.HostMetric, error) {
	var hm models.HostMetric
	err := db.Where("host_id = ?", hostID).Order("timestamp DESC").First(&hm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &hm, nil
}

// GetHostMetricsHistory returns the last `limit` metrics for a host in chronological order (oldest first).
func GetHostMetricsHistory(db *gorm.DB, hostID int64, limit int) ([]models.HostMetric, error) {
	var rows []models.HostMetric
	q := db.Where("host_id = ?", hostID)
	if limit > 0 {
		q = q.Limit(limit)
	}
	// Fetch newest first then reverse to return oldest-first for charts.
	err := q.Order("timestamp DESC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	// reverse slice to chronological order
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	return rows, nil
}

// GetHostMetricsByTimeRange returns metrics for a host between start and end (inclusive).
// Results are returned in chronological order (oldest first).
func GetHostMetricsByTimeRange(db *gorm.DB, hostID int64, start, end time.Time) ([]models.HostMetric, error) {
	var rows []models.HostMetric
	err := db.Where("host_id = ? AND timestamp BETWEEN ? AND ?", hostID, start, end).
		Order("timestamp ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// CleanupOldMetrics deletes metrics older than cutoff.
func CleanupOldMetrics(db *gorm.DB, cutoff time.Time) error {
	return db.Where("timestamp < ?", cutoff).Delete(&models.HostMetric{}).Error
}

// EnsureHostMetricIndexes creates helpful indexes for host_metrics table.
func EnsureHostMetricIndexes(db *gorm.DB) error {
	// Create single-column indexes if missing.
	if err := db.Migrator().CreateIndex(&models.HostMetric{}, "HostID"); err != nil {
		return err
	}
	if err := db.Migrator().CreateIndex(&models.HostMetric{}, "Timestamp"); err != nil {
		return err
	}
	// Create composite index on host_id + timestamp for queries that filter by host and order by time.
	// GORM will create an index name automatically.
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_host_metrics_host_id_timestamp ON host_metrics (host_id, timestamp DESC)").Error; err != nil {
		return err
	}
	return nil
}
