package redisrepo

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var Client *redis.Client

func Init() error {
	Client = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("REDIS_ADDR"), // e.g. "localhost:6379"
		Password: "",                            // update if you use `requirepass` in production
		DB:       0,                             // default DB
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Client.Ping(ctx).Err(); err != nil {
		log.Printf("Redis connection failed: %v", err)
		return err
	}
	return nil
}
