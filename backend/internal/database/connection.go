package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kineticops/backend/internal/config"
)

var DB *pgxpool.Pool

func ConnectDB(cfg *config.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("❌ DB connection error: %v", err)
	}

	if err = pool.Ping(ctx); err != nil {
		log.Fatalf("❌ DB ping failed: %v", err)
	}

	DB = pool
	//log.Println("✅ Connected to PostgreSQL successfully")
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("🔌 Database connection closed")
	}
}
