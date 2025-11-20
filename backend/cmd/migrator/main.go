package main

import (
	"database/sql"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/sakkurohilla/kineticops/backend/internal/logging"

	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("MIGRATE_DSN")
	dir := os.Getenv("MIGRATIONS_DIR")
	if dsn == "" {
		flag.StringVar(&dsn, "dsn", "host=postgres port=5432 user=akash password=akash dbname=kineticops sslmode=disable", "Postgres DSN")
	}
	if dir == "" {
		flag.StringVar(&dir, "dir", "./migrations/postgres", "migrations directory")
	}
	flag.Parse()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logging.Errorf("failed to open db: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create schema_migrations table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		logging.Errorf("failed to create schema_migrations table: %v", err)
		os.Exit(1)
	}

	// Get already applied migrations
	applied := make(map[string]bool)
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		logging.Errorf("failed to query schema_migrations: %v", err)
		os.Exit(1)
	}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			logging.Errorf("failed to scan version: %v", err)
			os.Exit(1)
		}
		applied[version] = true
	}
	rows.Close()

	files, err := os.ReadDir(dir)
	if err != nil {
		logging.Errorf("failed to read migrations dir %s: %v", dir, err)
		os.Exit(1)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if filepath.Ext(name) != ".sql" {
			continue
		}
		// Only apply .up.sql migrations (simple convention)
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		// Skip if already applied
		version := strings.TrimSuffix(name, ".sql")
		if applied[version] {
			logging.Infof("Skipping already applied migration: %s", name)
			continue
		}

		path := filepath.Join(dir, name)
		logging.Infof("Applying migration %s...", path)
		content, err := os.ReadFile(path)
		if err != nil {
			logging.Errorf("failed to read %s: %v", path, err)
			os.Exit(1)
		}
		if _, err := db.Exec(string(content)); err != nil {
			logging.Errorf("migration %s failed: %v", path, err)
			os.Exit(1)
		}

		// Record migration as applied
		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			logging.Errorf("failed to record migration %s: %v", version, err)
			os.Exit(1)
		}

		logging.Infof("Applied %s", path)
	}

	logging.Infof("All migrations applied successfully")
}
