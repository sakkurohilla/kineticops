package models

import "time"

type Host struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	IPAddress string    `db:"ip_address" json:"ip_address"`
	Status    string    `db:"status" json:"status"`
	LastSeen  time.Time `db:"last_seen" json:"last_seen"`
	OwnerID   int       `db:"owner_id" json:"owner_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
