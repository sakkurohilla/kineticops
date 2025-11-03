package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sakkurohilla/kineticops/agent/cmd"
	"github.com/sakkurohilla/kineticops/agent/config"
	"github.com/sakkurohilla/kineticops/agent/utils"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	var (
		configPath = flag.String("c", "/etc/kineticops-agent/config.yaml", "Configuration file path")
		showVersion = flag.Bool("version", false, "Show version information")
		testConfig = flag.Bool("test", false, "Test configuration and exit")
		verbose    = flag.Bool("v", false, "Verbose logging")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("KineticOps Agent %s\n", version)
		fmt.Printf("Build time: %s\n", buildTime)
		fmt.Printf("Git commit: %s\n", gitCommit)
		os.Exit(0)
	}

	// Initialize logger
	logger := utils.NewLogger(*verbose)
	logger.Info("Starting KineticOps Agent", "version", version)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	if *testConfig {
		logger.Info("Configuration test passed")
		os.Exit(0)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create and start agent
	agent, err := cmd.NewAgent(cfg, logger)
	if err != nil {
		logger.Error("Failed to create agent", "error", err)
		os.Exit(1)
	}

	// Start agent in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- agent.Run(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", "signal", sig)
		cancel()
		
		// Wait for graceful shutdown with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		
		if err := agent.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error during shutdown", "error", err)
			os.Exit(1)
		}
		
		logger.Info("Agent stopped gracefully")
		
	case err := <-errChan:
		if err != nil {
			logger.Error("Agent error", "error", err)
			os.Exit(1)
		}
	}
}