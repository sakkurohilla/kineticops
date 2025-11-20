package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
)

// TokenRotationService handles agent token rotation
type TokenRotationService struct{}

func NewTokenRotationService() *TokenRotationService {
	return &TokenRotationService{}
}

// RotateToken generates a new token for an agent while maintaining grace period
func (s *TokenRotationService) RotateToken(agentID int64) (*models.AgentToken, error) {
	// Get current active token
	var currentToken models.AgentToken
	err := postgres.DB.Where("agent_id = ? AND is_active = ?", agentID, true).First(&currentToken).Error
	if err != nil {
		// No existing token, create first one
		return s.GenerateToken(agentID, 1)
	}

	// Generate new token
	newTokenStr, err := generateSecureToken()
	if err != nil {
		return nil, err
	}

	// Create new token with incremented version
	newToken := &models.AgentToken{
		AgentID:       agentID,
		Token:         newTokenStr,
		PreviousToken: currentToken.Token,
		Version:       currentToken.Version + 1,
		IsActive:      true,
		ExpiresAt:     timePtr(time.Now().Add(90 * 24 * time.Hour)), // 90 days
		RotatedAt:     timePtr(time.Now()),
	}

	// Start transaction
	tx := postgres.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create new token
	if err := tx.Create(newToken).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Keep old token active for grace period (24 hours)
	graceExpiry := time.Now().Add(24 * time.Hour)
	if err := tx.Model(&currentToken).Updates(map[string]interface{}{
		"expires_at": graceExpiry,
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update agent table with new token
	if err := tx.Model(&models.Agent{}).Where("id = ?", agentID).Update("agent_token", newTokenStr).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	logging.Infof("Token rotated for agent %d, version %d", agentID, newToken.Version)
	return newToken, nil
}

// GenerateToken creates the first token for an agent
func (s *TokenRotationService) GenerateToken(agentID int64, version int) (*models.AgentToken, error) {
	tokenStr, err := generateSecureToken()
	if err != nil {
		return nil, err
	}

	token := &models.AgentToken{
		AgentID:   agentID,
		Token:     tokenStr,
		Version:   version,
		IsActive:  true,
		ExpiresAt: timePtr(time.Now().Add(90 * 24 * time.Hour)),
	}

	if err := postgres.DB.Create(token).Error; err != nil {
		return nil, err
	}

	return token, nil
}

// ValidateToken checks if a token is valid and active
func (s *TokenRotationService) ValidateToken(tokenStr string) (*models.AgentToken, error) {
	var token models.AgentToken
	err := postgres.DB.Where("token = ? AND is_active = ?", tokenStr, true).First(&token).Error
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if expired
	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return &token, nil
}

// RevokeToken deactivates a token
func (s *TokenRotationService) RevokeToken(tokenStr string) error {
	return postgres.DB.Model(&models.AgentToken{}).
		Where("token = ?", tokenStr).
		Updates(map[string]interface{}{
			"is_active":  false,
			"revoked_at": time.Now(),
		}).Error
}

// CleanupExpiredTokens removes old expired tokens
func (s *TokenRotationService) CleanupExpiredTokens(ctx context.Context) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Delete tokens expired more than 7 days ago
			cutoff := time.Now().Add(-7 * 24 * time.Hour)
			result := postgres.DB.Where("expires_at < ? AND is_active = ?", cutoff, false).Delete(&models.AgentToken{})
			if result.Error != nil {
				logging.Errorf("Failed to cleanup expired tokens: %v", result.Error)
			} else if result.RowsAffected > 0 {
				logging.Infof("Cleaned up %d expired tokens", result.RowsAffected)
			}
		}
	}
}

func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Global token rotation service
var TokenRotationSvc = NewTokenRotationService()
