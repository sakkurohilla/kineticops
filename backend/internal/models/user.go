package models

import "time"

type User struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // don't expose in JSON
	FirstName    string    `json:"first_name" gorm:"column:first_name"`
	LastName     string    `json:"last_name" gorm:"column:last_name"`
	Phone        string    `json:"phone" gorm:"column:phone"`
	Company      string    `json:"company" gorm:"column:company"`
	Location     string    `json:"location" gorm:"column:location"`
	Role         string    `json:"role" gorm:"column:role"`
	Timezone     string    `json:"timezone" gorm:"column:timezone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
