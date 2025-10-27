package models

import "time"

// PasswordResetToken stores reset token information in memory/redis
type PasswordResetToken struct {
	Token     string    `json:"token"`
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Used      bool      `json:"used"`
}
