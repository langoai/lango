package memory

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildObservationTools_UseCurrentSessionWhenSessionKeyOmitted(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()
	require.NoError(t, store.SaveObservation(ctx, Observation{
		SessionKey: "session-1",
		Content:    "first session observation",
		TokenCount: 5,
	}))
	require.NoError(t, store.SaveObservation(ctx, Observation{
		SessionKey: "session-2",
		Content:    "second session observation",
		TokenCount: 5,
	}))
	require.NoError(t, store.SaveReflection(ctx, Reflection{
		SessionKey: "session-1",
		Content:    "first session reflection",
		TokenCount: 8,
		Generation: 1,
	}))
	require.NoError(t, store.SaveReflection(ctx, Reflection{
		SessionKey: "session-2",
		Content:    "second session reflection",
		TokenCount: 8,
		Generation: 1,
	}))

	tools := BuildObservationTools(store)

	observationResult, err := findMemoryTool(t, tools, "memory_list_observations").Handler(
		session.WithSessionKey(context.Background(), "session-1"),
		map[string]interface{}{},
	)
	require.NoError(t, err)
	observationPayload := observationResult.(map[string]interface{})
	observationEntries := observationPayload["observations"].([]Observation)
	require.Len(t, observationEntries, 1)
	assert.Equal(t, "first session observation", observationEntries[0].Content)

	reflectionResult, err := findMemoryTool(t, tools, "memory_list_reflections").Handler(
		session.WithSessionKey(context.Background(), "session-1"),
		map[string]interface{}{},
	)
	require.NoError(t, err)
	reflectionPayload := reflectionResult.(map[string]interface{})
	reflectionEntries := reflectionPayload["reflections"].([]Reflection)
	require.Len(t, reflectionEntries, 1)
	assert.Equal(t, "first session reflection", reflectionEntries[0].Content)
}

func findMemoryTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
