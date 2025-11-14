package config

import (
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var mu sync.Mutex

// Config captures runtime configuration loaded from environment variables.
type Config struct {
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	MongoURI         string
	RedisAddr        string
	RedpandaBroker   string
	AppEnv           string
	AppPort          string
	JWTSecret        string
	AgentToken       string
}

func Load() *Config {
	// Guard viper usage with a mutex to avoid concurrent map writes when called
	// from multiple goroutines. We intentionally re-run godotenv.Load so a
	// SIGHUP-triggered Reload can update environment-backed settings.
	mu.Lock()
	defer mu.Unlock()

	// Try loading repository-local backend/.env first (used by scripts), then fall
	// back to a top-level .env if present. Both are optional; environment
	// variables will still be picked up via AutomaticEnv().
	_ = godotenv.Load("backend/.env")
	_ = godotenv.Load(".env")
	viper.AutomaticEnv()

	// sensible defaults
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")

	return &Config{
		PostgresHost:     viper.GetString("POSTGRES_HOST"),
		PostgresPort:     viper.GetString("POSTGRES_PORT"),
		PostgresUser:     viper.GetString("POSTGRES_USER"),
		PostgresPassword: viper.GetString("POSTGRES_PASSWORD"),
		PostgresDB:       viper.GetString("POSTGRES_DB"),
		MongoURI:         viper.GetString("MONGO_URI"),
		RedisAddr:        viper.GetString("REDIS_ADDR"),
		RedpandaBroker:   viper.GetString("REDPANDA_BROKER"),
		AppEnv:           viper.GetString("APP_ENV"),
		AppPort:          viper.GetString("APP_PORT"),
		JWTSecret:        viper.GetString("JWT_SECRET"),
		AgentToken:       viper.GetString("AGENT_TOKEN"),
	}
}

// Reload re-reads environment files and updates viper's environment mapping.
// Call this when the process receives SIGHUP to pick up runtime configuration
// changes without restarting the whole process.
func Reload() {
	mu.Lock()
	defer mu.Unlock()
	// Re-load environment files and refresh viper's AutomaticEnv mapping.
	_ = godotenv.Load("backend/.env")
	_ = godotenv.Load(".env")
	viper.AutomaticEnv()
}
