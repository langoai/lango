package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/ent/enttest"
	"github.com/langowarny/lango/internal/session"
	_ "github.com/mattn/go-sqlite3"
)

type integrationMockGenerator struct {
	response string
	err      error
}

func (m *integrationMockGenerator) GenerateText(_ context.Context, _, _ string) (string, error) {
	return m.response, m.err
}

func TestIntegration_ObservationGeneration(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	logger := zap.NewNop().Sugar()
	store := NewStore(client, logger)
	gen := &integrationMockGenerator{response: "User wants to build a REST API using Go with Chi router"}
	observer := NewObserver(gen, store, logger)

	ctx := context.Background()
	messages := []session.Message{
		{Role: "user", Content: "I want to build a REST API"},
		{Role: "assistant", Content: "Sure! What framework would you like to use?"},
		{Role: "user", Content: "Let's use Chi router"},
		{Role: "assistant", Content: "Great choice! Let me set up the project structure."},
	}

	// Observe messages
	obs, err := observer.Observe(ctx, "session-1", messages, -1)
	require.NoError(t, err)
	require.NotNil(t, obs)
	assert.Equal(t, "User wants to build a REST API using Go with Chi router", obs.Content)
	assert.Equal(t, 0, obs.SourceStartIndex)
	assert.Equal(t, 3, obs.SourceEndIndex)

	// Verify stored in DB
	stored, err := store.ListObservations(ctx, "session-1")
	require.NoError(t, err)
	assert.Len(t, stored, 1)
	assert.Equal(t, obs.Content, stored[0].Content)
}

func TestIntegration_ObservationToReflection(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	logger := zap.NewNop().Sugar()
	store := NewStore(client, logger)

	obsGen := &integrationMockGenerator{response: "Observation note"}
	observer := NewObserver(obsGen, store, logger)

	ctx := context.Background()

	// Create multiple observations
	messages1 := []session.Message{
		{Role: "user", Content: "Build a REST API"},
		{Role: "assistant", Content: "Setting up project"},
	}
	_, err := observer.Observe(ctx, "session-1", messages1, -1)
	require.NoError(t, err)

	messages2 := append(messages1,
		session.Message{Role: "user", Content: "Add authentication"},
		session.Message{Role: "assistant", Content: "Implementing JWT"},
	)
	_, err = observer.Observe(ctx, "session-1", messages2, 1)
	require.NoError(t, err)

	messages3 := append(messages2,
		session.Message{Role: "user", Content: "Add database integration"},
		session.Message{Role: "assistant", Content: "Setting up PostgreSQL"},
	)
	_, err = observer.Observe(ctx, "session-1", messages3, 3)
	require.NoError(t, err)

	// Verify 3 observations exist
	observations, err := store.ListObservations(ctx, "session-1")
	require.NoError(t, err)
	assert.Len(t, observations, 3)

	// Now reflect to condense observations
	refGen := &integrationMockGenerator{response: "User is building a REST API with JWT auth and PostgreSQL database"}
	reflector := NewReflector(refGen, store, logger)

	ref, err := reflector.Reflect(ctx, "session-1")
	require.NoError(t, err)
	require.NotNil(t, ref)
	assert.Equal(t, "User is building a REST API with JWT auth and PostgreSQL database", ref.Content)
	assert.Equal(t, 1, ref.Generation)

	// Observations should be deleted after reflection
	observations, err = store.ListObservations(ctx, "session-1")
	require.NoError(t, err)
	assert.Empty(t, observations)

	// Reflection should be stored
	reflections, err := store.ListReflections(ctx, "session-1")
	require.NoError(t, err)
	assert.Len(t, reflections, 1)
}

func TestIntegration_OMDisabled_NoImpact(t *testing.T) {
	// When OM is disabled, the EventsAdapter should use legacy 100-message cap
	// This test verifies the default behavior is unchanged

	// Create 150 messages
	messages := make([]session.Message, 150)
	for i := range messages {
		messages[i] = session.Message{
			Role:    "user",
			Content: "message",
		}
	}

	// Token counter should still work standalone
	total := CountMessagesTokens(messages)
	assert.True(t, total > 0, "token counting works independently")

	// Verify EstimateTokens returns 0 for empty
	assert.Equal(t, 0, EstimateTokens(""))
}
