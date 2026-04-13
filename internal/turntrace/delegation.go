package turntrace

import (
	"encoding/json"
	"time"
)

// DelegationEdge represents a single agent-to-agent handoff.
type DelegationEdge struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Reason    string    `json:"reason,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	TraceID   string    `json:"trace_id"`
}

// AgentNode summarizes a single agent's participation in a session.
type AgentNode struct {
	Name           string    `json:"name"`
	DelegationsIn  int       `json:"delegations_in"`
	DelegationsOut int       `json:"delegations_out"`
	ToolCalls      int       `json:"tool_calls"`
	FirstSeen      time.Time `json:"first_seen"`
	LastSeen       time.Time `json:"last_seen"`
}

// DelegationGraph is a directed graph of agent-to-agent handoffs.
type DelegationGraph struct {
	SessionKey string                `json:"session_key"`
	Edges      []DelegationEdge      `json:"edges"`
	Agents     map[string]*AgentNode `json:"agents"`
}

// BuildDelegationGraph computes a delegation graph from traces and events.
func BuildDelegationGraph(sessionKey string, traces []Trace, events []Event) DelegationGraph {
	g := DelegationGraph{
		SessionKey: sessionKey,
		Agents:     make(map[string]*AgentNode),
	}

	for _, ev := range events {
		// Track agent presence from all events.
		ensureAgent(&g, ev.AgentName, ev.CreatedAt)

		switch ev.EventType {
		case EventDelegation, EventDelegationReturn:
			target := extractTarget(ev.PayloadJSON)
			if target == "" {
				continue
			}
			ensureAgent(&g, target, ev.CreatedAt)

			g.Edges = append(g.Edges, DelegationEdge{
				From:      ev.AgentName,
				To:        target,
				Timestamp: ev.CreatedAt,
				TraceID:   ev.TraceID,
			})
			if node := g.Agents[ev.AgentName]; node != nil {
				node.DelegationsOut++
			}
			if node := g.Agents[target]; node != nil {
				node.DelegationsIn++
			}

		case EventToolCall:
			if node := g.Agents[ev.AgentName]; node != nil {
				node.ToolCalls++
			}
		}
	}

	return g
}

func ensureAgent(g *DelegationGraph, name string, ts time.Time) {
	if name == "" {
		return
	}
	node, ok := g.Agents[name]
	if !ok {
		g.Agents[name] = &AgentNode{
			Name:      name,
			FirstSeen: ts,
			LastSeen:  ts,
		}
		return
	}
	if ts.Before(node.FirstSeen) {
		node.FirstSeen = ts
	}
	if ts.After(node.LastSeen) {
		node.LastSeen = ts
	}
}

func extractTarget(payloadJSON string) string {
	if payloadJSON == "" {
		return ""
	}
	var payload struct {
		To string `json:"to"`
	}
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		return ""
	}
	return payload.To
}
