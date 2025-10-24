package services

import (
	"errors"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/auth"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"gorm.io/gorm"
)

// Register a new user
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

// Authenticate user, return JWT token if success
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

	// Increase token duration to 1 hour instead of 15 minutes
	token, err := auth.GenerateJWT(user.ID, user.Username, 1*time.Hour)
	return token, user.ID, err
}

// GetUserByID retrieves a user by their ID
func GetUserByID(userID int64) (*models.User, error) {
	var user models.User
	result := postgres.DB.First(&user, userID)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, errors.New("user not found")
	} else if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// Password reset (mock)
func ForgotPassword(email string) error {
	// No real action, for mock/dev only
	return nil
}

// Audit log
func LogEvent(userID int64, event, details string) {
	// Optional: log to DB or stdout
}
