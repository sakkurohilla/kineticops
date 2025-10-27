package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sakkurohilla/kineticops/backend/internal/auth"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	redisrepo "github.com/sakkurohilla/kineticops/backend/internal/repository/redis"
	"gorm.io/gorm"
)

const (
	resetTokenPrefix = "password_reset:"
	resetTokenExpiry = 15 * time.Minute
)

// GeneratePasswordResetToken creates a reset token and stores it in Redis
func GeneratePasswordResetToken(email string) (string, error) {
	// Check if user exists
	var user models.User
	result := postgres.DB.Where("email = ?", email).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		// Don't reveal if email exists or not (security best practice)
		// Return success anyway to prevent email enumeration
		return "", nil
	} else if result.Error != nil {
		return "", result.Error
	}

	// Generate unique token
	token := uuid.New().String()

	// Create reset token object
	resetToken := models.PasswordResetToken{
		Token:     token,
		UserID:    user.ID,
		Email:     user.Email,
		ExpiresAt: time.Now().Add(resetTokenExpiry),
		CreatedAt: time.Now(),
		Used:      false,
	}

	// Store in Redis with expiry
	ctx := context.Background()
	tokenJSON, err := json.Marshal(resetToken)
	if err != nil {
		return "", err
	}

	redisKey := fmt.Sprintf("%s%s", resetTokenPrefix, token)
	err = redisrepo.Client.Set(ctx, redisKey, tokenJSON, resetTokenExpiry).Err()
	if err != nil {
		return "", err
	}

	LogEvent(user.ID, "password_reset_requested", user.Email)

	return token, nil
}

// VerifyPasswordResetToken checks if token is valid
func VerifyPasswordResetToken(token string) (*models.PasswordResetToken, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("%s%s", resetTokenPrefix, token)

	// Get token from Redis
	tokenJSON, err := redisrepo.Client.Get(ctx, redisKey).Result()
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	var resetToken models.PasswordResetToken
	if err := json.Unmarshal([]byte(tokenJSON), &resetToken); err != nil {
		return nil, errors.New("invalid token format")
	}

	// Check if already used
	if resetToken.Used {
		return nil, errors.New("token already used")
	}

	// Check expiry
	if time.Now().After(resetToken.ExpiresAt) {
		return nil, errors.New("token expired")
	}

	return &resetToken, nil
}

// ResetPassword resets user password with valid token
func ResetPassword(token, newPassword string) error {
	// Verify token
	resetToken, err := VerifyPasswordResetToken(token)
	if err != nil {
		return err
	}

	// Hash new password
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update user password
	result := postgres.DB.Model(&models.User{}).
		Where("id = ?", resetToken.UserID).
		Update("password_hash", hash)

	if result.Error != nil {
		return result.Error
	}

	// Mark token as used and delete from Redis
	ctx := context.Background()
	redisKey := fmt.Sprintf("%s%s", resetTokenPrefix, token)
	redisrepo.Client.Del(ctx, redisKey)

	LogEvent(resetToken.UserID, "password_reset_completed", resetToken.Email)

	return nil
}

// InvalidateAllUserTokens removes all reset tokens for a user (optional security feature)
func InvalidateAllUserTokens(userID int64) error {
	// This would require scanning Redis keys, which is expensive
	// For now, we rely on token expiry
	// In production, you might want to store user_id -> tokens mapping
	return nil
}
