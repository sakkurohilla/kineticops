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
	ListenAddress string
}

// getProjectRoot finds the base directory for environment
func getProjectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("cannot get current filepath")
	}
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
	listen := os.Getenv("LISTEN_ADDRESS")
	if listen == "" {
		listen = ":5000"
	}

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
		ListenAddress: listen,
	}
}
