package services

import (
	"errors"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/auth"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"gorm.io/gorm"
)

// register a new user
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

// authenticate user and return token if OK
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
	token, err := auth.GenerateJWT(user.ID, user.Username, time.Minute*15)
	return token, user.ID, err
}

// password reset (mock logic)
func ForgotPassword(email string) error {
	// In production: send email, create reset token in DB
	return nil
}

// create audit log
func LogEvent(userID int64, event, details string) {
	log := models.AuditLog{
		UserID:    userID,
		Event:     event,
		Timestamp: time.Now(),
		Details:   details,
	}
	// Do not block main handler if DB slow
	go postgres.DB.Create(&log)
}
