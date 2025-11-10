package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"gorm.io/gorm"
)

type EnhancedAlertService struct {
	db *gorm.DB
}

func NewEnhancedAlertService() *EnhancedAlertService {
	return &EnhancedAlertService{
		db: postgres.DB,
	}
}

// Advanced Alert Rules with NRQL-like queries
type AlertCondition struct {
	ID          int64     `json:"id"`
	TenantID    int64     `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Query       string    `json:"query"` // NRQL-like query
	Threshold   float64   `json:"threshold"`
	Operator    string    `json:"operator"` // above, below, equals
	Duration    int       `json:"duration"` // minutes
	Severity    string    `json:"severity"` // critical, high, medium, low
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

// Alert Channels (like New Relic notification channels)
type AlertChannel struct {
	ID        int64     `json:"id"`
	TenantID  int64     `json:"tenant_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`   // email, slack, webhook, pagerduty
	Config    string    `json:"config"` // JSON configuration
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// Alert Policies (grouping conditions and channels)
type AlertPolicy struct {
	ID          int64     `json:"id"`
	TenantID    int64     `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

// Alert Incidents (like New Relic incidents)
type AlertIncident struct {
	ID          int64      `json:"id"`
	TenantID    int64      `json:"tenant_id"`
	PolicyID    int64      `json:"policy_id"`
	ConditionID int64      `json:"condition_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"` // open, acknowledged, closed
	Severity    string     `json:"severity"`
	OpenedAt    time.Time  `json:"opened_at"`
	ClosedAt    *time.Time `json:"closed_at"`
	AckedAt     *time.Time `json:"acked_at"`
	AckedBy     string     `json:"acked_by"`
}

// Create Alert Condition
func (s *EnhancedAlertService) CreateCondition(condition *AlertCondition) error {
	return postgres.DB.Create(condition).Error
}

func (s *EnhancedAlertService) GetConditions(tenantID int64) ([]AlertCondition, error) {
	var conditions []AlertCondition
	err := postgres.DB.Where("tenant_id = ?", tenantID).Find(&conditions).Error
	return conditions, err
}

func (s *EnhancedAlertService) UpdateCondition(id int64, tenantID int64, updates map[string]interface{}) error {
	return postgres.DB.Model(&AlertCondition{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(updates).Error
}

func (s *EnhancedAlertService) DeleteCondition(id int64, tenantID int64) error {
	return postgres.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&AlertCondition{}).Error
}

// Create Alert Channel
func (s *EnhancedAlertService) CreateChannel(channel *AlertChannel) error {
	return postgres.DB.Create(channel).Error
}

func (s *EnhancedAlertService) GetChannels(tenantID int64) ([]AlertChannel, error) {
	var channels []AlertChannel
	err := postgres.DB.Where("tenant_id = ?", tenantID).Find(&channels).Error
	return channels, err
}

func (s *EnhancedAlertService) UpdateChannel(id int64, tenantID int64, updates map[string]interface{}) error {
	return postgres.DB.Model(&AlertChannel{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(updates).Error
}

func (s *EnhancedAlertService) DeleteChannel(id int64, tenantID int64) error {
	return postgres.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&AlertChannel{}).Error
}

// Create Alert Policy
func (s *EnhancedAlertService) CreatePolicy(policy *AlertPolicy) error {
	return postgres.DB.Create(policy).Error
}

func (s *EnhancedAlertService) GetPolicies(tenantID int64) ([]AlertPolicy, error) {
	var policies []AlertPolicy
	err := postgres.DB.Where("tenant_id = ?", tenantID).Find(&policies).Error
	return policies, err
}

func (s *EnhancedAlertService) UpdatePolicy(id int64, tenantID int64, updates map[string]interface{}) error {
	return postgres.DB.Model(&AlertPolicy{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(updates).Error
}

func (s *EnhancedAlertService) DeletePolicy(id int64, tenantID int64) error {
	return postgres.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&AlertPolicy{}).Error
}

// Incident Management
func (s *EnhancedAlertService) CreateIncident(incident *AlertIncident) error {
	return postgres.DB.Create(incident).Error
}

func (s *EnhancedAlertService) GetIncidents(tenantID int64, status string) ([]AlertIncident, error) {
	var incidents []AlertIncident
	query := postgres.DB.Where("tenant_id = ?", tenantID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("opened_at DESC").Find(&incidents).Error
	return incidents, err
}

func (s *EnhancedAlertService) AcknowledgeIncident(id int64, tenantID int64, ackedBy string) error {
	now := time.Now()
	return postgres.DB.Model(&AlertIncident{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(map[string]interface{}{
			"status":   "acknowledged",
			"acked_at": &now,
			"acked_by": ackedBy,
		}).Error
}

func (s *EnhancedAlertService) CloseIncident(id int64, tenantID int64) error {
	now := time.Now()
	return postgres.DB.Model(&AlertIncident{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(map[string]interface{}{
			"status":    "closed",
			"closed_at": &now,
		}).Error
}

// Alert Evaluation Engine
func (s *EnhancedAlertService) EvaluateConditions() {
	// Get all enabled conditions
	var conditions []AlertCondition
	err := postgres.DB.Where("enabled = ?", true).Find(&conditions).Error
	if err != nil {
		return
	}

	for _, condition := range conditions {
		go s.evaluateCondition(&condition)
	}
}

func (s *EnhancedAlertService) evaluateCondition(condition *AlertCondition) {
	// Parse and execute the query
	result, err := s.executeQuery(condition.Query, condition.TenantID)
	if err != nil {
		return
	}

	// Check if threshold is breached
	triggered := false
	switch condition.Operator {
	case "above":
		triggered = result > condition.Threshold
	case "below":
		triggered = result < condition.Threshold
	case "equals":
		triggered = result == condition.Threshold
	}

	if triggered {
		// Check if incident already exists
		var existingIncident AlertIncident
		err := postgres.DB.Where("condition_id = ? AND status = ?",
			condition.ID, "open").First(&existingIncident).Error

		if err != nil {
			// Create new incident
			incident := &AlertIncident{
				TenantID:    condition.TenantID,
				ConditionID: condition.ID,
				Title:       fmt.Sprintf("Alert: %s", condition.Name),
				Description: fmt.Sprintf("Condition '%s' triggered. Value: %.2f, Threshold: %.2f",
					condition.Name, result, condition.Threshold),
				Status:   "open",
				Severity: condition.Severity,
				OpenedAt: time.Now(),
			}
			s.CreateIncident(incident)

			// Send notifications
			s.sendNotifications(incident)
		}
	} else {
		// Close any open incidents for this condition
		now := time.Now()
		postgres.DB.Model(&AlertIncident{}).
			Where("condition_id = ? AND status = ?", condition.ID, "open").
			Updates(map[string]interface{}{
				"status":    "closed",
				"closed_at": &now,
			})
	}
}

func (s *EnhancedAlertService) executeQuery(_query string, tenantID int64) (float64, error) {
	// Simplified query execution - in production, implement full NRQL parser
	// For now, support basic metric queries

	// Example: "SELECT average(cpu_usage) FROM metrics WHERE host_id = 1"
	// This is a simplified implementation

	var result float64
	err := postgres.DB.Raw("SELECT AVG(value) FROM metrics WHERE tenant_id = ? AND name = 'cpu_usage' AND timestamp > NOW() - INTERVAL '5 minutes'",
		tenantID).Scan(&result).Error

	return result, err
}

func (s *EnhancedAlertService) sendNotifications(incident *AlertIncident) {
	// Get notification channels for this policy
	var channels []AlertChannel
	err := postgres.DB.Where("tenant_id = ? AND enabled = ?",
		incident.TenantID, true).Find(&channels).Error
	if err != nil {
		return
	}

	for _, channel := range channels {
		go s.sendNotification(&channel, incident)
	}
}

func (s *EnhancedAlertService) sendNotification(channel *AlertChannel, incident *AlertIncident) {
	switch channel.Type {
	case "email":
		s.sendEmailNotification(channel, incident)
	case "slack":
		s.sendSlackNotification(channel, incident)
	case "webhook":
		s.sendWebhookNotification(channel, incident)
	case "pagerduty":
		s.sendPagerDutyNotification(channel, incident)
	}
}

func (s *EnhancedAlertService) sendEmailNotification(channel *AlertChannel, incident *AlertIncident) {
	// Parse email config
	var config struct {
		Recipients []string `json:"recipients"`
		Subject    string   `json:"subject"`
	}
	json.Unmarshal([]byte(channel.Config), &config)

	// Send email (implement with your email service)
	fmt.Printf("EMAIL ALERT: %s - %s\n", incident.Title, incident.Description)
}

func (s *EnhancedAlertService) sendSlackNotification(channel *AlertChannel, incident *AlertIncident) {
	// Parse Slack config
	var config struct {
		WebhookURL string `json:"webhook_url"`
		Channel    string `json:"channel"`
	}
	json.Unmarshal([]byte(channel.Config), &config)

	// Send Slack message (implement with Slack API)
	fmt.Printf("SLACK ALERT: %s - %s\n", incident.Title, incident.Description)
}

func (s *EnhancedAlertService) sendWebhookNotification(channel *AlertChannel, incident *AlertIncident) {
	// Parse webhook config
	var config struct {
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers"`
	}
	json.Unmarshal([]byte(channel.Config), &config)

	// Send webhook (implement HTTP POST)
	fmt.Printf("WEBHOOK ALERT: %s - %s\n", incident.Title, incident.Description)
}

func (s *EnhancedAlertService) sendPagerDutyNotification(channel *AlertChannel, incident *AlertIncident) {
	// Parse PagerDuty config
	var config struct {
		IntegrationKey string `json:"integration_key"`
		Severity       string `json:"severity"`
	}
	json.Unmarshal([]byte(channel.Config), &config)

	// Send PagerDuty alert (implement with PagerDuty API)
	fmt.Printf("PAGERDUTY ALERT: %s - %s\n", incident.Title, incident.Description)
}

// Start the alert evaluation scheduler
func (s *EnhancedAlertService) StartScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			s.EvaluateConditions()
		}
	}()
}
