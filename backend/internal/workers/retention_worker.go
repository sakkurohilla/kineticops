package workers

import (
	"context"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	MetricsRetentionDays   = 30 // Keep metrics for 30 days
	LogsRetentionDays      = 30 // Keep logs for 30 days
	AuditLogsRetentionDays = 90 // Keep audit logs for 90 days (compliance)
	CleanupInterval        = 6 * time.Hour
)

type RetentionWorker struct {
	mongoClient *mongo.Client
}

func NewRetentionWorker(mongoClient *mongo.Client) *RetentionWorker {
	return &RetentionWorker{
		mongoClient: mongoClient,
	}
}

// Start begins the retention policy enforcement worker
func (w *RetentionWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	logging.Infof("Retention worker started (interval: %v)", CleanupInterval)

	// Run immediately on start
	w.cleanupOldData(ctx)

	for {
		select {
		case <-ctx.Done():
			logging.Infof("Retention worker stopped")
			return
		case <-ticker.C:
			w.cleanupOldData(ctx)
		}
	}
}

func (w *RetentionWorker) cleanupOldData(ctx context.Context) {
	logging.Infof("Starting retention policy cleanup")

	// Cleanup metrics
	metricsDeleted := w.cleanupMetrics(ctx)

	// Cleanup logs
	logsDeleted := w.cleanupLogs(ctx)

	// Cleanup aggregated metrics
	aggDeleted := w.cleanupAggregatedMetrics(ctx)

	// Cleanup audit logs (longer retention)
	auditDeleted := w.cleanupAuditLogs(ctx)

	logging.Infof("Retention cleanup completed: metrics=%d, logs=%d, aggregations=%d, audit_logs=%d",
		metricsDeleted, logsDeleted, aggDeleted, auditDeleted)
}

func (w *RetentionWorker) cleanupMetrics(_ context.Context) int64 {
	cutoffDate := time.Now().AddDate(0, 0, -MetricsRetentionDays)

	// Delete old metrics from PostgreSQL
	tx := postgres.DB.Exec(`
		DELETE FROM metrics 
		WHERE timestamp < $1
	`, cutoffDate)

	if tx.Error != nil {
		logging.Errorf("Failed to cleanup old metrics: %v", tx.Error)
		return 0
	}

	if tx.RowsAffected > 0 {
		logging.Infof("Deleted %d old metrics (older than %d days)", tx.RowsAffected, MetricsRetentionDays)
	}

	return tx.RowsAffected
}

func (w *RetentionWorker) cleanupLogs(ctx context.Context) int64 {
	cutoffDate := time.Now().AddDate(0, 0, -LogsRetentionDays)

	// Delete old logs from MongoDB
	collection := w.mongoClient.Database("kineticops").Collection("logs")

	filter := bson.M{
		"timestamp": bson.M{
			"$lt": cutoffDate,
		},
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		logging.Errorf("Failed to cleanup old logs: %v", err)
		return 0
	}

	if result.DeletedCount > 0 {
		logging.Infof("Deleted %d old logs (older than %d days)", result.DeletedCount, LogsRetentionDays)
	}

	return result.DeletedCount
}

func (w *RetentionWorker) cleanupAggregatedMetrics(_ context.Context) int64 {
	cutoffDate := time.Now().AddDate(0, 0, -MetricsRetentionDays)

	// Delete old aggregated metrics
	tx := postgres.DB.Exec(`
		DELETE FROM metric_aggregations 
		WHERE interval_time < $1
	`, cutoffDate)

	if tx.Error != nil {
		logging.Errorf("Failed to cleanup old aggregated metrics: %v", tx.Error)
		return 0
	}

	if tx.RowsAffected > 0 {
		logging.Infof("Deleted %d old aggregated metrics (older than %d days)", tx.RowsAffected, MetricsRetentionDays)
	}

	return tx.RowsAffected
}

func (w *RetentionWorker) cleanupAuditLogs(_ context.Context) int64 {
	cutoffDate := time.Now().AddDate(0, 0, -AuditLogsRetentionDays)

	// Delete old audit logs (longer retention for compliance)
	tx := postgres.DB.Exec(`
		DELETE FROM audit_logs 
		WHERE created_at < $1
	`, cutoffDate)

	if tx.Error != nil {
		logging.Errorf("Failed to cleanup old audit logs: %v", tx.Error)
		return 0
	}

	if tx.RowsAffected > 0 {
		logging.Infof("Deleted %d old audit logs (older than %d days)", tx.RowsAffected, AuditLogsRetentionDays)
	}

	return tx.RowsAffected
}

// CleanupProcessMetrics removes old process metrics
func (w *RetentionWorker) CleanupProcessMetrics(ctx context.Context) int64 {
	cutoffDate := time.Now().AddDate(0, 0, -7) // Keep process metrics for 7 days only

	tx := postgres.DB.Exec(`
		DELETE FROM process_metrics 
		WHERE timestamp < $1
	`, cutoffDate)

	if tx.Error != nil {
		logging.Errorf("Failed to cleanup old process metrics: %v", tx.Error)
		return 0
	}

	if tx.RowsAffected > 0 {
		logging.Infof("Deleted %d old process metrics (older than 7 days)", tx.RowsAffected)
	}

	return tx.RowsAffected
}
