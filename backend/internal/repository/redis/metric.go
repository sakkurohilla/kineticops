package redisrepo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
)

var ctx = context.Background()

func GetMetricsCache(key string) ([]models.Metric, error) {
	rdb := GetRedisClient()
	str, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var metrics []models.Metric
	err = json.Unmarshal([]byte(str), &metrics)
	return metrics, err
}

func SetMetricsCache(key string, metrics []models.Metric) error {
	rdb := GetRedisClient()
	b, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, key, b, 5*time.Minute).Err()
}
