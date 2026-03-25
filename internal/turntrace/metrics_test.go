package turntrace

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeAgentMetrics(t *testing.T) {
	now := time.Now()
	end1 := now.Add(2 * time.Second)
	end2 := now.Add(5 * time.Second)
	end3 := now.Add(10 * time.Second)

	traces := []Trace{
		{TraceID: "t1", Outcome: OutcomeSuccess, StartedAt: now, EndedAt: &end1},
		{TraceID: "t2", Outcome: OutcomeTimeout, StartedAt: now, EndedAt: &end2},
		{TraceID: "t3", Outcome: OutcomeSuccess, StartedAt: now, EndedAt: &end3},
	}
	events := []Event{
		{TraceID: "t1", EventType: EventDelegation, AgentName: "orchestrator", PayloadJSON: `{"to":"operator"}`},
		{TraceID: "t1", EventType: EventToolCall, AgentName: "operator"},
		{TraceID: "t1", EventType: EventToolCall, AgentName: "operator"},
		{TraceID: "t2", EventType: EventDelegation, AgentName: "orchestrator", PayloadJSON: `{"to":"navigator"}`},
		{TraceID: "t3", EventType: EventDelegation, AgentName: "orchestrator", PayloadJSON: `{"to":"operator"}`},
		{TraceID: "t3", EventType: EventToolCall, AgentName: "operator"},
	}

	result := ComputeAgentMetrics(traces, events)

	require.Contains(t, result, "operator")
	op := result["operator"]
	assert.Equal(t, 2, op.TotalTurns)
	assert.Equal(t, 2, op.SuccessCount)
	assert.Equal(t, 0, op.FailureCount)
	assert.Equal(t, 3, op.ToolCallCount)
	assert.Equal(t, 1.0, op.SuccessRate)

	require.Contains(t, result, "navigator")
	nav := result["navigator"]
	assert.Equal(t, 1, nav.TotalTurns)
	assert.Equal(t, 0, nav.SuccessCount)
	assert.Equal(t, 1, nav.FailureCount)
	assert.Equal(t, 0.0, nav.SuccessRate)
}

func TestComputeAgentMetrics_NonDelegatedAttributedToOrchestrator(t *testing.T) {
	now := time.Now()
	end := now.Add(3 * time.Second)

	traces := []Trace{
		{TraceID: "t1", Entrypoint: "tui", Outcome: OutcomeSuccess, StartedAt: now, EndedAt: &end},
	}
	// No delegation events — direct turn
	events := []Event{
		{TraceID: "t1", EventType: EventToolCall, AgentName: "lango-orchestrator"},
	}

	result := ComputeAgentMetrics(traces, events)
	require.Contains(t, result, "lango-orchestrator")
	assert.Equal(t, 1, result["lango-orchestrator"].TotalTurns)
	assert.NotContains(t, result, "tui")

	traces2 := []Trace{
		{TraceID: "t2", Entrypoint: "gateway", Outcome: OutcomeSuccess, StartedAt: now, EndedAt: &end},
	}
	events2 := []Event{
		{TraceID: "t2", EventType: EventText, AgentName: "lango-agent"},
	}
	result2 := ComputeAgentMetrics(traces2, events2)
	require.Contains(t, result2, "lango-agent")
	assert.Equal(t, 1, result2["lango-agent"].TotalTurns)
	assert.NotContains(t, result2, "gateway")
}

func TestComputeAgentMetrics_Empty(t *testing.T) {
	result := ComputeAgentMetrics(nil, nil)
	assert.Empty(t, result)
}

func TestPercentile(t *testing.T) {
	tests := []struct {
		give    string
		sorted  []time.Duration
		p       float64
		wantMin time.Duration
		wantMax time.Duration
	}{
		{
			give:   "empty",
			sorted: nil,
			p:      0.5,
		},
		{
			give:    "single",
			sorted:  []time.Duration{100 * time.Millisecond},
			p:       0.5,
			wantMin: 100 * time.Millisecond,
			wantMax: 100 * time.Millisecond,
		},
		{
			give:    "p50 of 10 items",
			sorted:  makeDurations(10),
			p:       0.5,
			wantMin: 4 * time.Second,
			wantMax: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result := percentile(tt.sorted, tt.p)
			if tt.sorted == nil {
				assert.Equal(t, time.Duration(0), result)
				return
			}
			assert.GreaterOrEqual(t, result, tt.wantMin)
			assert.LessOrEqual(t, result, tt.wantMax)
		})
	}
}

func makeDurations(n int) []time.Duration {
	d := make([]time.Duration, n)
	for i := range d {
		d[i] = time.Duration(i) * time.Second
	}
	return d
}
