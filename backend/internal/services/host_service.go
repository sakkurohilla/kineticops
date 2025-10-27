package services

import (
	"strconv"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"gorm.io/gorm"
)

// RegisterHost creates a new host
func RegisterHost(hostname, ip, os, group, tags string, tenantID int64, token string) (*models.Host, error) {
	host := &models.Host{
		Hostname:    hostname,
		IP:          ip,
		OS:          os,
		Group:       group,
		Tags:        tags,
		TenantID:    tenantID,
		AgentStatus: "offline",
		RegToken:    token,
		LastSeen:    time.Now(),
	}
	err := postgres.CreateHost(postgres.DB, host)
	return host, err
}

// ListHosts returns all hosts for a tenant
func ListHosts(tenantID int64, limitStr, offsetStr string) ([]models.Host, error) {
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var hosts []models.Host
	result := postgres.DB.Where("tenant_id = ?", tenantID).
		Limit(limit).
		Offset(offset).
		Find(&hosts)

	// Return empty array instead of error when no records found
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}

	return hosts, nil
}

// GetHostByID retrieves a single host
func GetHostByID(hostID, tenantID int64) (*models.Host, error) {
	return postgres.GetHost(postgres.DB, hostID)
}

// UpdateHostFields updates specific host fields
func UpdateHostFields(hostID int64, fields map[string]interface{}) error {
	return postgres.UpdateHost(postgres.DB, hostID, fields)
}

// DeleteHostByID deletes a host
func DeleteHostByID(hostID int64) error {
	return postgres.DeleteHost(postgres.DB, hostID)
}
