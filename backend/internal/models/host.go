package models

import (
	"time"

	"github.com/google/uuid"
)

type Host struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	IPAddress   string    `json:"ip_address"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // online/offline
	CreatedAt   time.Time `json:"created_at"`
}

// Constructor
func NewHost(userID uuid.UUID, name, ip, desc string) *Host {
	return &Host{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        name,
		IPAddress:   ip,
		Description: desc,
		Status:      "offline",
		CreatedAt:   time.Now(),
	}
}
