package config

import (
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var loadOnce sync.Once

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
	// Initialize viper and environment only once to avoid concurrent map writes
	// when Load() is called from multiple goroutines (for example request
	// handlers). Use sync.Once to ensure viper.SetDefault is executed a single
	// time.
	loadOnce.Do(func() {
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
	})
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
