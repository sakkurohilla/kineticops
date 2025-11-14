package telemetry

import (
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/sakkurohilla/kineticops/backend/internal/logging"
)

var (
	broadcastQueueDrops  prometheus.Counter
	clientSendDrops      prometheus.Counter
	clientDisconnects    prometheus.Counter
	hubQueueLength       prometheus.Gauge
	totalClients         prometheus.Gauge
	ingestionQueueLength prometheus.Gauge
)

func initPromMetrics() {
	broadcastQueueDrops = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ws_broadcast_queue_drops_total",
		Help: "Number of messages dropped because the broadcast queue was full",
	})
	clientSendDrops = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ws_client_send_drops_total",
		Help: "Number of messages dropped for a client due to rate limiting",
	})
	clientDisconnects = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ws_client_disconnects_total",
		Help: "Number of client disconnects due to blocked send queue",
	})
	hubQueueLength = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ws_hub_broadcast_queue_length",
		Help: "Current length of the hub broadcast queue",
	})
	totalClients = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ws_total_clients",
		Help: "Current number of connected websocket clients",
	})
	ingestionQueueLength = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "metrics_ingestion_queue_length",
		Help: "Number of metrics currently buffered in the ingestion batcher",
	})

	prometheus.MustRegister(broadcastQueueDrops, clientSendDrops, clientDisconnects, hubQueueLength, totalClients)
	prometheus.MustRegister(ingestionQueueLength)
}

// StartPrometheusServer starts a separate HTTP server exposing /metrics and pprof endpoints.
// It is safe to call multiple times (only first call starts the server).
func StartPrometheusServer(addr string) {
	initPromMetrics()
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		// register pprof endpoints explicitly
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		server := &http.Server{
			Addr:    addr,
			Handler: mux,
			// keep timeouts short
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		logging.Infof("[METRICS] starting Prometheus/pprof server on %s", addr)
		if err := server.ListenAndServe(); err != nil {
			logging.Errorf("[METRICS] metrics server exited: %v", err)
		}
	}()
}

func IncBroadcastQueueDrop() {
	if broadcastQueueDrops != nil {
		broadcastQueueDrops.Inc()
	}
}

func IncClientSendDrop() {
	if clientSendDrops != nil {
		clientSendDrops.Inc()
	}
}

func IncClientDisconnect() {
	if clientDisconnects != nil {
		clientDisconnects.Inc()
	}
}

func SetHubQueueLength(n int) {
	if hubQueueLength != nil {
		hubQueueLength.Set(float64(n))
	}
}

func SetTotalClients(n int) {
	if totalClients != nil {
		totalClients.Set(float64(n))
	}
}

// SetMetricIngestionQueueLength sets the current size of the metric ingestion buffer
func SetMetricIngestionQueueLength(n int) {
	if ingestionQueueLength != nil {
		ingestionQueueLength.Set(float64(n))
	}
}
