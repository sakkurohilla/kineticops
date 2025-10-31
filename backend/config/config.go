package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

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
	_ = godotenv.Load(".env")
	viper.AutomaticEnv()
	viper.SetDefault("APP_PORT", "8080")
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
