package models

import (
	"time"
)

type AuditLog struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     // nullable for guest events
	Event     string    // "register", "login", "logout", "failed_login", "password_reset"
	Timestamp time.Time `gorm:"autoCreateTime"`
	Details   string    // info (e.g., IP or target email)
}
