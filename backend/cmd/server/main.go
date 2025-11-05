package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/config"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/api/routes"
	kafkaevents "github.com/sakkurohilla/kineticops/backend/internal/messaging/redpanda"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	redisrepo "github.com/sakkurohilla/kineticops/backend/internal/repository/redis"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
	"github.com/sakkurohilla/kineticops/backend/internal/telemetry"
	ws "github.com/sakkurohilla/kineticops/backend/internal/websocket"
	"github.com/sakkurohilla/kineticops/backend/internal/workers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func initDBs(cfg *config.Config) {
	// PostgreSQL
	if err := postgres.Init(); err != nil {
		log.Fatalf("PostgreSQL connection error: %v", err)
	}

	// Redis
	if err := redisrepo.Init(); err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	// MongoDB (optional, for logs)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	if cfg.MongoURI == "" {
		log.Println("[WARN] MongoDB URI not configured; skipping MongoDB initialization")
	} else {
		mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
		if err != nil {
			// Mask credentials in logs where possible (URI may contain creds)
			masked := cfg.MongoURI
			// naive mask: remove substring between @ and the next / if present
			// e.g. mongodb://user:pass@host:port/db
			if at := strings.Index(masked, "@"); at != -1 {
				// find slash after @
				if slash := strings.Index(masked[at:], "/"); slash != -1 {
					masked = masked[:at+1] + "[hidden]" + masked[at+slash:]
				} else {
					masked = masked[:at+1] + "[hidden]"
				}
			}
			log.Printf("[WARN] MongoDB connection error for URI=%s: %v", masked, err)
			return
		}
		if pingErr := mongoClient.Ping(ctx, nil); pingErr != nil {
			log.Printf("[WARN] MongoDB ping error: %v", pingErr)
			return
		}
		models.LogCollection = mongoClient.Database("kineticops").Collection("logs")
		log.Println("MongoDB (logs) connected.")
	}

	log.Println("All DBs initialized.")
}

func main() {
	cfg := config.Load()
	fmt.Println("[INFO] Starting KineticOps Server...")
	fmt.Println("[DEBUG] JWTSecret loaded:", cfg.JWTSecret != "")

	initDBs(cfg)

	// Initialize repositories and services
	agentRepo := postgres.NewAgentRepository(postgres.SqlxDB)
	hostRepo := postgres.NewHostRepository(postgres.SqlxDB)
	workflowRepo := postgres.NewWorkflowRepository(postgres.SqlxDB)
	sshService := services.NewSSHService()
	agentService := services.NewAgentService(agentRepo, hostRepo, sshService)
	workflowService := services.NewWorkflowService(workflowRepo, agentRepo, hostRepo, sshService, cfg.JWTSecret)

	// Initialize handlers with services
	handlers.InitAgentHandlers(agentService)
	handlers.InitHostAgentService(agentService)
	handlers.InitWorkflowHandlers(workflowService)

	// Initialize telemetry (OpenTelemetry) - returns shutdown func
	shutdownTelemetry := telemetry.InitTelemetry()
	defer shutdownTelemetry()

	// Start Prometheus + pprof on :9090 for metrics and profiling
	telemetry.StartPrometheusServer(":9090")

	// START METRIC COLLECTOR WORKER - ADD THIS
	workers.StartMetricCollector()

	// Start enterprise retention service (30 days retention)
	retentionService := services.NewRetentionService(30)
	retentionService.StartRetentionWorker()

	// Start downsampling service for performance optimization
	downsamplingService := services.NewDownsamplingService()
	downsamplingService.StartDownsamplingWorker()

	// Set up Redpanda/Kafka
	brokers := []string{"localhost:9092"}
	topic := "metrics-events"

	// Initialize Kafka/Redpanda producer with a few retries to tolerate broker startup ordering.
	var producerInitErr error
	for i := 0; i < 3; i++ {
		if _, err := kafkaevents.InitProducer(brokers, topic); err != nil {
			producerInitErr = err
			log.Printf("Redpanda producer init attempt %d failed: %v", i+1, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		producerInitErr = nil
		break
	}
	if producerInitErr != nil {
		log.Printf("[WARN] Redpanda/Kafka producer initialization failed after retries: %v", producerInitErr)
	} else {
		log.Println("Redpanda producer initialized")
	}

	// WebSocket hub
	wsHub := ws.NewHub()
	go wsHub.Run()
	// make hub available globally for fallbacks (e.g. when Kafka is down)
	ws.SetGlobalHub(wsHub)

	// Warm-start hub with latest metrics from the metrics table
	if hosts, err := services.ListHosts(0, 1000, 0); err == nil {
		for _, h := range hosts {
			var metrics []struct {
				Name      string    `json:"name"`
				Value     float64   `json:"value"`
				Timestamp time.Time `json:"timestamp"`
			}
			err := postgres.DB.Raw(`
				WITH latest_metrics AS (
					SELECT name, value, timestamp,
						ROW_NUMBER() OVER (PARTITION BY name ORDER BY timestamp DESC) as rn
					FROM metrics WHERE host_id = ?
				)
				SELECT name, value, timestamp FROM latest_metrics WHERE rn = 1
			`, h.ID).Scan(&metrics).Error

			if err == nil && len(metrics) > 0 {
				payload := map[string]interface{}{"host_id": h.ID, "seq": telemetry.NextSeq()}
				var latestTimestamp time.Time
				for _, m := range metrics {
					payload[m.Name] = m.Value
					if m.Timestamp.After(latestTimestamp) {
						latestTimestamp = m.Timestamp
					}
				}
				payload["timestamp"] = latestTimestamp.Format(time.RFC3339)
				if b, err := json.Marshal(payload); err == nil {
					wsHub.RememberMessage(b)
					wsHub.Broadcast(b)
				}
			}
		}
	}

	// Kafka consumer to process agent events and broadcast to WebSocket clients
	kafkaevents.StartConsumer(brokers, topic, func(msg []byte) {
		fmt.Println("[DEBUG] Processing Kafka message:", string(msg))
		// Broadcast as metric update
		wsHub.RememberMessage(msg)
		wsHub.Broadcast(msg)
	})

	// Fiber app with error handler
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// Global middlewares
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())
	app.Use(middleware.RateLimiter())

	// Register ALL routes through unified router
	routes.RegisterAllRoutes(app)

	// WebSocket JWT-enabled endpoint
	app.Get("/ws", websocket.New(ws.WsHandler(wsHub, cfg.JWTSecret)))

	log.Printf("✅ Server started successfully on port %s", cfg.AppPort)
	log.Printf("📊 Health check: http://localhost:%s/health", cfg.AppPort)
	log.Printf("🔌 API Base: http://localhost:%s/api/v1", cfg.AppPort)
	log.Printf("🌍 Network: http://0.0.0.0:%s", cfg.AppPort)
	log.Printf("⚙️  Metric Collector: Running every 60s")

	// Listen on all interfaces (0.0.0.0)
	if err := app.Listen("0.0.0.0:" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
