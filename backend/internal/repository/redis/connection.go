package redisrepo

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"

	"github.com/sakkurohilla/kineticops/backend/internal/logging"
)

var Client *redis.Client

func Init() error {
	Client = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("REDIS_ADDR"), // e.g. "localhost:6379"
		Password: "",                            // update if you use `requirepass` in production
		DB:       0,                             // default DB
	})

	// Retry ping with exponential backoff a few times to handle transient starts
	var lastErr error
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := Client.Ping(ctx).Err(); err != nil {
			lastErr = err
			logging.Warnf("Redis ping attempt %d to %s failed: %v", i+1, viper.GetString("REDIS_ADDR"), err)
			cancel()
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		cancel()
		lastErr = nil
		break
	}
	if lastErr != nil {
		logging.Errorf("Redis connection failed after retries to %s: %v", viper.GetString("REDIS_ADDR"), lastErr)
		return lastErr
	}
	return nil
}
func GetRedisClient() *redis.Client {
	return Client
}
