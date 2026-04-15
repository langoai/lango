package adk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internal "github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

type mockStore struct {
	sessions    map[string]*internal.Session
	messages    map[string][]internal.Message // DB-only message storage
	expiredKeys map[string]bool               // keys that simulate expired sessions
	deleteErr   error                         // if set, Delete returns this error
}

func newMockStore() *mockStore {
	return &mockStore{
		sessions:    make(map[string]*internal.Session),
		messages:    make(map[string][]internal.Message),
		expiredKeys: make(map[string]bool),
	}
}

func (m *mockStore) Create(s *internal.Session) error {
	m.sessions[s.Key] = s
	return nil
}
func (m *mockStore) Get(key string) (*internal.Session, error) {
	if m.expiredKeys[key] {
		return nil, fmt.Errorf("get session %q: %w", key, internal.ErrSessionExpired)
	}
	s, ok := m.sessions[key]
	if !ok {
		return nil, nil
	}
	return s, nil
}
func (m *mockStore) Update(s *internal.Session) error {
	m.sessions[s.Key] = s
	return nil
}
func (m *mockStore) Delete(key string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.sessions, key)
	delete(m.expiredKeys, key)
	return nil
}
func (m *mockStore) AppendMessage(key string, msg internal.Message) error {
	// Store in separate messages map (simulates DB-only storage, not in-memory History)
	m.messages[key] = append(m.messages[key], msg)
	return nil
}
func (m *mockStore) AnnotateTimeout(_ string, _ string) error { return nil }
func (m *mockStore) End(_ string) error                       { return nil }
func (m *mockStore) Close() error                             { return nil }
func (m *mockStore) GetSalt(name string) ([]byte, error)      { return nil, nil }
func (m *mockStore) SetSalt(name string, salt []byte) error   { return nil }
func (m *mockStore) ListSessions(_ context.Context) ([]internal.SessionSummary, error) {
	return nil, nil
}

// --- StateAdapter tests ---

func TestStateAdapter_SetGet(t *testing.T) {
	t.Parallel()

	sess := &internal.Session{
		Key:      "test-session",
		Metadata: make(map[string]string),
	}
	store := newMockStore()
	store.Create(sess)

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	state := adapter.State()

	// Test Set string
	err := state.Set("foo", "bar")
	require.NoError(t, err)

	// Verify update in store
	updatedSess, _ := store.Get("test-session")
	assert.Equal(t, "bar", updatedSess.Metadata["foo"])

	// Test Get string
	val, err := state.Get("foo")
	require.NoError(t, err)
	assert.Equal(t, "bar", val)

	// Test Set complex object (should be JSON encoded)
	obj := map[string]int{"a": 1}
	err = state.Set("obj", obj)
	require.NoError(t, err)

	// Verify JSON in metadata
	expectedJSON, _ := json.Marshal(obj)
	assert.Equal(t, string(expectedJSON), updatedSess.Metadata["obj"])

	// Test Get complex object
	val, err = state.Get("obj")
	require.NoError(t, err)
	valMap, ok := val.(map[string]any)
	require.True(t, ok, "expected map[string]any, got %T", val)
	assert.Equal(t, float64(1), valMap["a"]) // JSON numbers are float64
}

func TestStateAdapter_GetNonExistent(t *testing.T) {
	t.Parallel()

	sess := &internal.Session{
		Key:      "test-session",
		Metadata: make(map[string]string),
	}
	store := newMockStore()
	store.Create(sess)

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	state := adapter.State()

	_, err := state.Get("nonexistent")
	assert.ErrorIs(t, err, session.ErrStateKeyNotExist)
}

func TestStateAdapter_SetNilMetadata(t *testing.T) {
	t.Parallel()

	sess := &internal.Session{
		Key: "test-session",
		// Metadata is nil
	}
	store := newMockStore()
	store.Create(sess)

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	state := adapter.State()

	// Set should initialize metadata if nil
	err := state.Set("key", "value")
	require.NoError(t, err)

	val, err := state.Get("key")
	require.NoError(t, err)
	assert.Equal(t, "value", val)
}

func TestStateAdapter_All(t *testing.T) {
	t.Parallel()

	sess := &internal.Session{
		Key: "test-session",
		Metadata: map[string]string{
			"key1": "value1",
			"key2": `{"nested": true}`,
		},
	}
	store := newMockStore()
	store.Create(sess)

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	state := adapter.State()

	count := 0
	for k, v := range state.All() {
		count++
		switch k {
		case "key1":
			assert.Equal(t, "value1", v)
		case "key2":
			m, ok := v.(map[string]any)
			require.True(t, ok, "expected map for key2, got %T", v)
			assert.Equal(t, true, m["nested"])
		default:
			t.Errorf("unexpected key %q", k)
		}
	}
	assert.Equal(t, 2, count)
}

// --- SessionAdapter tests ---

func TestSessionAdapter_BasicFields(t *testing.T) {
	t.Parallel()

	now := time.Now()
	sess := &internal.Session{
		Key:       "sess-123",
		UpdatedAt: now,
	}
	store := newMockStore()
	adapter := NewSessionAdapter(sess, store, "lango-agent")

	assert.Equal(t, "sess-123", adapter.ID())
	assert.Equal(t, "lango", adapter.AppName())
	assert.Equal(t, "user", adapter.UserID())
	assert.True(t, adapter.LastUpdateTime().Equal(now))
}

// --- EventsAdapter tests ---

func TestEventsAdapter(t *testing.T) {
	t.Parallel()

	now := time.Now()
	sess := &internal.Session{
		History: []internal.Message{
			{Role: "user", Content: "hello", Timestamp: now},
			{Role: "assistant", Content: "hi", Timestamp: now.Add(time.Second)},
		},
	}

	adapter := NewSessionAdapter(sess, &mockStore{}, "lango-agent")
	events := adapter.Events()

	count := 0
	for event := range events.All() {
		count++
		assert.False(t, event.Timestamp.IsZero(), "expected non-zero timestamp")
	}

	assert.Equal(t, 2, count)
}

func TestEventsAdapter_AuthorMapping(t *testing.T) {
	t.Parallel()

	now := time.Now()
	sess := &internal.Session{
		History: []internal.Message{
			{Role: "user", Content: "hello", Timestamp: now},
			{Role: "assistant", Content: "hi", Timestamp: now.Add(time.Second)},
			{Role: "tool", Content: "result", Timestamp: now.Add(2 * time.Second)},
			{Role: "function", Content: "response", Timestamp: now.Add(3 * time.Second)},
		},
	}

	adapter := NewSessionAdapter(sess, &mockStore{}, "lango-agent")
	events := adapter.Events()

	expectedAuthors := []string{"user", "lango-agent", "tool", "tool"}
	i := 0
	for evt := range events.All() {
		if i < len(expectedAuthors) {
			assert.Equal(t, expectedAuthors[i], evt.Author, "event %d author mismatch", i)
		}
		i++
	}
	assert.Equal(t, 4, i)
}

func TestEventsAdapter_AuthorMapping_MultiAgent(t *testing.T) {
	t.Parallel()

	now := time.Now()
	sess := &internal.Session{
		History: []internal.Message{
			{Role: "user", Content: "hello", Timestamp: now},
			// Stored author from a previous multi-agent event.
			{Role: "assistant", Content: "hi", Author: "lango-orchestrator", Timestamp: now.Add(time.Second)},
			// Interleave user message to prevent role merging.
			{Role: "user", Content: "follow up", Timestamp: now.Add(2 * time.Second)},
			// No stored author — should fall back to rootAgentName.
			{Role: "assistant", Content: "ok", Timestamp: now.Add(3 * time.Second)},
		},
	}

	adapter := NewSessionAdapter(sess, &mockStore{}, "lango-orchestrator")
	events := adapter.Events()

	expectedAuthors := []string{"user", "lango-orchestrator", "user", "lango-orchestrator"}
	i := 0
	for evt := range events.All() {
		if i < len(expectedAuthors) {
			assert.Equal(t, expectedAuthors[i], evt.Author, "event %d author mismatch", i)
		}
		i++
	}
	assert.Equal(t, 4, i)
}

func TestEventsAdapter_Truncation(t *testing.T) {
	t.Parallel()

	// Create 150 small messages with alternating roles — all fit within default token budget.
	var msgs []internal.Message
	now := time.Now()
	roles := []types.MessageRole{"user", "assistant"}
	for i := range 150 {
		msgs = append(msgs, internal.Message{
			Role:      roles[i%2],
			Content:   "msg",
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}

	sess := &internal.Session{History: msgs}
	adapter := NewSessionAdapter(sess, &mockStore{}, "lango-agent")
	events := adapter.Events()

	// All 150 small messages should fit within the default token budget.
	assert.Equal(t, 150, events.Len())

	// Count events from All()
	count := 0
	for range events.All() {
		count++
	}
	assert.Equal(t, 150, count)

	// With an explicit small budget, messages should be truncated.
	budgetEvents := adapter.EventsWithTokenBudget(30)
	assert.Less(t, budgetEvents.Len(), 150, "expected truncation with small budget")
}

func TestEventsAdapter_WithToolCalls(t *testing.T) {
	t.Parallel()

	now := time.Now()
	sess := &internal.Session{
		History: []internal.Message{
			{
				Role:    "assistant",
				Content: "",
				ToolCalls: []internal.ToolCall{
					{
						ID:    "call_1",
						Name:  "exec",
						Input: `{"command":"ls"}`,
					},
				},
				Timestamp: now,
			},
		},
	}

	adapter := NewSessionAdapter(sess, &mockStore{}, "lango-agent")
	events := adapter.Events()

	count := 0
	for evt := range events.All() {
		count++
		require.NotNil(t, evt.Content)
		hasFunctionCall := false
		for _, p := range evt.Content.Parts {
			if p.FunctionCall != nil {
				hasFunctionCall = true
				assert.Equal(t, "exec", p.FunctionCall.Name)
				assert.Equal(t, "ls", p.FunctionCall.Args["command"])
			}
		}
		assert.True(t, hasFunctionCall, "expected a FunctionCall part in event")
	}
	assert.Equal(t, 1, count)
}

func TestEventsAdapter_EmptyHistory(t *testing.T) {
	t.Parallel()

	sess := &internal.Session{}
	adapter := NewSessionAdapter(sess, &mockStore{}, "lango-agent")
	events := adapter.Events()

	assert.Equal(t, 0, events.Len())

	count := 0
	for range events.All() {
		count++
	}
	assert.Equal(t, 0, count)
}

func TestEventsAdapter_At(t *testing.T) {
	t.Parallel()

	now := time.Now()
	sess := &internal.Session{
		History: []internal.Message{
			{Role: "user", Content: "first", Timestamp: now},
			{Role: "assistant", Content: "second", Timestamp: now.Add(time.Second)},
			{Role: "user", Content: "third", Timestamp: now.Add(2 * time.Second)},
		},
	}

	adapter := NewSessionAdapter(sess, &mockStore{}, "lango-agent")
	events := adapter.Events()

	evt0 := events.At(0)
	require.NotNil(t, evt0, "expected non-nil event at index 0")
	assert.Equal(t, "first", evt0.LLMResponse.Content.Parts[0].Text)

	evt2 := events.At(2)
	require.NotNil(t, evt2, "expected non-nil event at index 2")
	assert.Equal(t, "third", evt2.LLMResponse.Content.Parts[0].Text)
}

// --- Token-Budget Truncation tests ---

func TestEventsAdapter_TokenBudgetTruncation(t *testing.T) {
	t.Parallel()

	t.Run("includes all messages within budget", func(t *testing.T) {
		t.Parallel()
		var msgs []internal.Message
		roles := []types.MessageRole{"user", "assistant"}
		for i := range 6 {
			msgs = append(msgs, internal.Message{
				Role:      roles[i%2],
				Content:   "short",
				Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			})
		}
		adapter := &EventsAdapter{
			history:     msgs,
			tokenBudget: 10000,
		}
		assert.Equal(t, 6, adapter.Len())
	})

	t.Run("truncates when budget exceeded", func(t *testing.T) {
		t.Parallel()
		var msgs []internal.Message
		// Each message has 400 chars content = ~100 tokens + 4 overhead = ~104 tokens
		for range 20 {
			content := ""
			for range 400 {
				content += "a"
			}
			msgs = append(msgs, internal.Message{
				Role:      "user",
				Content:   content,
				Timestamp: time.Now(),
			})
		}
		adapter := &EventsAdapter{
			history:     msgs,
			tokenBudget: 500, // can fit ~4-5 messages
		}
		resultLen := adapter.Len()
		assert.Less(t, resultLen, 20, "expected truncation")
		assert.GreaterOrEqual(t, resultLen, 1, "expected at least 1 message")
	})

	t.Run("always includes at least one message", func(t *testing.T) {
		t.Parallel()
		msgs := []internal.Message{{
			Role:      "user",
			Content:   string(make([]byte, 40000)), // huge message
			Timestamp: time.Now(),
		}}
		adapter := &EventsAdapter{
			history:     msgs,
			tokenBudget: 10,
		}
		assert.Equal(t, 1, adapter.Len())
	})

	t.Run("empty history", func(t *testing.T) {
		t.Parallel()
		adapter := &EventsAdapter{
			history:     nil,
			tokenBudget: 100,
		}
		assert.Equal(t, 0, adapter.Len())
	})

	t.Run("preserves most recent messages", func(t *testing.T) {
		t.Parallel()
		var msgs []internal.Message
		for i := range 10 {
			content := ""
			for range 40 {
				content += "x"
			}
			msgs = append(msgs, internal.Message{
				Role:      "user",
				Content:   content, // ~10 tokens + 4 overhead = 14 tokens each
				Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			})
		}
		// Budget for ~2 messages: 28 tokens
		adapter := &EventsAdapter{
			history:     msgs,
			tokenBudget: 30,
		}
		truncated := adapter.truncatedHistory()
		require.Len(t, truncated, 2)
		// Should be the last 2 messages
		assert.Equal(t, msgs[8].Content, truncated[0].Content, "expected 9th message (index 8)")
		assert.Equal(t, msgs[9].Content, truncated[1].Content, "expected 10th message (index 9)")
	})
}

func TestEventsAdapter_DefaultTokenBudget(t *testing.T) {
	t.Parallel()

	var msgs []internal.Message
	roles := []types.MessageRole{"user", "assistant"}
	for i := range 150 {
		msgs = append(msgs, internal.Message{
			Role:      roles[i%2],
			Content:   "msg",
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		})
	}
	// tokenBudget=0 means use DefaultTokenBudget
	adapter := &EventsAdapter{
		history:     msgs,
		tokenBudget: 0,
	}
	// With DefaultTokenBudget (32000) and tiny messages (~1 token each),
	// all 150 messages should fit within the budget.
	assert.Equal(t, 150, adapter.Len())
}

// --- FunctionResponse reconstruction tests ---

func TestEventsAdapter_FunctionResponseReconstruction(t *testing.T) {
	t.Parallel()

	now := time.Now()

	t.Run("new format with ToolCalls metadata", func(t *testing.T) {
		t.Parallel()
		sess := &internal.Session{
			History: []internal.Message{
				{Role: "user", Content: "run ls", Timestamp: now},
				{
					Role: "assistant",
					ToolCalls: []internal.ToolCall{
						{ID: "adk-abc-123", Name: "exec", Input: `{"command":"ls"}`},
					},
					Timestamp: now.Add(time.Second),
				},
				{
					Role: "tool",
					ToolCalls: []internal.ToolCall{
						{ID: "adk-abc-123", Name: "exec", Output: `{"result":"file.txt"}`},
					},
					Content:   `{"result":"file.txt"}`,
					Timestamp: now.Add(2 * time.Second),
				},
			},
		}

		adapter := &EventsAdapter{history: sess.History, rootAgentName: "lango-agent"}
		var events []*session.Event
		for evt := range adapter.All() {
			events = append(events, evt)
		}

		require.Len(t, events, 3)

		// Verify assistant event has FunctionCall with ID
		assistantEvt := events[1]
		assert.Equal(t, "assistant", assistantEvt.Content.Role)
		var fc *genai.FunctionCall
		for _, p := range assistantEvt.Content.Parts {
			if p.FunctionCall != nil {
				fc = p.FunctionCall
			}
		}
		require.NotNil(t, fc, "expected FunctionCall part in assistant event")
		assert.Equal(t, "adk-abc-123", fc.ID)
		assert.Equal(t, "exec", fc.Name)

		// Verify tool event has FunctionResponse
		toolEvt := events[2]
		assert.Equal(t, "function", toolEvt.Content.Role)
		var fr *genai.FunctionResponse
		for _, p := range toolEvt.Content.Parts {
			if p.FunctionResponse != nil {
				fr = p.FunctionResponse
			}
		}
		require.NotNil(t, fr, "expected FunctionResponse part in tool event")
		assert.Equal(t, "adk-abc-123", fr.ID)
		assert.Equal(t, "exec", fr.Name)
		assert.Equal(t, "file.txt", fr.Response["result"])
	})

	t.Run("legacy format without ToolCalls on tool message", func(t *testing.T) {
		t.Parallel()
		sess := &internal.Session{
			History: []internal.Message{
				{Role: "user", Content: "run ls", Timestamp: now},
				{
					Role: "assistant",
					ToolCalls: []internal.ToolCall{
						{ID: "call_exec", Name: "exec", Input: `{"command":"ls"}`},
					},
					Timestamp: now.Add(time.Second),
				},
				{
					Role:      "tool",
					Content:   `{"result":"file.txt"}`,
					Timestamp: now.Add(2 * time.Second),
					// No ToolCalls — legacy format
				},
			},
		}

		adapter := &EventsAdapter{history: sess.History, rootAgentName: "lango-agent"}
		var events []*session.Event
		for evt := range adapter.All() {
			events = append(events, evt)
		}

		require.Len(t, events, 3)

		// Verify tool event has FunctionResponse reconstructed from legacy
		toolEvt := events[2]
		assert.Equal(t, "function", toolEvt.Content.Role)
		var fr *genai.FunctionResponse
		for _, p := range toolEvt.Content.Parts {
			if p.FunctionResponse != nil {
				fr = p.FunctionResponse
			}
		}
		require.NotNil(t, fr, "expected FunctionResponse part in legacy tool event")
		assert.Equal(t, "call_exec", fr.ID)
		assert.Equal(t, "exec", fr.Name)
	})

	t.Run("tool message without preceding assistant ToolCalls falls back to text", func(t *testing.T) {
		t.Parallel()
		sess := &internal.Session{
			History: []internal.Message{
				{Role: "user", Content: "hello", Timestamp: now},
				{
					Role:      "tool",
					Content:   "some result",
					Timestamp: now.Add(time.Second),
					// No preceding assistant with ToolCalls
				},
			},
		}

		adapter := &EventsAdapter{history: sess.History, rootAgentName: "lango-agent"}
		var events []*session.Event
		for evt := range adapter.All() {
			events = append(events, evt)
		}

		require.Len(t, events, 2)

		toolEvt := events[1]
		// Should fall back to text since no context to reconstruct FunctionResponse
		hasText := false
		for _, p := range toolEvt.Content.Parts {
			if p.Text != "" {
				hasText = true
			}
		}
		assert.True(t, hasText, "expected text part in tool event without FunctionResponse context")
	})
}

func TestEventsAdapter_FunctionResponseUserRole(t *testing.T) {
	t.Parallel()

	now := time.Now()
	// Simulate FunctionResponse stored with wrong role "user" (ADK bug).
	sess := &internal.Session{
		History: []internal.Message{
			{Role: "user", Content: "run ls", Timestamp: now},
			{
				Role: "assistant",
				ToolCalls: []internal.ToolCall{
					{ID: "call_abc", Name: "exec", Input: `{"command":"ls"}`},
				},
				Timestamp: now.Add(time.Second),
			},
			{
				Role: "user", // ADK incorrectly stores FunctionResponse as "user"
				ToolCalls: []internal.ToolCall{
					{ID: "call_abc", Name: "exec", Output: `{"result":"file.txt"}`},
				},
				Content:   `{"result":"file.txt"}`,
				Timestamp: now.Add(2 * time.Second),
			},
		},
	}

	adapter := &EventsAdapter{history: sess.History, rootAgentName: "lango-agent"}
	var events []*session.Event
	for evt := range adapter.All() {
		events = append(events, evt)
	}

	require.Len(t, events, 3)

	// The FunctionResponse event should be reconstructed with "function" role,
	// not "user", even though it was stored as "user".
	toolEvt := events[2]
	assert.Equal(t, "function", toolEvt.Content.Role)
	var fr *genai.FunctionResponse
	for _, p := range toolEvt.Content.Parts {
		if p.FunctionResponse != nil {
			fr = p.FunctionResponse
		}
	}
	require.NotNil(t, fr, "expected FunctionResponse part in corrected event")
	assert.Equal(t, "call_abc", fr.ID)
	assert.Equal(t, "exec", fr.Name)
	assert.Equal(t, "file.txt", fr.Response["result"])
	// Author should be "tool", not "user"
	assert.Equal(t, "tool", toolEvt.Author)
}

func TestEventsAdapter_FunctionResponseToolRole(t *testing.T) {
	t.Parallel()

	now := time.Now()
	// FunctionResponse stored with correct role "tool" — no correction needed.
	sess := &internal.Session{
		History: []internal.Message{
			{Role: "user", Content: "run ls", Timestamp: now},
			{
				Role: "assistant",
				ToolCalls: []internal.ToolCall{
					{ID: "call_abc", Name: "exec", Input: `{"command":"ls"}`},
				},
				Timestamp: now.Add(time.Second),
			},
			{
				Role: "tool", // Correct role
				ToolCalls: []internal.ToolCall{
					{ID: "call_abc", Name: "exec", Output: `{"result":"file.txt"}`},
				},
				Content:   `{"result":"file.txt"}`,
				Timestamp: now.Add(2 * time.Second),
			},
		},
	}

	adapter := &EventsAdapter{history: sess.History, rootAgentName: "lango-agent"}
	var events []*session.Event
	for evt := range adapter.All() {
		events = append(events, evt)
	}

	require.Len(t, events, 3)

	toolEvt := events[2]
	assert.Equal(t, "function", toolEvt.Content.Role)
	var fr *genai.FunctionResponse
	for _, p := range toolEvt.Content.Parts {
		if p.FunctionResponse != nil {
			fr = p.FunctionResponse
		}
	}
	require.NotNil(t, fr, "expected FunctionResponse part")
	assert.Equal(t, "call_abc", fr.ID)
	assert.Equal(t, "exec", fr.Name)
}

func TestEventsAdapter_ConsecutiveRoleMerging(t *testing.T) {
	t.Parallel()

	now := time.Now()

	t.Run("consecutive assistant turns are merged", func(t *testing.T) {
		t.Parallel()
		sess := &internal.Session{
			History: []internal.Message{
				{Role: "user", Content: "hello", Timestamp: now},
				{Role: "assistant", Content: "part1", Timestamp: now.Add(time.Second)},
				{Role: "assistant", Content: "part2", Timestamp: now.Add(2 * time.Second)},
			},
		}
		adapter := &EventsAdapter{history: sess.History, rootAgentName: "lango-agent"}
		var events []*session.Event
		for evt := range adapter.All() {
			events = append(events, evt)
		}
		require.Len(t, events, 2, "expected 2 events (merged)")
		// Second event should have 2 text parts from the merged assistant turns.
		assert.Len(t, events[1].Content.Parts, 2, "expected 2 parts in merged event")
	})

	t.Run("alternating roles are not merged", func(t *testing.T) {
		t.Parallel()
		sess := &internal.Session{
			History: []internal.Message{
				{Role: "user", Content: "hello", Timestamp: now},
				{Role: "assistant", Content: "hi", Timestamp: now.Add(time.Second)},
				{Role: "user", Content: "bye", Timestamp: now.Add(2 * time.Second)},
			},
		}
		adapter := &EventsAdapter{history: sess.History, rootAgentName: "lango-agent"}
		var events []*session.Event
		for evt := range adapter.All() {
			events = append(events, evt)
		}
		assert.Len(t, events, 3, "expected 3 events (no merging)")
	})

	t.Run("Len matches All count", func(t *testing.T) {
		t.Parallel()
		sess := &internal.Session{
			History: []internal.Message{
				{Role: "user", Content: "a", Timestamp: now},
				{Role: "assistant", Content: "b", Timestamp: now.Add(time.Second)},
				{Role: "assistant", Content: "c", Timestamp: now.Add(2 * time.Second)},
				{Role: "user", Content: "d", Timestamp: now.Add(3 * time.Second)},
			},
		}
		adapter := &EventsAdapter{history: sess.History, rootAgentName: "lango-agent"}
		count := 0
		for range adapter.All() {
			count++
		}
		assert.Equal(t, adapter.Len(), count, "Len() should match All() count")
		assert.Equal(t, 3, count)
	})
}

func TestEventsAdapter_TruncationSequenceSafety(t *testing.T) {
	t.Parallel()

	t.Run("skips leading tool message after truncation", func(t *testing.T) {
		t.Parallel()
		var msgs []internal.Message
		// Create many messages so truncation kicks in
		for i := range 20 {
			content := ""
			for range 200 {
				content += "x"
			}
			msgs = append(msgs, internal.Message{
				Role:      "user",
				Content:   content,
				Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			})
		}
		// Place a tool message at a position likely to be at the truncation boundary
		msgs[15] = internal.Message{
			Role:      "tool",
			Content:   `{"result":"ok"}`,
			Timestamp: time.Now().Add(15 * time.Second),
		}

		adapter := &EventsAdapter{
			history:     msgs,
			tokenBudget: 400, // enough for ~6-7 messages
		}
		truncated := adapter.truncatedHistory()

		if len(truncated) > 0 {
			first := truncated[0]
			assert.NotEqual(t, "tool", string(first.Role), "truncated history should not start with tool message")
			assert.NotEqual(t, "function", string(first.Role), "truncated history should not start with function message")
		}
	})

	t.Run("does not skip trailing FunctionCall without truncation", func(t *testing.T) {
		t.Parallel()
		msgs := []internal.Message{
			{Role: "user", Content: "hello", Timestamp: time.Now()},
			{
				Role: "assistant",
				ToolCalls: []internal.ToolCall{
					{ID: "call_1", Name: "exec", Input: `{"cmd":"ls"}`},
				},
				Timestamp: time.Now().Add(time.Second),
			},
		}

		adapter := &EventsAdapter{
			history:     msgs,
			tokenBudget: 100000, // no truncation
		}
		truncated := adapter.truncatedHistory()

		assert.Len(t, truncated, 2, "expected 2 messages (no truncation)")
	})
}

// --- SessionServiceAdapter tests ---

func TestSessionServiceAdapter_Create(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	service := NewSessionServiceAdapter(store, "lango-agent")

	resp, err := service.Create(context.Background(), &session.CreateRequest{
		SessionID: "new-session",
		State: map[string]any{
			"key": "value",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "new-session", resp.Session.ID())

	// Verify state was set
	val, err := resp.Session.State().Get("key")
	require.NoError(t, err)
	assert.Equal(t, "value", val)
}

func TestSessionServiceAdapter_Get(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	store.Create(&internal.Session{
		Key:      "existing",
		Metadata: map[string]string{"foo": "bar"},
	})

	service := NewSessionServiceAdapter(store, "lango-agent")

	resp, err := service.Get(context.Background(), &session.GetRequest{
		SessionID: "existing",
	})
	require.NoError(t, err)
	assert.Equal(t, "existing", resp.Session.ID())
}

func TestSessionServiceAdapter_GetAutoCreate(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	service := NewSessionServiceAdapter(store, "lango-agent")

	// Get on a nonexistent session should auto-create it
	resp, err := service.Get(context.Background(), &session.GetRequest{
		SessionID: "auto-created",
	})
	require.NoError(t, err, "expected auto-create")
	require.Equal(t, "auto-created", resp.Session.ID())

	// Verify session now exists in store
	sess, err := store.Get("auto-created")
	require.NoError(t, err, "expected session in store")
	assert.Equal(t, "auto-created", sess.Key)
}

// uniqueMockStore simulates UNIQUE constraint errors on concurrent Create.
type uniqueMockStore struct {
	mu       sync.Mutex
	sessions map[string]*internal.Session
}

func newUniqueMockStore() *uniqueMockStore {
	return &uniqueMockStore{sessions: make(map[string]*internal.Session)}
}

func (m *uniqueMockStore) Create(s *internal.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.sessions[s.Key]; exists {
		return fmt.Errorf("create session %q: %w", s.Key, internal.ErrDuplicateSession)
	}
	m.sessions[s.Key] = s
	return nil
}

func (m *uniqueMockStore) Get(key string) (*internal.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[key]
	if !ok {
		return nil, fmt.Errorf("get session %q: %w", key, internal.ErrSessionNotFound)
	}
	return s, nil
}

func (m *uniqueMockStore) Update(s *internal.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.Key] = s
	return nil
}
func (m *uniqueMockStore) Delete(key string) error                      { return nil }
func (m *uniqueMockStore) AppendMessage(string, internal.Message) error { return nil }
func (m *uniqueMockStore) AnnotateTimeout(string, string) error         { return nil }
func (m *uniqueMockStore) End(string) error                             { return nil }
func (m *uniqueMockStore) Close() error                                 { return nil }
func (m *uniqueMockStore) GetSalt(string) ([]byte, error)               { return nil, nil }
func (m *uniqueMockStore) SetSalt(string, []byte) error                 { return nil }
func (m *uniqueMockStore) ListSessions(context.Context) ([]internal.SessionSummary, error) {
	return nil, nil
}

func TestSessionServiceAdapter_GetAutoCreate_Concurrent(t *testing.T) {
	t.Parallel()

	store := newUniqueMockStore()
	service := NewSessionServiceAdapter(store, "lango-agent")

	const goroutines = 10
	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	wg.Add(goroutines)
	for i := range goroutines {
		go func() {
			defer wg.Done()
			_, errs[i] = service.Get(context.Background(), &session.GetRequest{
				SessionID: "race-session",
			})
		}()
	}
	wg.Wait()

	for i, err := range errs {
		assert.NoError(t, err, "goroutine %d failed", i)
	}
}

func TestSessionServiceAdapter_Delete(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	store.Create(&internal.Session{Key: "to-delete"})

	service := NewSessionServiceAdapter(store, "lango-agent")

	err := service.Delete(context.Background(), &session.DeleteRequest{
		SessionID: "to-delete",
	})
	require.NoError(t, err)

	// Verify deleted
	s, _ := store.Get("to-delete")
	assert.Nil(t, s, "expected session to be deleted")
}

func TestSessionServiceAdapter_List(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	service := NewSessionServiceAdapter(store, "lango-agent")

	resp, err := service.List(context.Background(), &session.ListRequest{})
	require.NoError(t, err)
	// Currently returns empty
	require.NotNil(t, resp)
}

func TestSessionServiceAdapter_AppendEvent_UserMessage(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:     "sess-1",
		History: nil,
	}
	store.Create(sess)

	service := NewSessionServiceAdapter(store, "lango-agent")
	adapter := NewSessionAdapter(sess, store, "lango-agent")

	evt := &session.Event{
		Author:    "user",
		Timestamp: time.Now(),
	}

	err := service.AppendEvent(context.Background(), adapter, evt)
	require.NoError(t, err)

	// Verify message was appended
	updated, _ := store.Get("sess-1")
	require.Len(t, updated.History, 1)
	assert.Equal(t, "user", string(updated.History[0].Role))
}

// --- convertMessages tests ---

func TestConvertMessages_RoleMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want string
	}{
		{"user", "user"},
		{"model", "assistant"},
		{"function", "tool"},
		{"system", "system"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			msgs, err := convertMessages([]*genai.Content{{
				Role:  tt.give,
				Parts: []*genai.Part{{Text: "test"}},
			}})
			require.NoError(t, err)
			require.Len(t, msgs, 1)
			assert.Equal(t, tt.want, string(msgs[0].Role))
		})
	}
}

func TestConvertMessages_TextContent(t *testing.T) {
	t.Parallel()

	msgs, err := convertMessages([]*genai.Content{{
		Role:  "user",
		Parts: []*genai.Part{{Text: "hello world"}},
	}})
	require.NoError(t, err)
	assert.Equal(t, "hello world", msgs[0].Content)
}

func TestConvertMessages_FunctionCall(t *testing.T) {
	t.Parallel()

	msgs, err := convertMessages([]*genai.Content{{
		Role: "model",
		Parts: []*genai.Part{{
			FunctionCall: &genai.FunctionCall{
				Name: "exec",
				Args: map[string]any{"cmd": "ls"},
			},
		}},
	}})
	require.NoError(t, err)
	require.Len(t, msgs[0].ToolCalls, 1)
	assert.Equal(t, "exec", msgs[0].ToolCalls[0].Name)
}

func TestConvertMessages_FunctionResponse(t *testing.T) {
	t.Parallel()

	msgs, err := convertMessages([]*genai.Content{{
		Role: "function",
		Parts: []*genai.Part{{
			FunctionResponse: &genai.FunctionResponse{
				Name:     "exec",
				Response: map[string]any{"output": "file.txt"},
			},
		}},
	}})
	require.NoError(t, err)
	assert.Equal(t, "tool", string(msgs[0].Role))
	assert.NotEmpty(t, msgs[0].Content, "expected non-empty content from function response")
	require.NotNil(t, msgs[0].Metadata)
	assert.Equal(t, "exec", msgs[0].Metadata["tool_call_id"])
}

func TestConvertMessages_Empty(t *testing.T) {
	t.Parallel()

	msgs, err := convertMessages(nil)
	require.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestConvertMessages_MultipleFunctionResponsesSplit(t *testing.T) {
	t.Parallel()

	// Simulate EventsAdapter merging 3 consecutive tool-role events into
	// a single Content with 3 FunctionResponse parts.
	merged := &genai.Content{
		Role: "function",
		Parts: []*genai.Part{
			{FunctionResponse: &genai.FunctionResponse{
				ID: "call_wallet", Name: "payment_wallet_info",
				Response: map[string]any{"address": "0xabc"},
			}},
			{FunctionResponse: &genai.FunctionResponse{
				ID: "call_balance", Name: "payment_balance",
				Response: map[string]any{"balance": "100"},
			}},
			{FunctionResponse: &genai.FunctionResponse{
				ID: "call_info", Name: "smart_account_info",
				Response: map[string]any{"deployed": true},
			}},
		},
	}

	msgs, err := convertMessages([]*genai.Content{merged})
	require.NoError(t, err)
	require.Len(t, msgs, 3, "merged FunctionResponses must split into 3 separate messages")

	ids := make(map[string]bool, 3)
	for _, m := range msgs {
		assert.Equal(t, "tool", m.Role)
		id, ok := m.Metadata["tool_call_id"].(string)
		require.True(t, ok, "each message must have a tool_call_id")
		ids[id] = true
		assert.NotEmpty(t, m.Content, "each message must have response content")
	}
	assert.True(t, ids["call_wallet"], "call_wallet must be present")
	assert.True(t, ids["call_balance"], "call_balance must be present")
	assert.True(t, ids["call_info"], "call_info must be present")
}

func TestConvertMessages_SingleFunctionResponseUnchanged(t *testing.T) {
	t.Parallel()

	// Single FunctionResponse should NOT be split — backward-compatible.
	content := &genai.Content{
		Role: "function",
		Parts: []*genai.Part{
			{FunctionResponse: &genai.FunctionResponse{
				ID: "call_1", Name: "exec",
				Response: map[string]any{"output": "ok"},
			}},
		},
	}

	msgs, err := convertMessages([]*genai.Content{content})
	require.NoError(t, err)
	require.Len(t, msgs, 1, "single response stays as one message")
	assert.Equal(t, "tool", msgs[0].Role)
	assert.Equal(t, "call_1", msgs[0].Metadata["tool_call_id"])
}

func TestConvertMessages_FunctionCallsStayMerged(t *testing.T) {
	t.Parallel()

	// FunctionCall parts in assistant message should remain merged (existing behavior).
	content := &genai.Content{
		Role: "model",
		Parts: []*genai.Part{
			{FunctionCall: &genai.FunctionCall{ID: "call_a", Name: "exec", Args: map[string]any{"cmd": "ls"}}},
			{FunctionCall: &genai.FunctionCall{ID: "call_b", Name: "search", Args: map[string]any{"q": "test"}}},
		},
	}

	msgs, err := convertMessages([]*genai.Content{content})
	require.NoError(t, err)
	require.Len(t, msgs, 1, "FunctionCall parts must stay merged in one assistant message")
	assert.Equal(t, "assistant", msgs[0].Role)
	assert.Len(t, msgs[0].ToolCalls, 2)
}

func TestConvertTools_NilConfig(t *testing.T) {
	t.Parallel()

	tools, err := convertTools(nil)
	require.NoError(t, err)
	assert.Empty(t, tools)
}

func TestConvertTools_NilTools(t *testing.T) {
	t.Parallel()

	cfg := &genai.GenerateContentConfig{}
	tools, err := convertTools(cfg)
	require.NoError(t, err)
	assert.Empty(t, tools)
}

func TestConvertTools_WithFunctionDeclarations(t *testing.T) {
	t.Parallel()

	cfg := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{{
			FunctionDeclarations: []*genai.FunctionDeclaration{{
				Name:        "test_tool",
				Description: "A test tool",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"arg1": {Type: genai.TypeString, Description: "First arg"},
					},
				},
			}},
		}},
	}

	tools, err := convertTools(cfg)
	require.NoError(t, err)
	require.Len(t, tools, 1)
	assert.Equal(t, "test_tool", tools[0].Name)
	assert.Equal(t, "A test tool", tools[0].Description)
}
