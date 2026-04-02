package observability

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/toolchain"
)

func TestPrometheusExporter_Handler(t *testing.T) {
	t.Parallel()

	exp := NewPrometheusExporter()
	bus := eventbus.New()
	exp.Subscribe(bus)

	// Publish some events.
	bus.Publish(toolchain.ToolExecutedEvent{
		ToolName: "exec",
		Success:  true,
		Duration: 150 * time.Millisecond,
	})
	bus.Publish(toolchain.ToolExecutedEvent{
		ToolName: "exec",
		Success:  false,
		Duration: 50 * time.Millisecond,
	})
	bus.Publish(eventbus.PolicyDecisionEvent{
		Verdict: "block",
		Reason:  "catastrophic",
	})
	bus.Publish(eventbus.TokenUsageEvent{
		InputTokens:  100,
		OutputTokens: 50,
	})

	// Link collector so tracked_sessions gauge updates from token events.
	collector := NewCollector()
	exp.SetCollector(collector)
	// Record a token event through the collector to create a session.
	collector.RecordTokenUsage(TokenUsage{SessionKey: "sess-1", InputTokens: 10})
	collector.RecordTokenUsage(TokenUsage{SessionKey: "sess-2", InputTokens: 20})
	collector.RecordTokenUsage(TokenUsage{SessionKey: "sess-3", InputTokens: 30})
	collector.RecordTokenUsage(TokenUsage{SessionKey: "sess-4", InputTokens: 40})
	collector.RecordTokenUsage(TokenUsage{SessionKey: "sess-5", InputTokens: 50})
	exp.updateSessionGauge()

	// Serve the handler.
	ts := httptest.NewServer(exp.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	text := string(body)

	// Verify Prometheus text format contains expected metrics.
	assert.Contains(t, text, "lango_tool_executions_total")
	assert.Contains(t, text, `tool="exec"`)
	assert.Contains(t, text, "lango_tool_duration_seconds")
	assert.Contains(t, text, "lango_policy_decisions_total")
	assert.Contains(t, text, `verdict="block"`)
	assert.Contains(t, text, "lango_token_usage_total")
	assert.Contains(t, text, `type="input"`)
	assert.Contains(t, text, `type="output"`)
	assert.Contains(t, text, "lango_tracked_sessions 5")
}

func TestPrometheusExporter_NoEvents(t *testing.T) {
	t.Parallel()

	exp := NewPrometheusExporter()

	ts := httptest.NewServer(exp.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Metrics exist but with zero values or no samples.
	assert.True(t, strings.Contains(string(body), "lango_tracked_sessions 0") ||
		strings.Contains(string(body), "lango_tracked_sessions"), "tracked sessions gauge should exist")
}
