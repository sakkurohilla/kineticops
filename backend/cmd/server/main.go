package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/config"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/api/routes"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	kafkaevents "github.com/sakkurohilla/kineticops/backend/internal/messaging/redpanda"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	redisrepo "github.com/sakkurohilla/kineticops/backend/internal/repository/redis"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
	"github.com/sakkurohilla/kineticops/backend/internal/telemetry"
	ws "github.com/sakkurohilla/kineticops/backend/internal/websocket"
	"github.com/sakkurohilla/kineticops/backend/internal/workers"

	// Import timezone package early to set process-local timezone (time.Local)
	// to Asia/Kolkata via its init() function.
	_ "github.com/sakkurohilla/kineticops/backend/internal/timezone"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func initDBs(cfg *config.Config) {
	// PostgreSQL
	if err := postgres.Init(); err != nil {
		logging.Errorf("PostgreSQL connection error: %v", err)
	}

	// Redis
	if err := redisrepo.Init(); err != nil {
		logging.Errorf("Redis connection error: %v", err)
	}

	// MongoDB (optional, for logs)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	if cfg.MongoURI == "" {
		logging.Warnf("[WARN] MongoDB URI not configured; skipping MongoDB initialization")
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
			logging.Warnf("[WARN] MongoDB connection error for URI=%s: %v", masked, err)
			return
		}
		if pingErr := mongoClient.Ping(ctx, nil); pingErr != nil {
			logging.Warnf("[WARN] MongoDB ping error: %v", pingErr)
			return
		}
		models.LogCollection = mongoClient.Database("kineticops").Collection("logs")
		logging.Infof("MongoDB (logs) connected.")

		// Ensure indexes for efficient log search and retention
		go func() {
			ctxIdx, cancelIdx := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelIdx()

			// Text index for fast full-text search across message and full_text
			textIndex := mongo.IndexModel{
				Keys:    bson.D{{Key: "full_text", Value: "text"}, {Key: "message", Value: "text"}},
				Options: options.Index().SetName("idx_logs_fulltext"),
			}

			// Compound index for tenant/host/time queries
			compoundIdx := mongo.IndexModel{
				Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "host_id", Value: 1}, {Key: "timestamp", Value: -1}},
				Options: options.Index().SetName("idx_logs_tenant_host_time"),
			}

			if _, err := models.LogCollection.Indexes().CreateOne(ctxIdx, textIndex); err != nil {
				logging.Warnf("[WARN] failed to create text index on logs: %v", err)
			} else {
				logging.Infof("[INFO] created/ensured text index on logs")
			}

			if _, err := models.LogCollection.Indexes().CreateOne(ctxIdx, compoundIdx); err != nil {
				logging.Warnf("[WARN] failed to create compound index on logs: %v", err)
			} else {
				logging.Infof("[INFO] created/ensured compound index on logs")
			}

			// Create TTL index to enforce log retention (30 days by default)
			ttlIdx := mongo.IndexModel{
				Keys:    bson.D{{Key: "timestamp", Value: 1}},
				Options: options.Index().SetExpireAfterSeconds(60 * 60 * 24 * 30).SetName("idx_logs_ttl_30d"),
			}

			if _, err := models.LogCollection.Indexes().CreateOne(ctxIdx, ttlIdx); err != nil {
				logging.Warnf("[WARN] failed to create TTL index on logs: %v", err)
			} else {
				logging.Infof("[INFO] created/ensured TTL index on logs (30d)")
			}
		}()
	}

	logging.Infof("All DBs initialized.")
}

func main() {
	cfg := config.Load()

	// Ensure timezone is set as early as possible via blank import of
	// backend/internal/timezone (see backend/internal/timezone/timezone.go).
	// That package's init() sets time.Local = Asia/Kolkata.
	_ = cfg

	logging.Infof("[INFO] Starting KineticOps Server...")
	logging.Infof("[DEBUG] JWTSecret loaded: %v", cfg.JWTSecret != "")

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
	handlers.InitWorkflowHandlers(workflowService)

	// Initialize circuit breakers for external services
	middleware.InitCircuitBreakers()
	logging.Infof("Circuit breakers initialized")

	// Initialize telemetry (OpenTelemetry) - returns shutdown func
	shutdownTelemetry := telemetry.InitTelemetry()
	defer shutdownTelemetry()

	// Elasticsearch integration removed per user request. No-op.

	// Start Prometheus + pprof on :9090 for metrics and profiling
	telemetry.StartPrometheusServer(":9090")

	// START METRIC COLLECTOR WORKER - ADD THIS
	workers.StartMetricCollector()

	// Start the metric batcher to improve ingestion throughput. Batches up to 500
	// metrics or flushes every 5 seconds.
	services.StartMetricBatcher(500, 5*time.Second)

	// Start enterprise retention service (30 days retention)
	retentionService := services.NewRetentionService(30)
	retentionService.StartRetentionWorker()

	// Start downsampling service for performance optimization
	downsamplingService := services.NewDownsamplingService()
	downsamplingService.StartDownsamplingWorker()

	// Start alert scheduler worker
	workers.StartAlertScheduler()

	// Start metric aggregation service for analytics
	services.MetricAggregationSvc.StartAggregationWorker(context.Background())

	// Start trend analysis service
	go services.TrendAnalysisSvc.StartTrendAnalysisWorker(context.Background())

	// Start agent health monitoring
	go services.AgentHealthSvc.CheckOfflineAgents(context.Background())

	// Start token rotation cleanup
	go services.TokenRotationSvc.CleanupExpiredTokens(context.Background())

	// Start enhanced retention worker
	enhancedRetentionWorker := workers.NewRetentionWorker(mongoClient)
	go enhancedRetentionWorker.Start(context.Background())
	logging.Infof("Enhanced retention worker started")

	// Set up Redpanda/Kafka
	// Read brokers from config (REDPANDA_BROKER). It may be a comma-separated
	// list like "redpanda:9092,other:9092". Fall back to localhost:9092 when
	// not configured to preserve previous behavior.
	var brokers []string
	if cfg.RedpandaBroker != "" {
		// split and trim
		for _, b := range strings.Split(cfg.RedpandaBroker, ",") {
			if t := strings.TrimSpace(b); t != "" {
				brokers = append(brokers, t)
			}
		}
	}
	if len(brokers) == 0 {
		// When running via docker-compose the Redpanda service is reachable
		// at the compose hostname `redpanda:9092`. Use that as the default
		// instead of localhost so services inside the compose network can
		// reach the broker by name.
		brokers = []string{"redpanda:9092"}
	}
	topic := "metrics-events"

	// Initialize Kafka/Redpanda producer with a few retries to tolerate broker startup ordering.
	var producerInitErr error
	for i := 0; i < 3; i++ {
		if _, err := kafkaevents.InitProducer(brokers, topic); err != nil {
			producerInitErr = err
			logging.Warnf("Redpanda producer init attempt %d failed: %v", i+1, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		producerInitErr = nil
		break
	}
	if producerInitErr != nil {
		logging.Warnf("[WARN] Redpanda/Kafka producer initialization failed after retries: %v", producerInitErr)
	} else {
		logging.Infof("Redpanda producer initialized")
	}

	// Start a reingest consumer that listens for failed metric batches and retries insertion
	workers.StartReingestConsumer(brokers, "metrics-failed", "kineticops-reingest")

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
		logging.Infof("[DEBUG] Processing Kafka message: %s", string(msg))
		// Broadcast as metric update
		wsHub.RememberMessage(msg)
		wsHub.Broadcast(msg)
	})

	// Session service available for session management
	_ = services.NewSessionService(redisrepo.Client)
	logging.Infof("Session service initialized")

	// Initialize health check handler
	healthHandler := handlers.NewHealthCheckHandler(redisrepo.Client, mongoClient)

	// Fiber app with error handler
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// Global middlewares - order matters!
	app.Use(middleware.SecurityHeaders())                     // Security headers (HSTS, CSP, etc.)
	app.Use(middleware.RequestID())                           // Add request ID first
	app.Use(middleware.PanicRecovery())                       // Recover from panics
	app.Use(middleware.ErrorLogger())                         // Log all errors
	app.Use(middleware.Logger())                              // HTTP request logging
	app.Use(middleware.CORS())                                // CORS headers
	app.Use(middleware.CSRFMiddleware)                        // CSRF protection
	app.Use(middleware.ValidateAndSanitizeInput())            // Input validation and XSS/SQL injection prevention
	app.Use(middleware.AdvancedRateLimiter(redisrepo.Client)) // Advanced rate limiting

	// Health check endpoints (before other routes)
	app.Get("/health", handlers.BasicHealthCheck)
	app.Get("/health/detailed", healthHandler.DetailedHealthCheck)
	app.Get("/health/ready", healthHandler.ReadinessCheck)
	app.Get("/health/live", handlers.LivenessCheck)

	// Register ALL routes through unified router
	routes.RegisterAllRoutes(app)

	// WebSocket JWT-enabled endpoint
	app.Get("/ws", websocket.New(ws.WsHandler(wsHub, cfg.JWTSecret)))

	// Serve frontend static files (if the frontend `dist/` exists). This allows
	// serving the SPA from the same origin as the API so browsers can connect
	// to the WebSocket without cross-origin issues. We only enable the SPA
	// fallback for non-API and non-WS routes.
	// Note: The frontend build output is expected at `frontend/dist` relative to
	// the repository root. If you deploy elsewhere, adapt the path or use a
	// reverse proxy.
	// Serve built frontend. When running the backend module from its own
	// directory the frontend build is located at ../frontend/dist. Use the
	// parent-relative path so both container and local dev runs resolve it.
	app.Static("/assets", "../frontend/dist/assets")

	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		// skip API and WS endpoints
		if strings.HasPrefix(path, "/api/") || path == "/ws" || strings.HasPrefix(path, "/api") {
			return c.Next()
		}
		// attempt to serve static file if exists
		if strings.HasPrefix(path, "/assets/") {
			return c.Next()
		}
		// otherwise, serve SPA index.html
		// Serve the SPA index from the frontend build output relative to the
		// repository root (one level up when running from backend/).
		return c.SendFile("../frontend/dist/index.html")
	})

	logging.Infof("✅ Server started successfully on port %s", cfg.AppPort)
	logging.Infof("📊 Health check: http://localhost:%s/health", cfg.AppPort)
	logging.Infof("🔌 API Base: http://localhost:%s/api/v1", cfg.AppPort)
	logging.Infof("🌍 Network: http://0.0.0.0:%s", cfg.AppPort)
	logging.Infof("⚙️  Metric Collector: Running every 60s")
	logging.Infof("🔒 Security: Headers, Rate Limiting, Input Validation, Circuit Breakers enabled")
	logging.Infof("💾 Retention: 30-day metrics, 30-day logs, 90-day audit logs")

	// Watch for SIGHUP to reload configuration at runtime without restart.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)
		for {
			<-sig
			logging.Infof("Received SIGHUP, reloading configuration...")
			config.Reload()
			logging.Infof("Configuration reload complete.")
		}
	}()

	// Start server in goroutine for graceful shutdown
	go func() {
		if err := app.Listen("0.0.0.0:" + cfg.AppPort); err != nil {
			logging.Errorf("server Listen error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Infof("Shutdown signal received, starting graceful shutdown...")

	// Graceful shutdown with 30 second timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := app.Shutdown(); err != nil {
		logging.Errorf("Error during server shutdown: %v", err)
	}

	// Close database connections
	if mongoClient != nil {
		if err := mongoClient.Disconnect(shutdownCtx); err != nil {
			logging.Errorf("Error disconnecting MongoDB: %v", err)
		} else {
			logging.Infof("MongoDB connection closed")
		}
	}

	if redisrepo.Client != nil {
		if err := redisrepo.Client.Close(); err != nil {
			logging.Errorf("Error closing Redis: %v", err)
		} else {
			logging.Infof("Redis connection closed")
		}
	}

	if postgres.SqlxDB != nil {
		if err := postgres.SqlxDB.Close(); err != nil {
			logging.Errorf("Error closing PostgreSQL: %v", err)
		} else {
			logging.Infof("PostgreSQL connection closed")
		}
	}

	// Flush logs
	logging.Flush()

	logging.Infof("Graceful shutdown completed. Goodbye!")
}
