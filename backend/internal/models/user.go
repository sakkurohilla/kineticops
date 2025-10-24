package models

import "time"

type User struct {
	ID           int64  `json:"id" gorm:"primaryKey"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"` // don't expose in JSON
	CreatedAt    time.Time
}
