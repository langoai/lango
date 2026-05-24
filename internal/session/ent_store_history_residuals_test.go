package session

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEntStoreCreatePersistsInitialHistoryAuthorAndToolCalls(t *testing.T) {
	store := newTestEntStore(t)
	timestamp := time.Unix(1712345678, 987654321).UTC()
	wantToolCall := ToolCall{
		ID:               "call-1",
		Name:             "shell",
		Input:            `{"cmd":"printf hello"}`,
		Output:           "hello",
		Thought:          true,
		ThoughtSignature: []byte("sig"),
	}
	want := &Session{
		Key:         "sess-history-residuals",
		AgentID:     "agent-1",
		ChannelType: "cli",
		ChannelID:   "terminal",
		Model:       "test-model",
		Metadata:    map[string]string{"purpose": "coverage"},
		History: []Message{
			{
				Role:      "assistant",
				Content:   "I will inspect the workspace.",
				Timestamp: timestamp,
				Author:    "operator",
				ToolCalls: []ToolCall{wantToolCall},
			},
		},
	}

	require.NoError(t, store.Create(want))

	got, err := store.Get(want.Key)
	require.NoError(t, err)
	require.Len(t, got.History, 1)
	gotMessage := got.History[0]
	require.Equal(t, want.History[0].Role, gotMessage.Role)
	require.Equal(t, want.History[0].Content, gotMessage.Content)
	require.True(t, gotMessage.Timestamp.Equal(timestamp), "timestamp: want %s, got %s", timestamp, gotMessage.Timestamp)
	require.Equal(t, "operator", gotMessage.Author)
	require.Equal(t, []ToolCall{wantToolCall}, gotMessage.ToolCalls)

	row, err := store.client.Message.Query().Only(context.Background())
	require.NoError(t, err)
	require.Equal(t, want.History[0].Content, row.Content)
	require.True(t, row.ContentCiphertext == nil || len(*row.ContentCiphertext) == 0)
	require.True(t, row.ToolCallsCiphertext == nil || len(*row.ToolCallsCiphertext) == 0)
	require.True(t, row.ToolCallsNonce == nil || len(*row.ToolCallsNonce) == 0)
	require.True(t, row.ToolCallsKeyVersion == nil || *row.ToolCallsKeyVersion == 0)
	require.Equal(t, "operator", row.Author)
	require.Len(t, row.ToolCalls, 1)
	require.Equal(t, wantToolCall.ID, row.ToolCalls[0].ID)
	require.Equal(t, wantToolCall.Name, row.ToolCalls[0].Name)
	require.Equal(t, wantToolCall.Input, row.ToolCalls[0].Input)
	require.Equal(t, wantToolCall.Output, row.ToolCalls[0].Output)
	require.Equal(t, wantToolCall.Thought, row.ToolCalls[0].Thought)
	require.Equal(t, wantToolCall.ThoughtSignature, row.ToolCalls[0].ThoughtSignature)
}
