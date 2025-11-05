package services

import (
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// ResolveUserFromInstallationToken resolves user from installation token
func ResolveUserFromInstallationToken(token string) (*models.InstallationToken, error) {
	var installToken models.InstallationToken
	err := postgres.DB.Where("token = ?", token).
		Preload("User").
		First(&installToken).Error

	if err != nil {
		return nil, err
	}

	return &installToken, nil
}
