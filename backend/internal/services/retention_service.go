package services

import (
	"fmt"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// RetentionService handles data cleanup for enterprise scale
type RetentionService struct {
	retentionDays int
}

// NewRetentionService creates a new retention service
func NewRetentionService(retentionDays int) *RetentionService {
	if retentionDays <= 0 {
		retentionDays = 30 // Default 30 days
	}
	return &RetentionService{
		retentionDays: retentionDays,
	}
}

// CleanupOldMetrics removes metrics older than retention period
func (r *RetentionService) CleanupOldMetrics() error {
	cutoffTime := time.Now().AddDate(0, 0, -r.retentionDays)
	
	// Clean up old metrics from metrics table
	result := postgres.DB.Exec("DELETE FROM metrics WHERE timestamp < ?", cutoffTime)
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old metrics: %w", result.Error)
	}
	
	fmt.Printf("[RETENTION] Cleaned up %d old metric records\n", result.RowsAffected)
	return nil
}

// CleanupOldLogs removes logs older than retention period
func (r *RetentionService) CleanupOldLogs() error {
	// Logs are stored in MongoDB, skip PostgreSQL cleanup
	return nil
}

// StartRetentionWorker runs cleanup every 24 hours
func (r *RetentionService) StartRetentionWorker() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			if err := r.CleanupOldMetrics(); err != nil {
				fmt.Printf("[RETENTION] Error cleaning metrics: %v\n", err)
			}
			if err := r.CleanupOldLogs(); err != nil {
				fmt.Printf("[RETENTION] Error cleaning logs: %v\n", err)
			}
		}
	}()
	fmt.Println("[RETENTION] Started retention worker (24h cleanup cycle)")
}