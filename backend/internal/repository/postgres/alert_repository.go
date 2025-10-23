package postgres

import (
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"gorm.io/gorm"
)

// Alert Rules
func CreateAlertRule(db *gorm.DB, rule *models.AlertRule) error {
	return db.Create(rule).Error
}
func ListAlertRules(db *gorm.DB, tenantID int64) ([]models.AlertRule, error) {
	var rules []models.AlertRule
	err := db.Where("tenant_id = ?", tenantID).Find(&rules).Error
	return rules, err
}
func GetActiveAlertRulesForMetric(db *gorm.DB, tenantID int64, metric string) ([]models.AlertRule, error) {
	var rules []models.AlertRule
	err := db.Where("tenant_id = ? and metric_name = ?", tenantID, metric).Find(&rules).Error
	return rules, err
}

// Alerts (inc. dedup)
func CreateAlert(db *gorm.DB, alert *models.Alert) error {
	return db.Create(alert).Error
}
func FindActiveAlertByDedup(db *gorm.DB, hash string) (*models.Alert, error) {
	var alert models.Alert
	err := db.Where("dedup_hash = ? and status = 'OPEN'", hash).First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}
func CloseAlert(db *gorm.DB, id int64) error {
	now := time.Now()
	return db.Model(&models.Alert{}).Where("id = ?", id).Updates(
		map[string]interface{}{"status": "CLOSED", "closed_at": &now},
	).Error
}

// (Add more queries as needed)
