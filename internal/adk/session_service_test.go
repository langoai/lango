package adk

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internal "github.com/langoai/lango/internal/session"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func newTestEvent(author string, role string, text string) *session.Event {
	evt := &session.Event{
		Timestamp: time.Now(),
		Author:    author,
	}
	evt.Content = &genai.Content{
		Role:  role,
		Parts: []*genai.Part{{Text: text}},
	}
	return evt
}

func TestAppendEvent_UpdatesInMemoryHistory(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:       "test-session",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.Create(sess)

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	svc := NewSessionServiceAdapter(store, "lango-agent")

	evt := newTestEvent("user", "user", "hello")

	require.NoError(t, svc.AppendEvent(context.Background(), adapter, evt))

	// Verify in-memory history was updated
	require.Len(t, adapter.sess.History, 1)
	assert.Equal(t, "user", string(adapter.sess.History[0].Role))
	assert.Equal(t, "hello", adapter.sess.History[0].Content)

	// Events() should now return the message
	events := adapter.Events()
	assert.Equal(t, 1, events.Len())
}

func TestAppendEvent_MultipleEvents_AccumulateHistory(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:       "test-session",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.Create(sess)

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	svc := NewSessionServiceAdapter(store, "lango-agent")

	// Append user message
	require.NoError(t, svc.AppendEvent(context.Background(), adapter, newTestEvent("user", "user", "hello")))

	// Append assistant message
	require.NoError(t, svc.AppendEvent(context.Background(), adapter, newTestEvent("lango-agent", "model", "hi there")))

	// Verify both messages in in-memory history
	require.Len(t, adapter.sess.History, 2)
	assert.Equal(t, "user", string(adapter.sess.History[0].Role))
	assert.Equal(t, "assistant", string(adapter.sess.History[1].Role))

	// Events() should see both messages
	events := adapter.Events()
	assert.Equal(t, 2, events.Len())
}

func TestAppendEvent_StateDelta_SkipsHistory(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:       "test-session",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.Create(sess)

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	svc := NewSessionServiceAdapter(store, "lango-agent")

	// Pure state-delta event (no LLMResponse content)
	evt := &session.Event{
		Timestamp: time.Now(),
		Author:    "lango-agent",
		Actions: session.EventActions{
			StateDelta: map[string]any{"counter": 1},
		},
	}

	require.NoError(t, svc.AppendEvent(context.Background(), adapter, evt))

	// State-delta-only events should not append to history
	assert.Empty(t, adapter.sess.History, "expected 0 messages for state-delta event")
}

func TestAppendEvent_DBAndMemoryBothUpdated(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:       "test-session",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.Create(sess)

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	svc := NewSessionServiceAdapter(store, "lango-agent")

	evt := newTestEvent("user", "user", "hello")
	require.NoError(t, svc.AppendEvent(context.Background(), adapter, evt))

	// Verify DB store has the message
	dbMsgs := store.messages["test-session"]
	require.Len(t, dbMsgs, 1)
	assert.Equal(t, "hello", dbMsgs[0].Content)

	// Verify in-memory history also has the message
	require.Len(t, adapter.sess.History, 1)
}

func TestAppendEvent_PreservesAuthor(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:       "test-session",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.Create(sess)

	adapter := NewSessionAdapter(sess, store, "lango-orchestrator")
	svc := NewSessionServiceAdapter(store, "lango-orchestrator")

	evt := newTestEvent("lango-orchestrator", "model", "hello from orchestrator")
	require.NoError(t, svc.AppendEvent(context.Background(), adapter, evt))

	// Verify author was preserved in in-memory history
	require.Len(t, adapter.sess.History, 1)
	assert.Equal(t, "lango-orchestrator", adapter.sess.History[0].Author)

	// Verify author was preserved in DB store
	dbMsgs := store.messages["test-session"]
	require.Len(t, dbMsgs, 1)
	assert.Equal(t, "lango-orchestrator", dbMsgs[0].Author)
}

func TestAppendEvent_PreservesFunctionCallID(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:       "test-session",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.Create(sess)

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	svc := NewSessionServiceAdapter(store, "lango-agent")

	// Event with FunctionCall that has an original ID
	evt := &session.Event{
		Timestamp: time.Now(),
		Author:    "lango-agent",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{
				Role: "model",
				Parts: []*genai.Part{{
					FunctionCall: &genai.FunctionCall{
						ID:   "adk-original-uuid-123",
						Name: "exec",
						Args: map[string]any{"command": "ls"},
					},
				}},
			},
		},
	}

	require.NoError(t, svc.AppendEvent(context.Background(), adapter, evt))

	require.Len(t, adapter.sess.History, 1)
	msg := adapter.sess.History[0]
	require.Len(t, msg.ToolCalls, 1)
	assert.Equal(t, "adk-original-uuid-123", msg.ToolCalls[0].ID)
	assert.Equal(t, "exec", msg.ToolCalls[0].Name)
}

func TestAppendEvent_FunctionCallFallbackID(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:       "test-session",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.Create(sess)

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	svc := NewSessionServiceAdapter(store, "lango-agent")

	// FunctionCall without ID — should get synthetic fallback
	evt := &session.Event{
		Timestamp: time.Now(),
		Author:    "lango-agent",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{
				Role: "model",
				Parts: []*genai.Part{{
					FunctionCall: &genai.FunctionCall{
						Name: "search",
						Args: map[string]any{"query": "test"},
					},
				}},
			},
		},
	}

	require.NoError(t, svc.AppendEvent(context.Background(), adapter, evt))

	msg := adapter.sess.History[0]
	assert.Equal(t, "call_search", msg.ToolCalls[0].ID)
}

func TestAppendEvent_SavesFunctionResponseMetadata(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:       "test-session",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.Create(sess)

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	svc := NewSessionServiceAdapter(store, "lango-agent")

	// Event with FunctionResponse
	evt := &session.Event{
		Timestamp: time.Now(),
		Author:    "tool",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{
				Role: "function",
				Parts: []*genai.Part{{
					FunctionResponse: &genai.FunctionResponse{
						ID:       "adk-original-uuid-123",
						Name:     "exec",
						Response: map[string]any{"output": "file.txt"},
					},
				}},
			},
		},
	}

	require.NoError(t, svc.AppendEvent(context.Background(), adapter, evt))

	require.Len(t, adapter.sess.History, 1)
	msg := adapter.sess.History[0]

	// Should have ToolCalls with FunctionResponse metadata
	require.Len(t, msg.ToolCalls, 1)
	tc := msg.ToolCalls[0]
	assert.Equal(t, "adk-original-uuid-123", tc.ID)
	assert.Equal(t, "exec", tc.Name)
	assert.NotEmpty(t, tc.Output)

	// Content should also contain the response for backward compatibility
	assert.NotEmpty(t, msg.Content, "expected non-empty Content for backward compat")
}

func TestSessionServiceAdapter_Get_ExpiredSession_AutoRenews(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	// Seed an expired session
	store.sessions["expired-sess"] = &internal.Session{
		Key:      "expired-sess",
		Metadata: map[string]string{"old": "data"},
	}
	store.expiredKeys["expired-sess"] = true

	service := NewSessionServiceAdapter(store, "lango-agent")

	resp, err := service.Get(context.Background(), &session.GetRequest{
		SessionID: "expired-sess",
	})
	require.NoError(t, err, "expected auto-renew")
	assert.Equal(t, "expired-sess", resp.Session.ID())

	// Old session should have been deleted and replaced
	assert.False(t, store.expiredKeys["expired-sess"], "expected expiredKeys entry to be cleared after delete")

	// Verify session exists in store (recreated)
	sess, err := store.Get("expired-sess")
	require.NoError(t, err, "expected session in store after auto-renew")
	require.NotNil(t, sess, "expected non-nil session after auto-renew")
	// Old metadata should not carry over (new session is blank)
	assert.NotEqual(t, "data", sess.Metadata["old"], "expected old metadata to be cleared in renewed session")
}

func TestSessionServiceAdapter_Get_ExpiredSession_DeleteFails(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	store.sessions["fail-del"] = &internal.Session{Key: "fail-del"}
	store.expiredKeys["fail-del"] = true
	store.deleteErr = fmt.Errorf("disk full")

	service := NewSessionServiceAdapter(store, "lango-agent")

	_, err := service.Get(context.Background(), &session.GetRequest{
		SessionID: "fail-del",
	})
	require.Error(t, err, "expected error when delete fails")
	assert.True(t, errors.Is(err, store.deleteErr), "expected wrapped disk full error")
}

// Verify the LLMResponse field is unused in model import (for compile check)
var _ = model.LLMResponse{}
