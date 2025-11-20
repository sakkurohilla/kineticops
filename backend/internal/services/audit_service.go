package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// AuditService handles audit logging
type AuditService struct{}

func NewAuditService() *AuditService {
	return &AuditService{}
}

// LogAction logs an audit event
func (s *AuditService) LogAction(ctx context.Context, userID int64, username, action, resource, resourceID string, status string, details interface{}, err error) {
	auditLog := &models.AuditLog{
		TenantID:   1, // Default tenant
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Status:     status,
	}

	// Extract IP and User-Agent from context if available
	if c, ok := ctx.Value("fiber").(*fiber.Ctx); ok {
		auditLog.IPAddress = c.IP()
		auditLog.UserAgent = c.Get("User-Agent")
	}

	// Add error message if failed
	if err != nil {
		auditLog.ErrorMessage = err.Error()
	}

	// Add details as JSON
	if details != nil {
		if detailsJSON, jerr := json.Marshal(details); jerr == nil {
			auditLog.Details = string(detailsJSON)
		}
	}

	// Store in database
	if dberr := postgres.DB.Create(auditLog).Error; dberr != nil {
		logging.Errorf("Failed to create audit log: %v", dberr)
	}
}

// LogFromFiber is a convenience method for logging from Fiber handlers
func (s *AuditService) LogFromFiber(c *fiber.Ctx, userID int64, username, action, resource, resourceID string, status string, details interface{}, err error) {
	auditLog := &models.AuditLog{
		TenantID:   1,
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  c.IP(),
		UserAgent:  c.Get("User-Agent"),
		Status:     status,
	}

	if err != nil {
		auditLog.ErrorMessage = err.Error()
	}

	if details != nil {
		if detailsJSON, jerr := json.Marshal(details); jerr == nil {
			auditLog.Details = string(detailsJSON)
		}
	}

	if dberr := postgres.DB.Create(auditLog).Error; dberr != nil {
		logging.Errorf("Failed to create audit log: %v", dberr)
	}
}

// GetAuditLogs retrieves audit logs with filters
func (s *AuditService) GetAuditLogs(tenantID, userID int64, action, resource string, startTime, endTime time.Time, limit, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := postgres.DB.Model(&models.AuditLog{})

	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error

	return logs, total, err
}

// Global audit service instance
var AuditSvc = NewAuditService()
