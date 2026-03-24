package turntrace

import (
	"sort"
	"time"
)

// AgentMetrics holds per-agent performance counters.
type AgentMetrics struct {
	AgentName      string          `json:"agent_name"`
	TotalTurns     int             `json:"total_turns"`
	SuccessCount   int             `json:"success_count"`
	FailureCount   int             `json:"failure_count"`
	ToolCallCount  int             `json:"tool_call_count"`
	DelegationsIn  int             `json:"delegations_in"`
	DelegationsOut int             `json:"delegations_out"`
	Durations      []time.Duration `json:"-"`
}

// AgentMetricsSummary includes computed percentiles.
type AgentMetricsSummary struct {
	AgentMetrics
	P50Duration time.Duration `json:"p50_duration_ms"`
	P95Duration time.Duration `json:"p95_duration_ms"`
	P99Duration time.Duration `json:"p99_duration_ms"`
	SuccessRate float64       `json:"success_rate"`
}

// ComputeAgentMetrics derives per-agent performance statistics from traces and events.
func ComputeAgentMetrics(traces []Trace, events []Event) map[string]*AgentMetricsSummary {
	raw := make(map[string]*AgentMetrics)

	ensure := func(name string) *AgentMetrics {
		if name == "" {
			return nil
		}
		m, ok := raw[name]
		if !ok {
			m = &AgentMetrics{AgentName: name}
			raw[name] = m
		}
		return m
	}

	// Build delegation graph from events.
	for _, ev := range events {
		m := ensure(ev.AgentName)
		if m == nil {
			continue
		}
		switch ev.EventType {
		case EventToolCall:
			m.ToolCallCount++
		case EventDelegation:
			m.DelegationsOut++
			target := extractTarget(ev.PayloadJSON)
			if t := ensure(target); t != nil {
				t.DelegationsIn++
			}
		case EventDelegationReturn:
			m.DelegationsOut++
			target := extractTarget(ev.PayloadJSON)
			if t := ensure(target); t != nil {
				t.DelegationsIn++
			}
		}
	}

	// Attribute traces to agents via entrypoint or first delegation target.
	traceAgentMap := buildTraceAgentMap(events)
	for _, trace := range traces {
		agentName := traceAgentMap[trace.TraceID]
		if agentName == "" {
			agentName = trace.Entrypoint
		}
		m := ensure(agentName)
		if m == nil {
			continue
		}
		m.TotalTurns++
		if trace.Outcome == OutcomeSuccess {
			m.SuccessCount++
		} else if trace.Outcome != OutcomeRunning {
			m.FailureCount++
		}
		if trace.EndedAt != nil {
			m.Durations = append(m.Durations, trace.EndedAt.Sub(trace.StartedAt))
		}
	}

	// Compute summaries.
	result := make(map[string]*AgentMetricsSummary, len(raw))
	for name, m := range raw {
		s := &AgentMetricsSummary{AgentMetrics: *m}
		if m.TotalTurns > 0 {
			s.SuccessRate = float64(m.SuccessCount) / float64(m.TotalTurns)
		}
		if len(m.Durations) > 0 {
			sort.Slice(m.Durations, func(i, j int) bool { return m.Durations[i] < m.Durations[j] })
			s.P50Duration = percentile(m.Durations, 0.50)
			s.P95Duration = percentile(m.Durations, 0.95)
			s.P99Duration = percentile(m.Durations, 0.99)
		}
		result[name] = s
	}
	return result
}

// buildTraceAgentMap maps trace IDs to the first delegation target agent.
func buildTraceAgentMap(events []Event) map[string]string {
	m := make(map[string]string)
	for _, ev := range events {
		if _, ok := m[ev.TraceID]; ok {
			continue
		}
		if ev.EventType == EventDelegation {
			target := extractTarget(ev.PayloadJSON)
			if target != "" {
				m[ev.TraceID] = target
			}
		}
	}
	return m
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
