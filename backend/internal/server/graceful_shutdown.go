package server

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"go.mongodb.org/mongo-driver/mongo"
)

type ShutdownManager struct {
	app         *fiber.App
	mongoClient *mongo.Client
	redisClient *redis.Client
	workers     []WorkerShutdown
	mu          sync.Mutex
}

type WorkerShutdown struct {
	Name   string
	Cancel context.CancelFunc
}

func NewShutdownManager(app *fiber.App, mongoClient *mongo.Client, redisClient *redis.Client) *ShutdownManager {
	return &ShutdownManager{
		app:         app,
		mongoClient: mongoClient,
		redisClient: redisClient,
		workers:     make([]WorkerShutdown, 0),
	}
}

// RegisterWorker registers a worker for graceful shutdown
func (sm *ShutdownManager) RegisterWorker(name string, cancel context.CancelFunc) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.workers = append(sm.workers, WorkerShutdown{
		Name:   name,
		Cancel: cancel,
	})
}

// WaitForShutdown waits for interrupt signal and performs graceful shutdown
func (sm *ShutdownManager) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logging.Infof("Shutdown signal received, starting graceful shutdown...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown steps
	sm.shutdown(ctx)
}

func (sm *ShutdownManager) shutdown(ctx context.Context) {
	var wg sync.WaitGroup

	// Step 1: Stop accepting new requests
	logging.Infof("Step 1: Stopping HTTP server...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := sm.app.Shutdown(); err != nil {
			logging.Errorf("Error shutting down HTTP server: %v", err)
		} else {
			logging.Infof("HTTP server stopped")
		}
	}()

	// Step 2: Cancel all background workers
	logging.Infof("Step 2: Stopping %d background workers...", len(sm.workers))
	sm.mu.Lock()
	for _, worker := range sm.workers {
		logging.Infof("Stopping worker: %s", worker.Name)
		worker.Cancel()
	}
	sm.mu.Unlock()

	// Wait a bit for workers to finish current tasks
	time.Sleep(2 * time.Second)

	// Step 3: Close database connections
	logging.Infof("Step 3: Closing database connections...")

	// Close MongoDB
	wg.Add(1)
	go func() {
		defer wg.Done()
		if sm.mongoClient != nil {
			mongoCtx, mongoCancel := context.WithTimeout(ctx, 5*time.Second)
			defer mongoCancel()

			if err := sm.mongoClient.Disconnect(mongoCtx); err != nil {
				logging.Errorf("Error disconnecting MongoDB: %v", err)
			} else {
				logging.Infof("MongoDB connection closed")
			}
		}
	}()

	// Close Redis
	wg.Add(1)
	go func() {
		defer wg.Done()
		if sm.redisClient != nil {
			if err := sm.redisClient.Close(); err != nil {
				logging.Errorf("Error closing Redis: %v", err)
			} else {
				logging.Infof("Redis connection closed")
			}
		}
	}()

	// Close PostgreSQL
	wg.Add(1)
	go func() {
		defer wg.Done()
		if postgres.SqlxDB != nil {
			if err := postgres.SqlxDB.Close(); err != nil {
				logging.Errorf("Error closing PostgreSQL: %v", err)
			} else {
				logging.Infof("PostgreSQL connection closed")
			}
		}

		if postgres.DB != nil {
			db, err := postgres.DB.DB()
			if err == nil {
				if err := db.Close(); err != nil {
					logging.Errorf("Error closing GORM DB: %v", err)
				} else {
					logging.Infof("GORM DB connection closed")
				}
			}
		}
	}()

	// Wait for all shutdown tasks to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logging.Infof("Graceful shutdown completed successfully")
	case <-ctx.Done():
		logging.Warnf("Shutdown timeout exceeded, forcing exit")
	}

	// Step 4: Flush logs
	logging.Flush()

	logging.Infof("Goodbye!")
}
