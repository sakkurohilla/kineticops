package redisrepo

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var Client *redis.ClusterClient

func Init() error {
	addrs := []string{viper.GetString("REDIS_ADDR")} // e.g. ["localhost:6379"]
	Client = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addrs,
		// ... more options like PoolSize if needed
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Client.Ping(ctx).Err(); err != nil {
		log.Printf("Redis cluster connection error: %v", err)
		return err
	}
	return nil
}
