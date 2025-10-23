package postgres

import (
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"gorm.io/gorm"
)

func CreateHost(db *gorm.DB, host *models.Host) error {
	return db.Create(host).Error
}

func GetHost(db *gorm.DB, id int64) (*models.Host, error) {
	var host models.Host
	err := db.First(&host, id).Error
	return &host, err
}

func ListHosts(db *gorm.DB, tenantID int64, limit, offset int) ([]models.Host, error) {
	var hosts []models.Host
	err := db.Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Find(&hosts).Error
	return hosts, err
}

func UpdateHost(db *gorm.DB, id int64, fields map[string]interface{}) error {
	return db.Model(&models.Host{}).Where("id = ?", id).Updates(fields).Error
}

func DeleteHost(db *gorm.DB, id int64) error {
	return db.Delete(&models.Host{}, id).Error
}
