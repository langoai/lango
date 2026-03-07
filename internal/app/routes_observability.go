package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/observability/health"
	"github.com/langoai/lango/internal/observability/token"
)

// registerObservabilityRoutes adds observability HTTP endpoints to the router.
func registerObservabilityRoutes(r chi.Router, collector *observability.MetricsCollector, hr *health.Registry, store *token.EntTokenStore) {
	if collector == nil {
		return
	}

	r.Get("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		snap := collector.Snapshot()
		writeObsJSON(w, map[string]interface{}{
			"uptime":         snap.Uptime.Round(time.Second).String(),
			"startedAt":      snap.StartedAt.Format(time.RFC3339),
			"toolExecutions": snap.ToolExecutions,
			"tokenUsage": map[string]interface{}{
				"inputTokens":  snap.TokenUsageTotal.InputTokens,
				"outputTokens": snap.TokenUsageTotal.OutputTokens,
				"totalTokens":  snap.TokenUsageTotal.TotalTokens,
				"cacheTokens":  snap.TokenUsageTotal.CacheTokens,
			},
			"sessionCount": len(snap.SessionBreakdown),
			"agentCount":   len(snap.AgentBreakdown),
			"toolCount":    len(snap.ToolBreakdown),
		})
	})

	r.Get("/metrics/sessions", func(w http.ResponseWriter, _ *http.Request) {
		snap := collector.Snapshot()
		sessions := make([]map[string]interface{}, 0, len(snap.SessionBreakdown))
		for _, s := range snap.SessionBreakdown {
			sessions = append(sessions, map[string]interface{}{
				"sessionKey":   s.SessionKey,
				"inputTokens":  s.InputTokens,
				"outputTokens": s.OutputTokens,
				"totalTokens":  s.TotalTokens,
				"requestCount": s.RequestCount,
			})
		}
		writeObsJSON(w, map[string]interface{}{"sessions": sessions})
	})

	r.Get("/metrics/tools", func(w http.ResponseWriter, _ *http.Request) {
		snap := collector.Snapshot()
		tools := make([]map[string]interface{}, 0, len(snap.ToolBreakdown))
		for _, t := range snap.ToolBreakdown {
			errRate := 0.0
			if t.Count > 0 {
				errRate = float64(t.Errors) / float64(t.Count)
			}
			tools = append(tools, map[string]interface{}{
				"name":        t.Name,
				"count":       t.Count,
				"errors":      t.Errors,
				"avgDuration": t.AvgDuration.String(),
				"errorRate":   errRate,
			})
		}
		writeObsJSON(w, map[string]interface{}{"tools": tools})
	})

	r.Get("/metrics/agents", func(w http.ResponseWriter, _ *http.Request) {
		snap := collector.Snapshot()
		agents := make([]map[string]interface{}, 0, len(snap.AgentBreakdown))
		for _, a := range snap.AgentBreakdown {
			agents = append(agents, map[string]interface{}{
				"name":         a.Name,
				"inputTokens":  a.InputTokens,
				"outputTokens": a.OutputTokens,
				"toolCalls":    a.ToolCalls,
			})
		}
		writeObsJSON(w, map[string]interface{}{"agents": agents})
	})

	// History endpoint — requires persistent store
	if store != nil {
		r.Get("/metrics/history", func(w http.ResponseWriter, r *http.Request) {
			daysStr := r.URL.Query().Get("days")
			days := 7
			if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
				days = d
			}

			from := time.Now().AddDate(0, 0, -days)
			to := time.Now()

			records, err := store.QueryByTimeRange(r.Context(), from, to)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var totalInput, totalOutput int64
			items := make([]map[string]interface{}, len(records))
			for i, rec := range records {
				totalInput += rec.InputTokens
				totalOutput += rec.OutputTokens
				items[i] = map[string]interface{}{
					"provider":     rec.Provider,
					"model":        rec.Model,
					"sessionKey":   rec.SessionKey,
					"agentName":    rec.AgentName,
					"inputTokens":  rec.InputTokens,
					"outputTokens": rec.OutputTokens,
					"timestamp":    rec.Timestamp.Format(time.RFC3339),
				}
			}

			writeObsJSON(w, map[string]interface{}{
				"records": items,
				"total": map[string]interface{}{
					"inputTokens":  totalInput,
					"outputTokens": totalOutput,
					"recordCount":  len(records),
				},
			})
		})
	}

	// Detailed health endpoint
	if hr != nil {
		r.Get("/health/detailed", func(w http.ResponseWriter, r *http.Request) {
			result := hr.CheckAll(r.Context())
			writeObsJSON(w, result)
		})
	}
}

func writeObsJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
