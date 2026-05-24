package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/observability/health"
	"github.com/langoai/lango/internal/storage"
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

	registerObservabilityRoutes(r, collector, hr, nil, nil, nil)

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

	registerObservabilityRoutes(r, collector, hr, nil, nil, nil)

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

	// A nil health registry should leave /health/detailed unregistered.
	registerObservabilityRoutes(r, collector, nil, nil, nil, nil)

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

	registerObservabilityRoutes(r, collector, nil, nil, nil, nil)

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

func TestObservabilityRoutes_NilCollectorRegistersNoRoutes(t *testing.T) {
	t.Parallel()

	r := chi.NewRouter()
	registerObservabilityRoutes(r, nil, nil, nil, func(context.Context, time.Time) ([]storage.AlertRecord, error) {
		t.Fatal("alerts reader should not be registered when collector is nil")
		return nil, nil
	}, nil)

	for _, path := range []string{"/metrics", "/metrics/sessions", "/alerts"} {
		rec := performObservabilityRequest(r, path)
		assert.Equal(t, http.StatusNotFound, rec.Code, path)
	}
}

func TestObservabilityRoutes_MetricsSnapshotRoutes(t *testing.T) {
	t.Parallel()

	r := chi.NewRouter()
	collector := observability.NewCollector()
	collector.RecordTokenUsage(observability.TokenUsage{
		SessionKey:   "session-1",
		AgentName:    "agent-1",
		InputTokens:  11,
		OutputTokens: 7,
		TotalTokens:  18,
		CacheTokens:  3,
	})
	collector.RecordToolExecution("search", "agent-1", 150*time.Millisecond, true)
	collector.RecordToolExecution("search", "agent-1", 50*time.Millisecond, false)
	collector.RecordPolicyDecision("block", "policy-deny")
	collector.RecordPolicyDecision("observe", "policy-observe")

	registerObservabilityRoutes(r, collector, nil, nil, nil, nil)

	t.Run("summary", func(t *testing.T) {
		rec := performObservabilityRequest(r, "/metrics")

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var body struct {
			ToolExecutions int64 `json:"toolExecutions"`
			TokenUsage     struct {
				InputTokens  int64 `json:"inputTokens"`
				OutputTokens int64 `json:"outputTokens"`
				TotalTokens  int64 `json:"totalTokens"`
				CacheTokens  int64 `json:"cacheTokens"`
			} `json:"tokenUsage"`
			SessionCount int `json:"sessionCount"`
			AgentCount   int `json:"agentCount"`
			ToolCount    int `json:"toolCount"`
		}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		assert.Equal(t, int64(2), body.ToolExecutions)
		assert.Equal(t, int64(11), body.TokenUsage.InputTokens)
		assert.Equal(t, int64(7), body.TokenUsage.OutputTokens)
		assert.Equal(t, int64(18), body.TokenUsage.TotalTokens)
		assert.Equal(t, int64(3), body.TokenUsage.CacheTokens)
		assert.Equal(t, 1, body.SessionCount)
		assert.Equal(t, 1, body.AgentCount)
		assert.Equal(t, 1, body.ToolCount)
	})

	t.Run("sessions", func(t *testing.T) {
		rec := performObservabilityRequest(r, "/metrics/sessions")

		assert.Equal(t, http.StatusOK, rec.Code)

		var body struct {
			Sessions []struct {
				SessionKey   string `json:"sessionKey"`
				InputTokens  int64  `json:"inputTokens"`
				OutputTokens int64  `json:"outputTokens"`
				TotalTokens  int64  `json:"totalTokens"`
				RequestCount int64  `json:"requestCount"`
			} `json:"sessions"`
		}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		require.Len(t, body.Sessions, 1)
		assert.Equal(t, "session-1", body.Sessions[0].SessionKey)
		assert.Equal(t, int64(11), body.Sessions[0].InputTokens)
		assert.Equal(t, int64(7), body.Sessions[0].OutputTokens)
		assert.Equal(t, int64(18), body.Sessions[0].TotalTokens)
		assert.Equal(t, int64(1), body.Sessions[0].RequestCount)
	})

	t.Run("tools", func(t *testing.T) {
		rec := performObservabilityRequest(r, "/metrics/tools")

		assert.Equal(t, http.StatusOK, rec.Code)

		var body struct {
			Tools []struct {
				Name        string  `json:"name"`
				Count       int64   `json:"count"`
				Errors      int64   `json:"errors"`
				AvgDuration string  `json:"avgDuration"`
				ErrorRate   float64 `json:"errorRate"`
			} `json:"tools"`
		}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		require.Len(t, body.Tools, 1)
		assert.Equal(t, "search", body.Tools[0].Name)
		assert.Equal(t, int64(2), body.Tools[0].Count)
		assert.Equal(t, int64(1), body.Tools[0].Errors)
		assert.Equal(t, "100ms", body.Tools[0].AvgDuration)
		assert.Equal(t, 0.5, body.Tools[0].ErrorRate)
	})

	t.Run("policy", func(t *testing.T) {
		rec := performObservabilityRequest(r, "/metrics/policy")

		assert.Equal(t, http.StatusOK, rec.Code)

		var body struct {
			Blocks   int64            `json:"blocks"`
			Observes int64            `json:"observes"`
			ByReason map[string]int64 `json:"byReason"`
		}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		assert.Equal(t, int64(1), body.Blocks)
		assert.Equal(t, int64(1), body.Observes)
		assert.Equal(t, int64(1), body.ByReason["policy-deny"])
		assert.Equal(t, int64(1), body.ByReason["policy-observe"])
	})

	t.Run("agents", func(t *testing.T) {
		rec := performObservabilityRequest(r, "/metrics/agents")

		assert.Equal(t, http.StatusOK, rec.Code)

		var body struct {
			Agents []struct {
				Name         string `json:"name"`
				InputTokens  int64  `json:"inputTokens"`
				OutputTokens int64  `json:"outputTokens"`
				ToolCalls    int64  `json:"toolCalls"`
			} `json:"agents"`
		}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		require.Len(t, body.Agents, 1)
		assert.Equal(t, "agent-1", body.Agents[0].Name)
		assert.Equal(t, int64(11), body.Agents[0].InputTokens)
		assert.Equal(t, int64(7), body.Agents[0].OutputTokens)
		assert.Equal(t, int64(2), body.Agents[0].ToolCalls)
	})
}

func TestObservabilityRoutes_AlertsSuccess(t *testing.T) {
	t.Parallel()

	r := chi.NewRouter()
	collector := observability.NewCollector()
	alertTimestamp := time.Date(2026, 5, 20, 10, 30, 0, 0, time.UTC)
	var gotFrom time.Time
	registerObservabilityRoutes(r, collector, nil, nil, func(_ context.Context, from time.Time) ([]storage.AlertRecord, error) {
		gotFrom = from
		return []storage.AlertRecord{
			{
				ID:        "alert-1",
				Type:      "policy_block_rate",
				Actor:     "sentinel",
				Details:   map[string]interface{}{"threshold": float64(3)},
				Timestamp: alertTimestamp,
			},
		}, nil
	}, nil)

	before := time.Now().AddDate(0, 0, -3)
	rec := performObservabilityRequest(r, "/alerts?days=3")
	after := time.Now().AddDate(0, 0, -3)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, gotFrom.Before(before))
	assert.False(t, gotFrom.After(after))

	var body struct {
		Alerts []struct {
			ID        string                 `json:"id"`
			Type      string                 `json:"type"`
			Actor     string                 `json:"actor"`
			Details   map[string]interface{} `json:"details"`
			Timestamp string                 `json:"timestamp"`
		} `json:"alerts"`
		Total int `json:"total"`
		Days  int `json:"days"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Len(t, body.Alerts, 1)
	assert.Equal(t, "alert-1", body.Alerts[0].ID)
	assert.Equal(t, "policy_block_rate", body.Alerts[0].Type)
	assert.Equal(t, "sentinel", body.Alerts[0].Actor)
	assert.Equal(t, float64(3), body.Alerts[0].Details["threshold"])
	assert.Equal(t, alertTimestamp.Format(time.RFC3339), body.Alerts[0].Timestamp)
	assert.Equal(t, 1, body.Total)
	assert.Equal(t, 3, body.Days)
}

func TestObservabilityRoutes_AlertsReaderError(t *testing.T) {
	t.Parallel()

	r := chi.NewRouter()
	collector := observability.NewCollector()
	registerObservabilityRoutes(r, collector, nil, nil, func(context.Context, time.Time) ([]storage.AlertRecord, error) {
		return nil, errors.New("audit unavailable")
	}, nil)

	rec := performObservabilityRequest(r, "/alerts")

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, "audit unavailable\n", rec.Body.String())
}

func TestObservabilityWriteJSON_SetsContentTypeAndEncodesNewline(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	writeObsJSON(rec, map[string]interface{}{"ok": true})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, "{\"ok\":true}\n", rec.Body.String())
}

func performObservabilityRequest(handler http.Handler, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
