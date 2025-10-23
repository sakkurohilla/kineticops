package postgres

import (
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"gorm.io/gorm"
)

func SaveMetric(db *gorm.DB, m *models.Metric) error {
	return db.Create(m).Error
}

func ListMetrics(db *gorm.DB, hostID int64, name string, start, end time.Time, limit int) ([]models.Metric, error) {
	var m []models.Metric
	q := db.Where("host_id = ? AND name = ? AND timestamp BETWEEN ? AND ?", hostID, name, start, end)
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
