package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/toolchain"
)

// PrometheusExporter exposes Lango metrics in Prometheus exposition format.
// It subscribes to EventBus events and updates counters/gauges directly.
type PrometheusExporter struct {
	registry  *prometheus.Registry
	collector *MetricsCollector

	tokenUsage      *prometheus.CounterVec
	toolExecutions  *prometheus.CounterVec
	toolDuration    *prometheus.HistogramVec
	policyDecisions *prometheus.CounterVec
	trackedSessions prometheus.Gauge
}

// NewPrometheusExporter creates a new exporter with registered metrics.
func NewPrometheusExporter() *PrometheusExporter {
	reg := prometheus.NewRegistry()

	e := &PrometheusExporter{
		registry: reg,
		tokenUsage: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "lango_token_usage_total",
			Help: "Total token usage by type (input, output, cache).",
		}, []string{"type"}),
		toolExecutions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "lango_tool_executions_total",
			Help: "Total tool executions by tool name and success status.",
		}, []string{"tool", "success"}),
		toolDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "lango_tool_duration_seconds",
			Help:    "Tool execution duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"tool"}),
		policyDecisions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "lango_policy_decisions_total",
			Help: "Total policy decisions by verdict (allow, observe, block).",
		}, []string{"verdict"}),
		trackedSessions: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "lango_tracked_sessions",
			Help: "Number of sessions currently tracked by the metrics collector.",
		}),
	}

	reg.MustRegister(e.tokenUsage)
	reg.MustRegister(e.toolExecutions)
	reg.MustRegister(e.toolDuration)
	reg.MustRegister(e.policyDecisions)
	reg.MustRegister(e.trackedSessions)

	return e
}

// Subscribe wires the exporter to receive events from the EventBus.
func (e *PrometheusExporter) Subscribe(bus *eventbus.Bus) {
	eventbus.SubscribeTyped[toolchain.ToolExecutedEvent](bus, func(evt toolchain.ToolExecutedEvent) {
		success := "true"
		if !evt.Success {
			success = "false"
		}
		e.toolExecutions.WithLabelValues(evt.ToolName, success).Inc()
		e.toolDuration.WithLabelValues(evt.ToolName).Observe(evt.Duration.Seconds())
	})

	eventbus.SubscribeTyped[eventbus.PolicyDecisionEvent](bus, func(evt eventbus.PolicyDecisionEvent) {
		e.policyDecisions.WithLabelValues(evt.Verdict).Inc()
	})

	eventbus.SubscribeTyped[eventbus.TokenUsageEvent](bus, func(evt eventbus.TokenUsageEvent) {
		e.tokenUsage.WithLabelValues("input").Add(float64(evt.InputTokens))
		e.tokenUsage.WithLabelValues("output").Add(float64(evt.OutputTokens))
		if evt.CacheTokens > 0 {
			e.tokenUsage.WithLabelValues("cache").Add(float64(evt.CacheTokens))
		}
		e.updateSessionGauge()
	})
}

// SetCollector links the exporter to a MetricsCollector so it can
// update the tracked-sessions gauge from token usage events.
func (e *PrometheusExporter) SetCollector(c *MetricsCollector) {
	e.collector = c
}

// updateSessionGauge refreshes the tracked sessions gauge from the collector.
func (e *PrometheusExporter) updateSessionGauge() {
	if e.collector == nil {
		return
	}
	snap := e.collector.Snapshot()
	e.trackedSessions.Set(float64(len(snap.SessionBreakdown)))
}

// Handler returns an HTTP handler for the Prometheus exposition endpoint.
func (e *PrometheusExporter) Handler() http.Handler {
	return promhttp.HandlerFor(e.registry, promhttp.HandlerOpts{})
}
