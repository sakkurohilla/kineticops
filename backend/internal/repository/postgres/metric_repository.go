package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"gorm.io/gorm"
)

func SaveMetric(db *gorm.DB, m *models.Metric) error {
	return db.Create(m).Error
}

// SaveMetricsBatch performs a multi-row insert for metrics to improve ingestion throughput.
// It uses a single prepared INSERT statement with multiple VALUES rows. For very high
// throughput deployments consider using COPY FROM STDIN or a dedicated ingestion service.
func SaveMetricsBatch(db *gorm.DB, metrics []*models.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// Build a multi-row insert using DB.Exec. We rely on GORM's underlying DB for execution.
	// Keep SQL simple to avoid GORM model hooks; use parameter placeholders.
	sql := "INSERT INTO metrics (host_id, tenant_id, name, value, labels, timestamp) VALUES "
	params := []interface{}{}
	placeholders := []string{}

	for i, m := range metrics {
		idx := i * 6
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d)", idx+1, idx+2, idx+3, idx+4, idx+5, idx+6))
		params = append(params, m.HostID, m.TenantID, m.Name, m.Value, m.Labels, m.Timestamp)
	}

	sql = sql + strings.Join(placeholders, ",")

	// Execute raw SQL using the GORM DB's raw connection
	return db.Exec(sql, params...).Error
}

func ListMetrics(db *gorm.DB, tenantID, hostID int64, name string, start, end time.Time, limit int) ([]models.Metric, error) {
	var m []models.Metric
	q := db.Model(&models.Metric{})

	// Tenant isolation
	if tenantID > 0 {
		q = q.Where("tenant_id = ?", tenantID)
	}

	// Optional filters
	if hostID > 0 {
		q = q.Where("host_id = ?", hostID)
	}
	if name != "" {
		q = q.Where("name = ?", name)
	}

	q = q.Where("timestamp BETWEEN ? AND ?", start, end)

	if limit > 0 {
		q = q.Limit(limit)
	}

	err := q.Order("timestamp DESC").Find(&m).Error
	return m, err
}

// Get latest metric for a host/name
func LatestMetric(db *gorm.DB, hostID int64, name string) (*models.Metric, error) {
	var m models.Metric
	err := db.Where("host_id = ? AND name = ?", hostID, name).
		Order("timestamp DESC").First(&m).Error
	return &m, err
}

// Retention policy enforcement (delete old records)
func DeleteOldMetrics(db *gorm.DB, cutoff time.Time) error {
	return db.Where("timestamp < ?", cutoff).Delete(&models.Metric{}).Error
}
