package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

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
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed to read migrations dir %s: %v", dir, err)
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
		fmt.Printf("Applying migration %s...\n", path)
		content, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("failed to read %s: %v", path, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			log.Fatalf("migration %s failed: %v", path, err)
		}
		fmt.Printf("Applied %s\n", path)
	}

	fmt.Println("All migrations applied successfully")
}
