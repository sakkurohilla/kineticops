package postgres

import (
	"github.com/jmoiron/sqlx"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"gorm.io/gorm"
)

// GORM-based functions (existing)
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

// New sqlx-based repository for agent service
type HostRepository struct {
	db *sqlx.DB
}

func NewHostRepository(db *sqlx.DB) *HostRepository {
	return &HostRepository{db: db}
}

func (r *HostRepository) Create(host *models.Host) error {
	query := `
		INSERT INTO hosts (hostname, ip, ssh_user, ssh_password, ssh_key, ssh_port, os, "group", tags, tenant_id, reg_token, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP)
		RETURNING id, created_at`
	
	return r.db.QueryRow(query, host.Hostname, host.IP, host.SSHUser, host.SSHPassword, 
		host.SSHKey, int(host.SSHPort), host.OS, host.Group, host.Tags, 
		host.TenantID, host.RegToken).Scan(&host.ID, &host.CreatedAt)
}

func (r *HostRepository) GetByID(id int) (*models.Host, error) {
	var host models.Host
	query := `SELECT * FROM hosts WHERE id = $1`
	err := r.db.Get(&host, query, id)
	if err != nil {
		return nil, err
	}
	return &host, nil
}