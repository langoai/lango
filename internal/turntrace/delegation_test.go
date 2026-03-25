package turntrace

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildDelegationGraph(t *testing.T) {
	now := time.Now()

	tests := []struct {
		give       string
		events     []Event
		wantEdges  int
		wantAgents int
	}{
		{
			give:       "empty events",
			events:     nil,
			wantEdges:  0,
			wantAgents: 0,
		},
		{
			give: "single delegation",
			events: []Event{
				{TraceID: "t1", EventType: EventDelegation, AgentName: "orchestrator", PayloadJSON: `{"to":"operator"}`, CreatedAt: now},
			},
			wantEdges:  1,
			wantAgents: 2,
		},
		{
			give: "round trip delegation",
			events: []Event{
				{TraceID: "t1", EventType: EventDelegation, AgentName: "orchestrator", PayloadJSON: `{"to":"operator"}`, CreatedAt: now},
				{TraceID: "t1", EventType: EventToolCall, AgentName: "operator", CreatedAt: now.Add(time.Second)},
				{TraceID: "t1", EventType: EventDelegationReturn, AgentName: "operator", PayloadJSON: `{"to":"lango-orchestrator"}`, CreatedAt: now.Add(2 * time.Second)},
			},
			wantEdges:  2,
			wantAgents: 3, // orchestrator, operator, lango-orchestrator
		},
		{
			give: "tool calls counted",
			events: []Event{
				{TraceID: "t1", EventType: EventToolCall, AgentName: "operator", CreatedAt: now},
				{TraceID: "t1", EventType: EventToolCall, AgentName: "operator", CreatedAt: now.Add(time.Second)},
			},
			wantEdges:  0,
			wantAgents: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			g := BuildDelegationGraph("session-1", nil, tt.events)
			assert.Equal(t, tt.wantEdges, len(g.Edges))
			assert.Equal(t, tt.wantAgents, len(g.Agents))
		})
	}
}

func TestBuildDelegationGraph_ToolCallCount(t *testing.T) {
	now := time.Now()
	events := []Event{
		{TraceID: "t1", EventType: EventToolCall, AgentName: "operator", CreatedAt: now},
		{TraceID: "t1", EventType: EventToolCall, AgentName: "operator", CreatedAt: now.Add(time.Second)},
		{TraceID: "t1", EventType: EventToolCall, AgentName: "navigator", CreatedAt: now.Add(2 * time.Second)},
	}

	g := BuildDelegationGraph("s1", nil, events)
	require.Contains(t, g.Agents, "operator")
	assert.Equal(t, 2, g.Agents["operator"].ToolCalls)
	require.Contains(t, g.Agents, "navigator")
	assert.Equal(t, 1, g.Agents["navigator"].ToolCalls)
}
