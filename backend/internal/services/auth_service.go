package services

import (
	"errors"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/auth"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"gorm.io/gorm"
)

// RegisterUser creates a new user account
func RegisterUser(username, email, password string) error {
	var count int64
	postgres.DB.Model(&models.User{}).
		Where("username=? OR email=?", username, email).Count(&count)
	if count > 0 {
		return errors.New("username or email already exists")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	user := models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
	}
	return postgres.DB.Create(&user).Error
}

// LoginUser authenticates user and returns JWT token
func LoginUser(username, password string) (string, int64, error) {
	var user models.User
	res := postgres.DB.Where("username = ?", username).First(&user)

	if res.Error == gorm.ErrRecordNotFound {
		return "", 0, errors.New("invalid credentials")
	} else if res.Error != nil {
		return "", 0, errors.New("db error")
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return "", 0, errors.New("invalid credentials")
	}

	// Token duration: 1 hour
	token, err := auth.GenerateJWT(user.ID, user.Username, 1*time.Hour)
	return token, user.ID, err
}

// Audit log helper
func LogEvent(userID int64, event, details string) {
	// Optional: log to DB or stdout
	// You can implement proper audit logging here
}
