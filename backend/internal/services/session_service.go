package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
)

const (
	SessionTimeout        = 24 * time.Hour // Session expires after 24 hours
	SessionIdleTimeout    = 2 * time.Hour  // Session expires after 2 hours of inactivity
	MaxConcurrentSessions = 5              // Max concurrent sessions per user
)

type SessionService struct {
	redis *redis.Client
	mu    sync.RWMutex
}

type Session struct {
	UserID       int64     `json:"user_id"`
	SessionID    string    `json:"session_id"`
	Username     string    `json:"username"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	IsActive     bool      `json:"is_active"`
}

func NewSessionService(redisClient *redis.Client) *SessionService {
	return &SessionService{
		redis: redisClient,
	}
}

// CreateSession creates a new session and enforces concurrent session limits
func (s *SessionService) CreateSession(ctx context.Context, userID int64, username, sessionID, ipAddress, userAgent string) error {
	// Get all active sessions for this user
	sessionsKey := fmt.Sprintf("user:sessions:%d", userID)
	sessions, err := s.redis.SMembers(ctx, sessionsKey).Result()
	if err != nil && err != redis.Nil {
		logging.Errorf("Failed to get user sessions: %v", err)
	}

	// If user has max sessions, remove the oldest one
	if len(sessions) >= MaxConcurrentSessions {
		// Get the oldest session
		oldestSession := ""
		oldestTime := time.Now()

		for _, sid := range sessions {
			sessionKey := fmt.Sprintf("session:%s", sid)
			lastActivity, err := s.redis.HGet(ctx, sessionKey, "last_activity").Result()
			if err != nil {
				continue
			}

			activityTime, err := time.Parse(time.RFC3339, lastActivity)
			if err != nil {
				continue
			}

			if activityTime.Before(oldestTime) {
				oldestTime = activityTime
				oldestSession = sid
			}
		}

		if oldestSession != "" {
			s.RevokeSession(ctx, oldestSession)
			logging.Infof("Removed oldest session %s for user %d due to concurrent session limit", oldestSession, userID)
		}
	}

	// Create new session
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	session := map[string]interface{}{
		"user_id":       userID,
		"username":      username,
		"created_at":    time.Now().Format(time.RFC3339),
		"last_activity": time.Now().Format(time.RFC3339),
		"ip_address":    ipAddress,
		"user_agent":    userAgent,
		"is_active":     "true",
	}

	// Store session data
	if err := s.redis.HSet(ctx, sessionKey, session).Err(); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Set session expiration
	if err := s.redis.Expire(ctx, sessionKey, SessionTimeout).Err(); err != nil {
		return fmt.Errorf("failed to set session expiration: %w", err)
	}

	// Add session to user's session set
	if err := s.redis.SAdd(ctx, sessionsKey, sessionID).Err(); err != nil {
		return fmt.Errorf("failed to add session to user set: %w", err)
	}

	// Set user session set expiration
	s.redis.Expire(ctx, sessionsKey, SessionTimeout)

	logging.Infof("Created session %s for user %d (%s) from IP %s", sessionID, userID, username, ipAddress)
	return nil
}

// ValidateSession validates a session and updates last activity
func (s *SessionService) ValidateSession(ctx context.Context, sessionID string) (*Session, error) {
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	// Get session data
	data, err := s.redis.HGetAll(ctx, sessionKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("session not found")
	}

	// Check if session is active
	if data["is_active"] != "true" {
		return nil, fmt.Errorf("session is not active")
	}

	// Parse last activity
	lastActivity, err := time.Parse(time.RFC3339, data["last_activity"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse last activity: %w", err)
	}

	// Check idle timeout
	if time.Since(lastActivity) > SessionIdleTimeout {
		s.RevokeSession(ctx, sessionID)
		return nil, fmt.Errorf("session expired due to inactivity")
	}

	// Update last activity
	s.redis.HSet(ctx, sessionKey, "last_activity", time.Now().Format(time.RFC3339))

	// Extend session expiration
	s.redis.Expire(ctx, sessionKey, SessionTimeout)

	// Parse user ID
	var userID int64
	fmt.Sscanf(data["user_id"], "%d", &userID)

	createdAt, _ := time.Parse(time.RFC3339, data["created_at"])

	return &Session{
		UserID:       userID,
		SessionID:    sessionID,
		Username:     data["username"],
		CreatedAt:    createdAt,
		LastActivity: time.Now(),
		IPAddress:    data["ip_address"],
		UserAgent:    data["user_agent"],
		IsActive:     true,
	}, nil
}

// RevokeSession revokes a session
func (s *SessionService) RevokeSession(ctx context.Context, sessionID string) error {
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	// Get user ID before deleting
	userIDStr, err := s.redis.HGet(ctx, sessionKey, "user_id").Result()
	if err == nil {
		var userID int64
		fmt.Sscanf(userIDStr, "%d", &userID)

		// Remove from user's session set
		sessionsKey := fmt.Sprintf("user:sessions:%d", userID)
		s.redis.SRem(ctx, sessionsKey, sessionID)
	}

	// Delete session
	if err := s.redis.Del(ctx, sessionKey).Err(); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	logging.Infof("Revoked session %s", sessionID)
	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (s *SessionService) RevokeAllUserSessions(ctx context.Context, userID int64) error {
	sessionsKey := fmt.Sprintf("user:sessions:%d", userID)

	// Get all sessions
	sessions, err := s.redis.SMembers(ctx, sessionsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Revoke each session
	for _, sessionID := range sessions {
		s.RevokeSession(ctx, sessionID)
	}

	// Delete the session set
	s.redis.Del(ctx, sessionsKey)

	logging.Infof("Revoked all sessions for user %d", userID)
	return nil
}

// GetActiveSessions returns all active sessions for a user
func (s *SessionService) GetActiveSessions(ctx context.Context, userID int64) ([]Session, error) {
	sessionsKey := fmt.Sprintf("user:sessions:%d", userID)

	sessionIDs, err := s.redis.SMembers(ctx, sessionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	var sessions []Session
	for _, sessionID := range sessionIDs {
		sessionKey := fmt.Sprintf("session:%s", sessionID)
		data, err := s.redis.HGetAll(ctx, sessionKey).Result()
		if err != nil || len(data) == 0 {
			continue
		}

		if data["is_active"] != "true" {
			continue
		}

		var uid int64
		fmt.Sscanf(data["user_id"], "%d", &uid)

		createdAt, _ := time.Parse(time.RFC3339, data["created_at"])
		lastActivity, _ := time.Parse(time.RFC3339, data["last_activity"])

		sessions = append(sessions, Session{
			UserID:       uid,
			SessionID:    sessionID,
			Username:     data["username"],
			CreatedAt:    createdAt,
			LastActivity: lastActivity,
			IPAddress:    data["ip_address"],
			UserAgent:    data["user_agent"],
			IsActive:     true,
		})
	}

	return sessions, nil
}

// CleanupExpiredSessions removes expired sessions (run periodically)
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	// This is handled automatically by Redis TTL
	// But we can scan for orphaned session references
	logging.Infof("Session cleanup completed (handled by Redis TTL)")
	return nil
}
