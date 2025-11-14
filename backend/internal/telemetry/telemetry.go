package telemetry

import (
	"context"
	"sync/atomic"

	"github.com/sakkurohilla/kineticops/backend/internal/logging"
)

var seq uint64

// NextSeq returns a monotonic increasing sequence id for metric events.
func NextSeq() uint64 {
	return atomic.AddUint64(&seq, 1)
}

// LastSeq returns the current last sequence id (atomic read).
func LastSeq() uint64 {
	return atomic.LoadUint64(&seq)
}

var (
	collectionSuccess uint64
	collectionError   uint64
	kafkaPublish      uint64
	wsBroadcast       uint64
	wsSendErrors      uint64
)

// InitTelemetry configures a basic OpenTelemetry tracer and metric counters.
// For now it uses stdout trace exporter (use OTLP/Prometheus in production).
func InitTelemetry() func() {
	// Lightweight noop telemetry init. We use atomic counters and a
	// monotonic sequence id to help ordering and basic metrics. In
	// production this can be replaced with an OTLP/Prometheus exporter.
	logging.Infof("[TELEMETRY] initialized (noop) - using in-memory counters")
	return func() {
		// nothing to shutdown for noop telemetry
		logging.Infof("[TELEMETRY] shutdown")
	}
}

// Instrumentation helpers
func IncCollectionSuccess(ctx context.Context, v int64) {
	atomic.AddUint64(&collectionSuccess, uint64(v))
}

func IncCollectionError(ctx context.Context, v int64) {
	atomic.AddUint64(&collectionError, uint64(v))
}

func IncKafkaPublish(ctx context.Context, v int64) {
	atomic.AddUint64(&kafkaPublish, uint64(v))
}

func IncWSBroadcast(ctx context.Context, v int64) {
	atomic.AddUint64(&wsBroadcast, uint64(v))
}

func IncWSSendErrors(ctx context.Context, v int64) {
	atomic.AddUint64(&wsSendErrors, uint64(v))
}

// GetCounters returns a snapshot of in-memory telemetry counters.
func GetCounters() map[string]uint64 {
	return map[string]uint64{
		"collection_success_total": atomic.LoadUint64(&collectionSuccess),
		"collection_error_total":   atomic.LoadUint64(&collectionError),
		"kafka_publish_total":      atomic.LoadUint64(&kafkaPublish),
		"ws_broadcast_total":       atomic.LoadUint64(&wsBroadcast),
		"ws_send_errors_total":     atomic.LoadUint64(&wsSendErrors),
		"last_seq":                 atomic.LoadUint64(&seq),
	}
}

// StartTraceSpan returns a tracer span for lightweight local tracing
// No-op trace span helper in this lightweight implementation.
