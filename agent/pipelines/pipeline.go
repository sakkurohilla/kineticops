package pipelines

import (
	"context"
	"sync"
	"time"

	"github.com/sakkurohilla/kineticops/agent/outputs"
	"github.com/sakkurohilla/kineticops/agent/utils"
)

// PipelineManager manages the data pipeline
type PipelineManager struct {
	output     outputs.Output
	logger     *utils.Logger
	eventChan  chan map[string]interface{}
	batchSize  int
	batchTime  time.Duration
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// NewPipelineManager creates a new pipeline manager
func NewPipelineManager(output outputs.Output, logger *utils.Logger) *PipelineManager {
	return &PipelineManager{
		output:     output,
		logger:     logger,
		eventChan:  make(chan map[string]interface{}, 1000),
		batchSize:  100,
		batchTime:  5 * time.Second,
		stopChan:   make(chan struct{}),
	}
}

// Start starts the pipeline
func (p *PipelineManager) Start(ctx context.Context) error {
	p.logger.Info("Starting pipeline", "batch_size", p.batchSize, "batch_time", p.batchTime)

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.processBatches(ctx)
	}()

	return nil
}

// Stop stops the pipeline
func (p *PipelineManager) Stop() error {
	p.logger.Info("Stopping pipeline")
	close(p.stopChan)
	p.wg.Wait()
	return p.output.Close()
}

// Send sends an event through the pipeline
func (p *PipelineManager) Send(event map[string]interface{}) error {
	select {
	case p.eventChan <- event:
		return nil
	default:
		p.logger.Warn("Event channel full, dropping event")
		return nil
	}
}

// processBatches processes events in batches
func (p *PipelineManager) processBatches(ctx context.Context) {
	ticker := time.NewTicker(p.batchTime)
	defer ticker.Stop()

	var batch []map[string]interface{}

	for {
		select {
		case <-ctx.Done():
			// Send remaining events
			if len(batch) > 0 {
				p.sendBatch(batch)
			}
			return

		case <-p.stopChan:
			// Send remaining events
			if len(batch) > 0 {
				p.sendBatch(batch)
			}
			return

		case event := <-p.eventChan:
			batch = append(batch, event)
			
			// Send batch if it's full
			if len(batch) >= p.batchSize {
				p.sendBatch(batch)
				batch = nil
			}

		case <-ticker.C:
			// Send batch on timer
			if len(batch) > 0 {
				p.sendBatch(batch)
				batch = nil
			}
		}
	}
}

// sendBatch sends a batch of events
func (p *PipelineManager) sendBatch(batch []map[string]interface{}) {
	if len(batch) == 0 {
		return
	}

	p.logger.Info("Sending batch", "size", len(batch))

	// Process events (add common fields, etc.)
	processedBatch := p.processEvents(batch)

	// Send to output
	if err := p.output.Send(processedBatch); err != nil {
		p.logger.Error("Failed to send batch", "size", len(batch), "error", err)
		return
	}

	p.logger.Info("Successfully sent batch", "size", len(batch))
}

// processEvents processes events before sending
func (p *PipelineManager) processEvents(events []map[string]interface{}) []map[string]interface{} {
	processed := make([]map[string]interface{}, len(events))
	
	for i, event := range events {
		// Create a copy to avoid modifying original
		processedEvent := make(map[string]interface{})
		for k, v := range event {
			processedEvent[k] = v
		}

		// Add common fields
		if _, exists := processedEvent["@timestamp"]; !exists {
			processedEvent["@timestamp"] = time.Now().UTC().Format(time.RFC3339)
		}

		// Add ECS fields
		if _, exists := processedEvent["ecs"]; !exists {
			processedEvent["ecs"] = map[string]interface{}{
				"version": "1.12.0",
			}
		}

		processed[i] = processedEvent
	}

	return processed
}