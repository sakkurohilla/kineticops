package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresDSN   string
	RedisAddr     string
	JwtSecret     string
	RefreshSecret string
	FrontendURL   string
	BackendURL    string
}

// get project root by walking up directories from file location
func getProjectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("cannot get current filepath")
	}
	// Assuming project root is 2 levels above this file (adjust if needed)
	dir := filepath.Dir(filename)
	root := filepath.Join(dir, "../../")
	absRoot, err := filepath.Abs(root)
	if err != nil {
		log.Fatal(err)
	}
	return absRoot
}

func Load() *Config {
	root := getProjectRoot()
	envPath := filepath.Join(root, ".env.local")

	if err := godotenv.Load(envPath); err != nil {
		log.Printf("No .env.local file found at %s\n", envPath)
	} else {
		log.Printf("Loaded .env.local from %s\n", envPath)
	}

	dsn := os.Getenv("POSTGRES_DSN")
	redis := os.Getenv("REDIS_ADDR")
	jwt := os.Getenv("JWT_SECRET")
	refresh := os.Getenv("REFRESH_SECRET")
	frontend := os.Getenv("FRONTEND_URL")
	backend := os.Getenv("BACKEND_URL")

	if dsn == "" || redis == "" || jwt == "" || refresh == "" {
		log.Fatal("Required environment variables missing")
	}

	return &Config{
		PostgresDSN:   dsn,
		RedisAddr:     redis,
		JwtSecret:     jwt,
		RefreshSecret: refresh,
		FrontendURL:   frontend,
		BackendURL:    backend,
	}
}
