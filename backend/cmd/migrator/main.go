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
		logging.Infof("Applied %s", path)
	}

	logging.Infof("All migrations applied successfully")
}
