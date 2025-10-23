package services

import (
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

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
