package cmd

import (
	"context"

	"github.com/sakkurohilla/kineticops/agent/config"
	"github.com/sakkurohilla/kineticops/agent/modules/metrics"
	"github.com/sakkurohilla/kineticops/agent/outputs"
	"github.com/sakkurohilla/kineticops/agent/pipelines"
	"github.com/sakkurohilla/kineticops/agent/utils"
)

type Agent struct {
	config   *config.Config
	logger   *utils.Logger
	pipeline *pipelines.PipelineManager
	modules  []Module
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

	// Create pipeline
	pipeline := pipelines.NewPipelineManager(output, logger)

	// Create modules
	var modules []Module
	
	// System metrics module
	if cfg.Modules.System.Enabled {
		systemModule, err := metrics.NewSystemModule(&cfg.Modules.System, pipeline, logger)
		if err != nil {
			return nil, err
		}
		modules = append(modules, systemModule)
	}

	return &Agent{
		config:   cfg,
		logger:   logger,
		pipeline: pipeline,
		modules:  modules,
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