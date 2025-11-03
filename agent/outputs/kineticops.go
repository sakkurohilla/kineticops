package outputs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sakkurohilla/kineticops/agent/config"
	"github.com/sakkurohilla/kineticops/agent/utils"
)

// Output interface for all output types
type Output interface {
	Send(events []map[string]interface{}) error
	Close() error
}

// KineticOpsOutput sends data to KineticOps backend
type KineticOpsOutput struct {
	config *config.KineticOpsOutput
	logger *utils.Logger
	client *http.Client
}

// NewKineticOpsOutput creates a new KineticOps output
func NewKineticOpsOutput(cfg *config.KineticOpsOutput, logger *utils.Logger) (*KineticOpsOutput, error) {
	client := &http.Client{
		Timeout: cfg.Timeout,
	}

	return &KineticOpsOutput{
		config: cfg,
		logger: logger,
		client: client,
	}, nil
}

// Send sends events to KineticOps backend
func (k *KineticOpsOutput) Send(events []map[string]interface{}) error {
	if len(events) == 0 {
		return nil
	}

	// Try each host until one succeeds
	var lastErr error
	for _, host := range k.config.Hosts {
		if err := k.sendToHost(host, events); err != nil {
			k.logger.Error("Failed to send to host", "host", host, "error", err)
			lastErr = err
			continue
		}
		return nil
	}

	return fmt.Errorf("failed to send to any host: %w", lastErr)
}

// sendToHost sends events to a specific host
func (k *KineticOpsOutput) sendToHost(host string, events []map[string]interface{}) error {
	// Determine endpoint based on event type
	endpoint := k.getEndpoint(host, events[0])
	
	// Prepare payload
	payload, err := json.Marshal(map[string]interface{}{
		"events": events,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if k.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+k.config.Token)
	}

	// Send request with retries
	return k.sendWithRetry(req)
}

// getEndpoint determines the correct endpoint based on event type
func (k *KineticOpsOutput) getEndpoint(host string, event map[string]interface{}) string {
	// Check event kind to determine endpoint
	if eventData, ok := event["event"].(map[string]interface{}); ok {
		if kind, ok := eventData["kind"].(string); ok {
			switch kind {
			case "metric":
				return fmt.Sprintf("%s/api/v1/metrics/collect", host)
			case "event":
				return fmt.Sprintf("%s/api/v1/logs/collect", host)
			}
		}
	}

	// Default to metrics endpoint
	return fmt.Sprintf("%s/api/v1/metrics/collect", host)
}

// sendWithRetry sends request with retry logic
func (k *KineticOpsOutput) sendWithRetry(req *http.Request) error {
	var lastErr error
	
	for attempt := 0; attempt <= k.config.MaxRetry; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * time.Second
			k.logger.Debug("Retrying request", "attempt", attempt, "backoff", backoff)
			time.Sleep(backoff)
		}

		resp, err := k.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			k.logger.Debug("Successfully sent events", "status", resp.StatusCode)
			return nil
		}

		lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		
		// Don't retry on client errors (4xx)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			break
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", k.config.MaxRetry+1, lastErr)
}

// Close closes the output
func (k *KineticOpsOutput) Close() error {
	// Nothing to close for HTTP client
	return nil
}