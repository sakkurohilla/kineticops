package handlers

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"go.mongodb.org/mongo-driver/mongo"
)

type HealthCheckHandler struct {
	// Dependencies
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewHealthCheckHandler(redisClient *redis.Client, mongoClient *mongo.Client) *HealthCheckHandler {
	return &HealthCheckHandler{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}

// BasicHealthCheck provides a simple liveness check
func BasicHealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}

// DetailedHealthCheck provides comprehensive health information
func (h *HealthCheckHandler) DetailedHealthCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health := fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"services":  fiber.Map{},
	}

	overallHealthy := true
	services := make(map[string]interface{})

	// Check PostgreSQL
	pgStatus := h.checkPostgreSQL(ctx)
	services["postgresql"] = pgStatus
	if status, ok := pgStatus["status"].(string); ok && status != "healthy" {
		overallHealthy = false
	}

	// Check MongoDB
	mongoStatus := h.checkMongoDB(ctx)
	services["mongodb"] = mongoStatus
	if status, ok := mongoStatus["status"].(string); ok && status != "healthy" {
		overallHealthy = false
	}

	// Check Redis
	redisStatus := h.checkRedis(ctx)
	services["redis"] = redisStatus
	if status, ok := redisStatus["status"].(string); ok && status != "healthy" {
		overallHealthy = false
	}

	// Check circuit breakers
	circuitStatus := h.checkCircuitBreakers()
	services["circuit_breakers"] = circuitStatus

	health["services"] = services

	if !overallHealthy {
		health["status"] = "unhealthy"
		return c.Status(fiber.StatusServiceUnavailable).JSON(health)
	}

	return c.JSON(health)
}

func (h *HealthCheckHandler) checkPostgreSQL(_ context.Context) fiber.Map {
	start := time.Now()

	// Try to ping the database using SqlxDB
	var err error
	if postgres.SqlxDB != nil {
		err = postgres.SqlxDB.Ping()
	} else {
		err = fiber.NewError(fiber.StatusServiceUnavailable, "PostgreSQL not initialized")
	}
	latency := time.Since(start).Milliseconds()

	if err != nil {
		logging.Errorf("PostgreSQL health check failed: %v", err)
		return fiber.Map{
			"status":  "unhealthy",
			"error":   err.Error(),
			"latency": latency,
		}
	}

	// Get connection pool stats
	var stats interface{}
	if postgres.SqlxDB != nil {
		stats = postgres.SqlxDB.Stats()
	}

	return fiber.Map{
		"status":      "healthy",
		"latency":     latency,
		"connections": stats,
	}
}

func (h *HealthCheckHandler) checkMongoDB(ctx context.Context) fiber.Map {
	start := time.Now()

	// Try to ping MongoDB
	err := h.mongoClient.Ping(ctx, nil)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		logging.Errorf("MongoDB health check failed: %v", err)
		return fiber.Map{
			"status":  "unhealthy",
			"error":   err.Error(),
			"latency": latency,
		}
	}

	return fiber.Map{
		"status":  "healthy",
		"latency": latency,
	}
}

func (h *HealthCheckHandler) checkRedis(ctx context.Context) fiber.Map {
	start := time.Now()

	// Try to ping Redis
	_, err := h.redisClient.Ping(ctx).Result()
	latency := time.Since(start).Milliseconds()

	if err != nil {
		logging.Errorf("Redis health check failed: %v", err)
		return fiber.Map{
			"status":  "unhealthy",
			"error":   err.Error(),
			"latency": latency,
		}
	}

	// Get Redis info (optional)
	_, _ = h.redisClient.Info(ctx, "stats").Result()

	return fiber.Map{
		"status":  "healthy",
		"latency": latency,
	}
}

func (h *HealthCheckHandler) checkCircuitBreakers() fiber.Map {
	return fiber.Map{
		"mongodb":  middleware.MongoDBCircuitBreaker.GetMetrics(),
		"redis":    middleware.RedisCircuitBreaker.GetMetrics(),
		"redpanda": middleware.RedpandaCircuitBreaker.GetMetrics(),
	}
}

// ReadinessCheck checks if the service is ready to accept traffic
func (h *HealthCheckHandler) ReadinessCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Check critical dependencies
	if postgres.SqlxDB != nil {
		if err := postgres.SqlxDB.Ping(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"ready":  false,
				"reason": "PostgreSQL not available",
			})
		}
	}

	if err := h.mongoClient.Ping(ctx, nil); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"ready":  false,
			"reason": "MongoDB not available",
		})
	}

	if _, err := h.redisClient.Ping(ctx).Result(); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"ready":  false,
			"reason": "Redis not available",
		})
	}

	return c.JSON(fiber.Map{
		"ready": true,
	})
}

// LivenessCheck checks if the service is alive
func LivenessCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"alive": true,
	})
}
