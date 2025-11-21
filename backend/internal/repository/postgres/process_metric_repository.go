package postgres

import (
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"gorm.io/gorm"
)

// GetProcessMetricsByHost returns top processes for a host sorted by CPU or memory
func GetProcessMetricsByHost(db *gorm.DB, hostID int64, sortBy string, limit int) ([]models.ProcessMetric, error) {
	var processes []models.ProcessMetric

	orderClause := "cpu_percent DESC"
	if sortBy == "memory" {
		orderClause = "memory_percent DESC"
	}

	// Get processes from the last 2 minutes (most recent data)
	cutoff := time.Now().Add(-2 * time.Minute)

	err := db.Where("host_id = ? AND timestamp > ?", hostID, cutoff).
		Order(orderClause).
		Limit(limit).
		Find(&processes).Error

	return processes, err
}

// GetTopProcessesAcrossHosts returns top processes across all hosts for a tenant
func GetTopProcessesAcrossHosts(db *gorm.DB, tenantID int64, limit int) ([]models.ProcessMetric, error) {
	var processes []models.ProcessMetric

	cutoff := time.Now().Add(-2 * time.Minute)
	query := db.Where("timestamp > ?", cutoff).Order("cpu_percent DESC").Limit(limit)

	// If tenantID is provided, filter by tenant
	if tenantID > 0 {
		// Join with hosts table to filter by tenant
		query = query.Joins("JOIN hosts ON process_metrics.host_id = hosts.id").
			Where("hosts.tenant_id = ?", tenantID)
	}

	err := query.Find(&processes).Error
	return processes, err
}
