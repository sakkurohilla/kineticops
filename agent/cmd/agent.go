package cmd

import (
	"context"

	"github.com/sakkurohilla/kineticops/agent/config"
	"github.com/sakkurohilla/kineticops/agent/modules/logs"
	"github.com/sakkurohilla/kineticops/agent/modules/metrics"
	"github.com/sakkurohilla/kineticops/agent/outputs"
	"github.com/sakkurohilla/kineticops/agent/pipelines"
	"github.com/sakkurohilla/kineticops/agent/state"
	"github.com/sakkurohilla/kineticops/agent/utils"
)

type Agent struct {
	config   *config.Config
	logger   *utils.Logger
	pipeline *pipelines.PipelineManager
	modules  []Module
	stateMgr *state.Manager
}

type Module interface {
	Name() string
	IsEnabled() bool
	Start(ctx context.Context) error
	Stop() error
}

func NewAgent(cfg *config.Config, logger *utils.Logger) (*Agent, error) {
	// Create output
	output, err := outputs.NewKineticOpsOutput(&cfg.Output.KineticOps, logger)
	if err != nil {
		return nil, err
	}

	// Create pipeline (batching controlled by agent config)
	pipeline := pipelines.NewPipelineManager(output, logger, cfg.Agent.BatchSize, cfg.Agent.BatchTime)

	// Create state manager for modules that need persistent offsets
	stateDir := "/var/lib/kineticops-agent/state"
	stateMgr, err := state.NewManager(stateDir)
	if err != nil {
		return nil, err
	}

	// Create modules
	var modules []Module

	// Debug: log whether logs module parsed from config
	if logger != nil {
		logger.Info("Logs module configured", "enabled", cfg.Modules.Logs.Enabled, "inputs", len(cfg.Modules.Logs.Inputs))
	}

	// System metrics module
	if cfg.Modules.System.Enabled {
		systemModule, err := metrics.NewSystemModule(&cfg.Modules.System, pipeline, logger)
		if err != nil {
			return nil, err
		}
		modules = append(modules, systemModule)
	}

	// Logs module (file tailing)
	if cfg.Modules.Logs.Enabled {
		logsModule, err := logs.NewLogsModule(&cfg.Modules.Logs, pipeline, stateMgr, logger)
		if err != nil {
			return nil, err
		}
		modules = append(modules, logsModule)
	}

	return &Agent{
		config:   cfg,
		logger:   logger,
		pipeline: pipeline,
		modules:  modules,
		stateMgr: stateMgr,
	}, nil
}

func (a *Agent) Run(ctx context.Context) error {
	a.logger.Info("Starting KineticOps Agent")

	// Start pipeline
	if err := a.pipeline.Start(ctx); err != nil {
		return err
	}

	// Start all modules
	for _, module := range a.modules {
		if module.IsEnabled() {
			a.logger.Info("Starting module", "name", module.Name())
			go func(m Module) {
				if err := m.Start(ctx); err != nil {
					a.logger.Error("Module error", "name", m.Name(), "error", err)
				}
			}(module)
		}
	}

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

func (a *Agent) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down agent")

	// Stop all modules
	for _, module := range a.modules {
		if err := module.Stop(); err != nil {
			a.logger.Error("Error stopping module", "name", module.Name(), "error", err)
		}
	}

	// Stop pipeline
	return a.pipeline.Stop()
}
