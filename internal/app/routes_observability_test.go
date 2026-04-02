package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/observability/health"
)

// stubChecker is a health.Checker that returns a fixed status.
type stubChecker struct {
	name   string
	status health.Status
	msg    string
}

func (s *stubChecker) Name() string { return s.name }
func (s *stubChecker) Check(_ context.Context) health.ComponentHealth {
	return health.ComponentHealth{
		Name:        s.name,
		Status:      s.status,
		Message:     s.msg,
		LastChecked: time.Now(),
	}
}

func TestHealthDetailed_AllHealthy(t *testing.T) {
	t.Parallel()

	r := chi.NewRouter()
	collector := observability.NewCollector()
	hr := health.NewRegistry()
	hr.Register(&stubChecker{name: "memory", status: health.StatusHealthy, msg: "ok"})
	hr.Register(&stubChecker{name: "db", status: health.StatusHealthy, msg: "connected"})

	registerObservabilityRoutes(r, collector, hr, nil, nil)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/detailed")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body health.SystemHealth
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, health.StatusHealthy, body.Status)
	assert.Len(t, body.Components, 2)
}

func TestHealthDetailed_DegradedComponent(t *testing.T) {
	t.Parallel()

	r := chi.NewRouter()
	collector := observability.NewCollector()
	hr := health.NewRegistry()
	hr.Register(&stubChecker{name: "memory", status: health.StatusHealthy})
	hr.Register(&stubChecker{name: "provider", status: health.StatusDegraded, msg: "timeout"})

	registerObservabilityRoutes(r, collector, hr, nil, nil)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/detailed")
	require.NoError(t, err)
	defer resp.Body.Close()

	var body health.SystemHealth
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, health.StatusDegraded, body.Status, "worst-status should be degraded")
}

func TestHealthDetailed_NilRegistry_NoRoute(t *testing.T) {
	t.Parallel()

	r := chi.NewRouter()
	collector := observability.NewCollector()

	// hr = nil → /health/detailed should not be registered.
	registerObservabilityRoutes(r, collector, nil, nil, nil)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health/detailed")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMetrics_ReturnsSnapshot(t *testing.T) {
	t.Parallel()

	r := chi.NewRouter()
	collector := observability.NewCollector()

	registerObservabilityRoutes(r, collector, nil, nil, nil)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Contains(t, body, "uptime")
	assert.Contains(t, body, "tokenUsage")
	assert.Contains(t, body, "toolExecutions")
}
