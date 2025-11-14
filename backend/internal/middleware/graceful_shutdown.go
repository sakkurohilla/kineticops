package middleware

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
)

// GracefulShutdown handles the graceful shutdown of the server.
type GracefulShutdown struct {
	app      *fiber.App
	shutdown chan struct{}
}

// NewGracefulShutdown creates a new GracefulShutdown instance.
func NewGracefulShutdown(app *fiber.App) *GracefulShutdown {
	return &GracefulShutdown{
		app:      app,
		shutdown: make(chan struct{}),
	}
}

// Shutdown starts the graceful shutdown process.
func (gs *GracefulShutdown) Shutdown(ctx context.Context) {
	go func() {
		defer close(gs.shutdown)

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

		select {
		case <-signalChan:
			logging.Infof("Received shutdown signal. Gracefully shutting down...")
		case <-ctx.Done():
			logging.Infof("Context is done. Gracefully shutting down...")
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := gs.app.ShutdownWithContext(shutdownCtx); err != nil {
			logging.Errorf("Error during server shutdown: %v", err)
		}
	}()
}

// Wait blocks until the shutdown process is complete.
func (gs *GracefulShutdown) Wait() {
	<-gs.shutdown
	logging.Infof("Shutdown complete.")
}
