package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	redisrepo "github.com/sakkurohilla/kineticops/backend/internal/repository/redis"
)

// Optional schema validation for metric types
func ValidateMetricSchema(name string, val float64, labels string) error {
	if val < 0 && (name == "cpu_usage" || name == "ram_usage") {
		return fmt.Errorf("invalid metric value")
	}
	return nil
}

func CollectMetric(hostID, tenantID int64, name string, value float64, labels map[string]string) error {
	lb, _ := json.Marshal(labels)
	metric := &models.Metric{
		HostID:    hostID,
		TenantID:  tenantID,
		Name:      name,
		Value:     value,
		Timestamp: time.Now(),
		Labels:    string(lb),
	}
	return postgres.SaveMetric(postgres.DB, metric)
}

func ListMetrics(tenantID, hostID int64, name string, start, end time.Time, limit int) ([]models.Metric, error) {
	cacheKey := fmt.Sprintf("metrics:%d:%d:%s:%s:%s:%d", tenantID, hostID, name, start.Format(time.RFC3339), end.Format(time.RFC3339), limit)
	cached, err := redisrepo.GetMetricsCache(cacheKey)
	if err == nil && cached != nil {
		return cached, nil // Return cached result
	}

	data, err := postgres.ListMetrics(postgres.DB, tenantID, hostID, name, start, end, limit)
	if err == nil && data != nil {
		_ = redisrepo.SetMetricsCache(cacheKey, data)
	}
	return data, err
}

func LatestMetric(hostID int64, name string) (*models.Metric, error) {
	return postgres.LatestMetric(postgres.DB, hostID, name)
}

func EnforceRetentionPolicy(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return postgres.DeleteOldMetrics(postgres.DB, cutoff)
}
