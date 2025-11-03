package services

import (
	"fmt"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"gorm.io/gorm"
)

type APMService struct {
	db *gorm.DB
}

func NewAPMService() *APMService {
	return &APMService{
		db: postgres.DB,
	}
}

// Application Management
func (s *APMService) CreateApplication(app *models.Application) error {
	return postgres.DB.Create(app).Error
}

func (s *APMService) GetApplications(tenantID int64) ([]models.Application, error) {
	var apps []models.Application
	err := postgres.DB.Where("tenant_id = ?", tenantID).Find(&apps).Error
	return apps, err
}

func (s *APMService) GetApplication(id int64, tenantID int64) (*models.Application, error) {
	var app models.Application
	err := postgres.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&app).Error
	return &app, err
}

// Transaction Tracking
func (s *APMService) RecordTransaction(tx *models.Transaction) error {
	return postgres.DB.Create(tx).Error
}

func (s *APMService) GetTransactions(appID int64, tenantID int64, start, end time.Time, limit int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	query := postgres.DB.Where("application_id = ? AND tenant_id = ?", appID, tenantID)
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("timestamp BETWEEN ? AND ?", start, end)
	}
	
	err := query.Order("timestamp DESC").Limit(limit).Find(&transactions).Error
	return transactions, err
}

func (s *APMService) GetTransactionStats(appID int64, tenantID int64, start, end time.Time) (map[string]interface{}, error) {
	var stats struct {
		AvgResponseTime float64 `json:"avg_response_time"`
		Throughput      float64 `json:"throughput"`
		ErrorRate       float64 `json:"error_rate"`
		Apdex           float64 `json:"apdex"`
		TotalRequests   int64   `json:"total_requests"`
		ErrorCount      int64   `json:"error_count"`
	}

	query := `
		SELECT 
			AVG(response_time) as avg_response_time,
			COUNT(*) as total_requests,
			SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as error_count,
			AVG(apdex) as apdex
		FROM transactions 
		WHERE application_id = ? AND tenant_id = ? AND timestamp BETWEEN ? AND ?
	`
	
	err := postgres.DB.Raw(query, appID, tenantID, start, end).Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	// Calculate derived metrics
	if stats.TotalRequests > 0 {
		stats.ErrorRate = float64(stats.ErrorCount) / float64(stats.TotalRequests) * 100
		duration := end.Sub(start).Hours()
		if duration > 0 {
			stats.Throughput = float64(stats.TotalRequests) / duration
		}
	}

	result := map[string]interface{}{
		"avg_response_time": stats.AvgResponseTime,
		"throughput":        stats.Throughput,
		"error_rate":        stats.ErrorRate,
		"apdex":            stats.Apdex,
		"total_requests":   stats.TotalRequests,
		"error_count":      stats.ErrorCount,
	}

	return result, nil
}

// Distributed Tracing
func (s *APMService) CreateTrace(trace *models.Trace) error {
	return postgres.DB.Create(trace).Error
}

func (s *APMService) AddSpan(span *models.Span) error {
	return postgres.DB.Create(span).Error
}

func (s *APMService) GetTrace(traceID string, tenantID int64) (*models.Trace, []models.Span, error) {
	var trace models.Trace
	err := postgres.DB.Where("trace_id = ? AND tenant_id = ?", traceID, tenantID).First(&trace).Error
	if err != nil {
		return nil, nil, err
	}

	var spans []models.Span
	err = postgres.DB.Where("trace_id = ? AND tenant_id = ?", traceID, tenantID).
		Order("start_time ASC").Find(&spans).Error
	
	return &trace, spans, err
}

func (s *APMService) GetTraces(appID int64, tenantID int64, start, end time.Time, limit int) ([]models.Trace, error) {
	var traces []models.Trace
	query := postgres.DB.Where("application_id = ? AND tenant_id = ?", appID, tenantID)
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("timestamp BETWEEN ? AND ?", start, end)
	}
	
	err := query.Order("timestamp DESC").Limit(limit).Find(&traces).Error
	return traces, err
}

// Error Tracking
func (s *APMService) RecordError(errorEvent *models.ErrorEvent) error {
	// Check if error with same fingerprint exists
	var existing models.ErrorEvent
	err := postgres.DB.Where("fingerprint = ? AND tenant_id = ?", 
		errorEvent.Fingerprint, errorEvent.TenantID).First(&existing).Error
	
	if err == nil {
		// Update existing error
		existing.Count++
		existing.LastSeen = errorEvent.Timestamp
		return postgres.DB.Save(&existing).Error
	}
	
	// Create new error
	errorEvent.FirstSeen = errorEvent.Timestamp
	errorEvent.LastSeen = errorEvent.Timestamp
	return postgres.DB.Create(errorEvent).Error
}

func (s *APMService) GetErrors(appID int64, tenantID int64, start, end time.Time, limit int) ([]models.ErrorEvent, error) {
	var errors []models.ErrorEvent
	query := postgres.DB.Where("application_id = ? AND tenant_id = ?", appID, tenantID)
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("timestamp BETWEEN ? AND ?", start, end)
	}
	
	err := query.Order("last_seen DESC").Limit(limit).Find(&errors).Error
	return errors, err
}

func (s *APMService) GetErrorStats(appID int64, tenantID int64, start, end time.Time) (map[string]interface{}, error) {
	var stats struct {
		TotalErrors   int64 `json:"total_errors"`
		UniqueErrors  int64 `json:"unique_errors"`
		ErrorRate     float64 `json:"error_rate"`
		ResolvedCount int64 `json:"resolved_count"`
	}

	// Get error counts
	err := postgres.DB.Model(&models.ErrorEvent{}).
		Where("application_id = ? AND tenant_id = ? AND timestamp BETWEEN ? AND ?", 
			appID, tenantID, start, end).
		Select("COUNT(*) as unique_errors, SUM(count) as total_errors, SUM(CASE WHEN resolved THEN 1 ELSE 0 END) as resolved_count").
		Scan(&stats).Error
	
	if err != nil {
		return nil, err
	}

	// Calculate error rate (errors per transaction)
	var txCount int64
	postgres.DB.Model(&models.Transaction{}).
		Where("application_id = ? AND tenant_id = ? AND timestamp BETWEEN ? AND ?", 
			appID, tenantID, start, end).Count(&txCount)
	
	if txCount > 0 {
		stats.ErrorRate = float64(stats.TotalErrors) / float64(txCount) * 100
	}

	result := map[string]interface{}{
		"total_errors":   stats.TotalErrors,
		"unique_errors":  stats.UniqueErrors,
		"error_rate":     stats.ErrorRate,
		"resolved_count": stats.ResolvedCount,
	}

	return result, nil
}

// Database Query Tracking
func (s *APMService) RecordDatabaseQuery(query *models.DatabaseQuery) error {
	return postgres.DB.Create(query).Error
}

func (s *APMService) GetDatabaseQueries(appID int64, tenantID int64, start, end time.Time, limit int) ([]models.DatabaseQuery, error) {
	var queries []models.DatabaseQuery
	query := postgres.DB.Where("application_id = ? AND tenant_id = ?", appID, tenantID)
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("timestamp BETWEEN ? AND ?", start, end)
	}
	
	err := query.Order("duration DESC").Limit(limit).Find(&queries).Error
	return queries, err
}

func (s *APMService) GetSlowQueries(appID int64, tenantID int64, threshold float64, start, end time.Time, limit int) ([]models.DatabaseQuery, error) {
	var queries []models.DatabaseQuery
	query := postgres.DB.Where("application_id = ? AND tenant_id = ? AND duration > ?", 
		appID, tenantID, threshold)
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("timestamp BETWEEN ? AND ?", start, end)
	}
	
	err := query.Order("duration DESC").Limit(limit).Find(&queries).Error
	return queries, err
}

// External Service Tracking
func (s *APMService) RecordExternalService(service *models.ExternalService) error {
	return postgres.DB.Create(service).Error
}

func (s *APMService) GetExternalServices(appID int64, tenantID int64, start, end time.Time, limit int) ([]models.ExternalService, error) {
	var services []models.ExternalService
	query := postgres.DB.Where("application_id = ? AND tenant_id = ?", appID, tenantID)
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("timestamp BETWEEN ? AND ?", start, end)
	}
	
	err := query.Order("duration DESC").Limit(limit).Find(&services).Error
	return services, err
}

// Performance Metrics
func (s *APMService) RecordPerformanceMetric(metric *models.PerformanceMetric) error {
	return postgres.DB.Create(metric).Error
}

func (s *APMService) GetPerformanceMetrics(appID int64, tenantID int64, metricName string, start, end time.Time) ([]models.PerformanceMetric, error) {
	var metrics []models.PerformanceMetric
	query := postgres.DB.Where("application_id = ? AND tenant_id = ?", appID, tenantID)
	
	if metricName != "" {
		query = query.Where("metric_name = ?", metricName)
	}
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("timestamp BETWEEN ? AND ?", start, end)
	}
	
	err := query.Order("timestamp ASC").Find(&metrics).Error
	return metrics, err
}

// Custom Events
func (s *APMService) RecordCustomEvent(event *models.CustomEvent) error {
	return postgres.DB.Create(event).Error
}

func (s *APMService) GetCustomEvents(appID int64, tenantID int64, eventType string, start, end time.Time, limit int) ([]models.CustomEvent, error) {
	var events []models.CustomEvent
	query := postgres.DB.Where("application_id = ? AND tenant_id = ?", appID, tenantID)
	
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("timestamp BETWEEN ? AND ?", start, end)
	}
	
	err := query.Order("timestamp DESC").Limit(limit).Find(&events).Error
	return events, err
}

// Dashboard Data
func (s *APMService) GetApplicationOverview(appID int64, tenantID int64, start, end time.Time) (map[string]interface{}, error) {
	overview := make(map[string]interface{})
	
	// Get transaction stats
	txStats, err := s.GetTransactionStats(appID, tenantID, start, end)
	if err != nil {
		return nil, err
	}
	overview["transactions"] = txStats
	
	// Get error stats
	errorStats, err := s.GetErrorStats(appID, tenantID, start, end)
	if err != nil {
		return nil, err
	}
	overview["errors"] = errorStats
	
	// Get top transactions
	topTx, err := s.GetTransactions(appID, tenantID, start, end, 10)
	if err != nil {
		return nil, err
	}
	overview["top_transactions"] = topTx
	
	// Get recent errors
	recentErrors, err := s.GetErrors(appID, tenantID, start, end, 10)
	if err != nil {
		return nil, err
	}
	overview["recent_errors"] = recentErrors
	
	return overview, nil
}

// Helper function to generate fingerprint for errors
func GenerateErrorFingerprint(errorClass, errorMessage string) string {
	data := fmt.Sprintf("%s:%s", errorClass, errorMessage)
	// Simple hash - in production use proper hashing
	return fmt.Sprintf("%x", []byte(data))
}