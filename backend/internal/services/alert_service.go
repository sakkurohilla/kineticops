package services

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// ALERT RULE CRUD

func CreateAlertRule(rule *models.AlertRule) error {
	return postgres.CreateAlertRule(postgres.DB, rule)
}

func ListAlertRules(tenantID int64) ([]models.AlertRule, error) {
	return postgres.ListAlertRules(postgres.DB, tenantID)
}

// ALERT TRIGGERING & DEDUPLICATION

func hashForDedup(ruleID, hostID int64, metric string) string {
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%d-%d-%s", ruleID, hostID, metric)))
	return hex.EncodeToString(h.Sum(nil))
}

func CheckAndTriggerAlerts(tenantID int64, metricName string, hostID int64, value float64) error {
	// Use global postgres.DB
	rules, err := postgres.GetActiveAlertRulesForMetric(postgres.DB, tenantID, metricName)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if shouldTrigger(rule, value) {
			hash := hashForDedup(rule.ID, hostID, metricName)
			existing, _ := postgres.FindActiveAlertByDedup(postgres.DB, hash)
			if existing != nil {
				// already open, skip/dedup
				continue
			}
			// trigger
			alert := &models.Alert{
				TenantID:         tenantID,
				RuleID:           rule.ID,
				MetricName:       metricName,
				HostID:           hostID,
				Value:            value,
				Status:           "OPEN",
				DedupHash:        hash,
				EscalatedLevel:   0,
				TriggeredAt:      time.Now(),
				NotificationSent: false,
			}
			postgres.CreateAlert(postgres.DB, alert)
			if rule.NotificationWebhook != "" {
				go sendAlertWebhook(rule.NotificationWebhook, alert)
				alert.NotificationSent = true
			}
		}
	}
	return nil
}

func shouldTrigger(rule models.AlertRule, value float64) bool {
	switch rule.Operator {
	case ">":
		return value > rule.Threshold
	case "<":
		return value < rule.Threshold
	case "==":
		return value == rule.Threshold
	case "!=":
		return value != rule.Threshold
	}
	return false
}

func sendAlertWebhook(url string, alert *models.Alert) {
	payload, _ := json.Marshal(alert)
	_, _ = http.Post(url, "application/json", bytes.NewReader(payload))
}
