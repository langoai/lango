package memory

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolparam"
)

// BuildObservationTools creates tools for observational memory management.
func BuildObservationTools(ms *Store) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "memory_list_observations",
			Description: "List observations for a session. Returns compressed notes from conversation history.",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "memory",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"session_key": map[string]interface{}{"type": "string", "description": "Session key to list observations for (uses current session if empty)"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				sessionKey := toolparam.OptionalString(params, "session_key", session.SessionKeyFromContext(ctx))
				observations, err := ms.ListObservations(ctx, sessionKey)
				if err != nil {
					return nil, fmt.Errorf("list observations: %w", err)
				}
				return map[string]interface{}{"observations": observations, "count": len(observations)}, nil
			},
		},
		{
			Name:        "memory_list_reflections",
			Description: "List reflections for a session. Reflections are condensed observations across time.",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "memory",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"session_key": map[string]interface{}{"type": "string", "description": "Session key to list reflections for (uses current session if empty)"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				sessionKey := toolparam.OptionalString(params, "session_key", session.SessionKeyFromContext(ctx))
				reflections, err := ms.ListReflections(ctx, sessionKey)
				if err != nil {
					return nil, fmt.Errorf("list reflections: %w", err)
				}
				return map[string]interface{}{"reflections": reflections, "count": len(reflections)}, nil
			},
		},
	}
}
