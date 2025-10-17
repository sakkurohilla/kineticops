package repository

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BaseRepository struct {
	DB *pgxpool.Pool
}

func NewBaseRepository(db *pgxpool.Pool) *BaseRepository {
	return &BaseRepository{DB: db}
}

// Helper function to execute queries with context and timeout
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.DB.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("❌ Exec Error: %v", err)
		return err
	}
	return nil
}
