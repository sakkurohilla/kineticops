package models

type User struct {
	ID            int64  `gorm:"primaryKey"`
	Username      string `gorm:"unique"`
	Email         string `gorm:"unique"`
	PasswordHash  string
	OauthProvider string
	OauthID       string
	MFAEnabled    bool
	MFASecret     string
}
