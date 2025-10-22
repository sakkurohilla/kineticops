package models

import "time"

type User struct {
	ID       int64  `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
	Role     string `db:"role"`
}

type Workspace struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	OwnerID   int64     `db:"owner_id" json:"owner_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
