package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	lastStateChange time.Time

	// Configuration
	maxFailures      int           // Max failures before opening
	timeout          time.Duration // Time to wait before half-open
	successThreshold int           // Successes needed in half-open to close
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration, successThreshold int) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		maxFailures:      maxFailures,
		timeout:          timeout,
		successThreshold: successThreshold,
		lastStateChange:  time.Now(),
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	// Check if we should transition from Open to Half-Open
	if cb.state == StateOpen && time.Since(cb.lastStateChange) > cb.timeout {
		logging.Infof("Circuit breaker transitioning from Open to Half-Open")
		cb.state = StateHalfOpen
		cb.successCount = 0
		cb.lastStateChange = time.Now()
	}

	// Don't allow calls if circuit is open
	if cb.state == StateOpen {
		cb.mu.Unlock()
		return fiber.NewError(fiber.StatusServiceUnavailable, "Service temporarily unavailable (circuit breaker open)")
	}

	cb.mu.Unlock()

	// Execute the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

func (cb *CircuitBreaker) onSuccess() {
	cb.failureCount = 0

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			logging.Infof("Circuit breaker transitioning from Half-Open to Closed")
			cb.state = StateClosed
			cb.successCount = 0
			cb.lastStateChange = time.Now()
		}
	}
}

func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == StateHalfOpen {
		logging.Warnf("Circuit breaker transitioning from Half-Open to Open")
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
		return
	}

	if cb.failureCount >= cb.maxFailures {
		logging.Warnf("Circuit breaker transitioning from Closed to Open (failures: %d)", cb.failureCount)
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns current circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	var stateStr string
	switch cb.state {
	case StateOpen:
		stateStr = "open"
	case StateHalfOpen:
		stateStr = "half-open"
	default:
		stateStr = "closed"
	}

	return map[string]interface{}{
		"state":             stateStr,
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"last_failure_time": cb.lastFailureTime,
		"last_state_change": cb.lastStateChange,
		"time_in_state":     time.Since(cb.lastStateChange).Seconds(),
	}
}

// Global circuit breakers for different services
var (
	MongoDBCircuitBreaker  *CircuitBreaker
	RedisCircuitBreaker    *CircuitBreaker
	RedpandaCircuitBreaker *CircuitBreaker
)

func InitCircuitBreakers() {
	// MongoDB: 5 failures, 30s timeout, 2 successes to close
	MongoDBCircuitBreaker = NewCircuitBreaker(5, 30*time.Second, 2)

	// Redis: 10 failures, 10s timeout, 3 successes to close
	RedisCircuitBreaker = NewCircuitBreaker(10, 10*time.Second, 3)

	// Redpanda: 5 failures, 30s timeout, 2 successes to close
	RedpandaCircuitBreaker = NewCircuitBreaker(5, 30*time.Second, 2)

	logging.Infof("Circuit breakers initialized for MongoDB, Redis, and Redpanda")
}
