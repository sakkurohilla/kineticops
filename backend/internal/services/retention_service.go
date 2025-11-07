package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/repository/mongodb"
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
	// If TimescaleDB is present, prefer dropping chunks (efficient) otherwise fall back to DELETE.
	// Try drop_chunks function; if it fails (extension missing), use DELETE as a fallback.
	dropSQL := fmt.Sprintf("SELECT drop_chunks(INTERVAL '%d days', 'metrics')", r.retentionDays)
	if err := postgres.DB.Exec(dropSQL).Error; err != nil {
		// Fallback to plain delete
		result := postgres.DB.Exec("DELETE FROM metrics WHERE timestamp < ?", cutoffTime)
		if result.Error != nil {
			return fmt.Errorf("failed to cleanup old metrics: %w", result.Error)
		}
		fmt.Printf("[RETENTION] Fallback cleaned up %d old metric records\n", result.RowsAffected)
		return nil
	}

	fmt.Printf("[RETENTION] Dropped metrics older than %d days using Timescale drop_chunks\n", r.retentionDays)
	return nil
}

// CleanupOldLogs removes logs older than retention period
func (r *RetentionService) CleanupOldLogs() error {
	cutoffTime := time.Now().AddDate(0, 0, -r.retentionDays)
	// delegate to mongodb repository
	return mongodb.DeleteOldLogs(context.Background(), cutoffTime)
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
