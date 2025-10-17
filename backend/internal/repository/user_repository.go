package repository

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/kineticops/backend/internal/models"
)

type UserRepository struct {
	*BaseRepository
}

func NewUserRepository(base *BaseRepository) *UserRepository {
	return &UserRepository{BaseRepository: base}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, name, email, password, created_at) VALUES ($1, $2, $3, $4, $5)`
	err := r.Exec(ctx, query, user.ID, user.Name, user.Email, user.Password, user.CreatedAt)
	if err != nil {
		log.Printf("❌ CreateUser error: %v", err)
		return err
	}
	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	query := `SELECT id, name, email, password, created_at FROM users WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		log.Printf("❌ GetUserByID error: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, name, email, password, created_at FROM users WHERE email = $1`
	err := r.DB.QueryRow(ctx, query, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		log.Printf("❌ GetUserByEmail error: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `UPDATE users SET name = $1, email = $2, password = $3 WHERE id = $4`
	err := r.Exec(ctx, query, user.Name, user.Email, user.Password, user.ID)
	if err != nil {
		log.Printf("❌ UpdateUser error: %v", err)
		return err
	}
	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	err := r.Exec(ctx, query, id)
	if err != nil {
		log.Printf("❌ DeleteUser error: %v", err)
		return err
	}
	return nil
}
