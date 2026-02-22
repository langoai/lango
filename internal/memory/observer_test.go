package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/session"
	_ "github.com/mattn/go-sqlite3"
)

type mockGenerator struct {
	response string
	err      error
}

func (m *mockGenerator) GenerateText(_ context.Context, _, _ string) (string, error) {
	return m.response, m.err
}

func newTestObserver(t *testing.T, gen TextGenerator) (*Observer, *Store) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	logger := zap.NewNop().Sugar()
	store := NewStore(client, logger)
	observer := NewObserver(gen, store, logger)
	return observer, store
}

func TestObserve(t *testing.T) {
	t.Run("generates observation and saves it", func(t *testing.T) {
		gen := &mockGenerator{response: "User discussed building a REST API with Go."}
		observer, store := newTestObserver(t, gen)
		ctx := context.Background()

		messages := []session.Message{
			{Role: "user", Content: "I want to build a REST API", Timestamp: time.Now()},
			{Role: "assistant", Content: "Sure, let's use Go with Chi router", Timestamp: time.Now()},
			{Role: "user", Content: "Sounds good, let's start", Timestamp: time.Now()},
		}

		obs, err := observer.Observe(ctx, "session-obs-1", messages, -1)
		require.NoError(t, err)
		require.NotNil(t, obs)

		assert.Equal(t, "User discussed building a REST API with Go.", obs.Content)
		assert.Equal(t, "session-obs-1", obs.SessionKey)
		assert.Equal(t, 0, obs.SourceStartIndex)
		assert.Equal(t, 2, obs.SourceEndIndex)
		assert.Greater(t, obs.TokenCount, 0)

		// Verify it was persisted.
		saved, err := store.ListObservations(ctx, "session-obs-1")
		require.NoError(t, err)
		assert.Len(t, saved, 1)
		assert.Equal(t, obs.Content, saved[0].Content)
	})

	t.Run("partial observation from lastObservedIndex", func(t *testing.T) {
		gen := &mockGenerator{response: "User decided on PostgreSQL."}
		observer, _ := newTestObserver(t, gen)
		ctx := context.Background()

		messages := []session.Message{
			{Role: "user", Content: "I want to build a REST API", Timestamp: time.Now()},
			{Role: "assistant", Content: "Sure", Timestamp: time.Now()},
			{Role: "user", Content: "Let's use PostgreSQL", Timestamp: time.Now()},
			{Role: "assistant", Content: "Good choice", Timestamp: time.Now()},
		}

		obs, err := observer.Observe(ctx, "session-obs-2", messages, 1)
		require.NoError(t, err)
		require.NotNil(t, obs)

		assert.Equal(t, 2, obs.SourceStartIndex)
		assert.Equal(t, 3, obs.SourceEndIndex)
	})

	t.Run("empty messages returns nil", func(t *testing.T) {
		gen := &mockGenerator{response: "should not be called"}
		observer, _ := newTestObserver(t, gen)
		ctx := context.Background()

		obs, err := observer.Observe(ctx, "session-obs-3", nil, -1)
		require.NoError(t, err)
		assert.Nil(t, obs)
	})

	t.Run("all messages already observed returns nil", func(t *testing.T) {
		gen := &mockGenerator{response: "should not be called"}
		observer, _ := newTestObserver(t, gen)
		ctx := context.Background()

		messages := []session.Message{
			{Role: "user", Content: "Hello", Timestamp: time.Now()},
		}

		obs, err := observer.Observe(ctx, "session-obs-4", messages, 0)
		require.NoError(t, err)
		assert.Nil(t, obs)
	})
}

func TestFormatMessages(t *testing.T) {
	tests := []struct {
		give string
		msgs []session.Message
		want string
	}{
		{
			give: "simple messages",
			msgs: []session.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there"},
			},
			want: "[user]: Hello\n[assistant]: Hi there\n",
		},
		{
			give: "message with tool calls",
			msgs: []session.Message{
				{
					Role:    "assistant",
					Content: "Let me check",
					ToolCalls: []session.ToolCall{
						{Name: "read_file", Input: "main.go", Output: "package main"},
					},
				},
			},
			want: "[assistant]: Let me check\n  [tool:read_file] main.go\n  [result] package main\n",
		},
		{
			give: "tool call without output",
			msgs: []session.Message{
				{
					Role:    "assistant",
					Content: "Running",
					ToolCalls: []session.ToolCall{
						{Name: "exec", Input: "go build"},
					},
				},
			},
			want: "[assistant]: Running\n  [tool:exec] go build\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := formatMessages(tt.msgs)
			assert.Equal(t, tt.want, got)
		})
	}
}
