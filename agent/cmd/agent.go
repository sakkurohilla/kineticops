package cmd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sakkurohilla/kineticops/agent/config"
	"github.com/sakkurohilla/kineticops/agent/modules/metrics"
	"github.com/sakkurohilla/kineticops/agent/modules/logs"
	"github.com/sakkurohilla/kineticops/agent/outputs"
	"github.com/sakkurohilla/kineticops/agent/pipelines"
	"github.com/sakkurohilla/kineticops/agent/state"
	"github.com/sakkurohilla/kineticops/agent/utils"
)

// Agent represents the main agent instance
type Agent struct {
	config   *config.Config
	logger   *utils.Logger
	pipeline *pipelines.PipelineManager
	state    *state.Manager
	modules  []Module
	output   outputs.Output
	
	// Control channels
	stopChan chan struct{}
	doneChan chan struct{}
	wg       sync.WaitGroup
}

// Module interface for all data collection modules
type Module interface {
	Name() string
	Start(ctx context.Context) error
	Stop() error
	IsEnabled() bool
}

// NewAgent creates a new agent instance
func NewAgent(cfg *config.Config, logger *utils.Logger) (*Agent, error) {
	// Create state manager
	stateManager, err := state.NewManager("/var/lib/kineticops-agent/state")
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	// Create output
	output, err := outputs.NewKineticOpsOutput(&cfg.Output.KineticOps, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create output: %w", err)
	}

	// Create pipeline manager
	pipelineManager := pipelines.NewPipelineManager(output, logger)

	agent := &Agent{
		config:   cfg,
		logger:   logger,
		pipeline: pipelineManager,
		state:    stateManager,
		output:   output,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}

	// Initialize modules
	if err := agent.initModules(); err != nil {
		return nil, fmt.Errorf("failed to initialize modules: %w", err)
	}

	return agent, nil
}

// initModules initializes all enabled modules
func (a *Agent) initModules() error {
	var modules []Module

	// System metrics module
	if a.config.Modules.System.Enabled {
		systemModule, err := metrics.NewSystemModule(&a.config.Modules.System, a.pipeline, a.logger)
		if err != nil {
			return fmt.Errorf("failed to create system module: %w", err)
		}
		modules = append(modules, systemModule)
	}

	// Logs module
	if a.config.Modules.Logs.Enabled {
		logsModule, err := logs.NewLogsModule(&a.config.Modules.Logs, a.pipeline, a.state, a.logger)
		if err != nil {
			return fmt.Errorf("failed to create logs module: %w", err)
		}
		modules = append(modules, logsModule)
	}

	a.modules = modules
	a.logger.Info("Initialized modules", "count", len(modules))
	
	return nil
}

// Run starts the agent and all its modules
func (a *Agent) Run(ctx context.Context) error {
	a.logger.Info("Starting KineticOps Agent", "hostname", a.config.Agent.Hostname)

	// Start pipeline manager
	if err := a.pipeline.Start(ctx); err != nil {
		return fmt.Errorf("failed to start pipeline: %w", err)
	}

	// Start all modules
	for _, module := range a.modules {
		if module.IsEnabled() {
			a.wg.Add(1)
			go func(m Module) {
				defer a.wg.Done()
				a.logger.Info("Starting module", "name", m.Name())
				
				if err := m.Start(ctx); err != nil {
					a.logger.Error("Module failed", "name", m.Name(), "error", err)
				}
			}(module)
		}
	}

	// Start heartbeat
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.runHeartbeat(ctx)
	}()

	// Wait for context cancellation
	<-ctx.Done()
	a.logger.Info("Agent stopping...")

	close(a.doneChan)
	return nil
}

// Shutdown gracefully stops the agent
func (a *Agent) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down agent...")

	// Stop all modules
	for _, module := range a.modules {
		if err := module.Stop(); err != nil {
			a.logger.Error("Error stopping module", "name", module.Name(), "error", err)
		}
	}

	// Stop pipeline
	if err := a.pipeline.Stop(); err != nil {
		a.logger.Error("Error stopping pipeline", "error", err)
	}

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("All modules stopped gracefully")
	case <-ctx.Done():
		a.logger.Warn("Shutdown timeout reached, forcing exit")
	}

	return nil
}

// runHeartbeat sends periodic heartbeats to the backend
func (a *Agent) runHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(a.config.Agent.Period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends a heartbeat message
func (a *Agent) sendHeartbeat() {
	heartbeat := map[string]interface{}{
		"@timestamp": time.Now().UTC().Format(time.RFC3339),
		"agent": map[string]interface{}{
			"name":     a.config.Agent.Name,
			"hostname": a.config.Agent.Hostname,
			"version":  "1.0.0",
		},
		"host": map[string]interface{}{
			"hostname": a.config.Agent.Hostname,
		},
		"event": map[string]interface{}{
			"kind":     "metric",
			"category": "host",
			"type":     "info",
		},
		"message": "Agent heartbeat",
	}

	if err := a.pipeline.Send(heartbeat); err != nil {
		a.logger.Error("Failed to send heartbeat", "error", err)
	}
}