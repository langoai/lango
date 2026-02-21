package adk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	internal "github.com/langowarny/lango/internal/session"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

type mockStore struct {
	sessions map[string]*internal.Session
	messages map[string][]internal.Message // DB-only message storage
}

func newMockStore() *mockStore {
	return &mockStore{
		sessions: make(map[string]*internal.Session),
		messages: make(map[string][]internal.Message),
	}
}

func (m *mockStore) Create(s *internal.Session) error {
	m.sessions[s.Key] = s
	return nil
}
func (m *mockStore) Get(key string) (*internal.Session, error) {
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
	delete(m.sessions, key)
	return nil
}
func (m *mockStore) AppendMessage(key string, msg internal.Message) error {
	// Store in separate messages map (simulates DB-only storage, not in-memory History)
	m.messages[key] = append(m.messages[key], msg)
	return nil
}
func (m *mockStore) Close() error                           { return nil }
func (m *mockStore) GetSalt(name string) ([]byte, error)    { return nil, nil }
func (m *mockStore) SetSalt(name string, salt []byte) error { return nil }

// --- StateAdapter tests ---

func TestStateAdapter_SetGet(t *testing.T) {
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
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify update in store
	updatedSess, _ := store.Get("test-session")
	if updatedSess.Metadata["foo"] != "bar" {
		t.Errorf("expected 'bar', got %v", updatedSess.Metadata["foo"])
	}

	// Test Get string
	val, err := state.Get("foo")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "bar" {
		t.Errorf("expected 'bar', got %v", val)
	}

	// Test Set complex object (should be JSON encoded)
	obj := map[string]int{"a": 1}
	err = state.Set("obj", obj)
	if err != nil {
		t.Fatalf("Set complex failed: %v", err)
	}

	// Verify JSON in metadata
	expectedJSON, _ := json.Marshal(obj)
	if updatedSess.Metadata["obj"] != string(expectedJSON) {
		t.Errorf("expected JSON %s, got %s", string(expectedJSON), updatedSess.Metadata["obj"])
	}

	// Test Get complex object
	val, err = state.Get("obj")
	if err != nil {
		t.Fatalf("Get complex failed: %v", err)
	}
	valMap, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", val)
	}
	if valMap["a"] != float64(1) { // JSON numbers are float64
		t.Errorf("expected 1, got %v", valMap["a"])
	}
}

func TestStateAdapter_GetNonExistent(t *testing.T) {
	sess := &internal.Session{
		Key:      "test-session",
		Metadata: make(map[string]string),
	}
	store := newMockStore()
	store.Create(sess)

	adapter := NewSessionAdapter(sess, store, "lango-agent")
	state := adapter.State()

	_, err := state.Get("nonexistent")
	if err != session.ErrStateKeyNotExist {
		t.Errorf("expected ErrStateKeyNotExist, got %v", err)
	}
}

func TestStateAdapter_SetNilMetadata(t *testing.T) {
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
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := state.Get("key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value" {
		t.Errorf("expected 'value', got %v", val)
	}
}

func TestStateAdapter_All(t *testing.T) {
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
			if v != "value1" {
				t.Errorf("expected 'value1', got %v", v)
			}
		case "key2":
			m, ok := v.(map[string]any)
			if !ok {
				t.Errorf("expected map for key2, got %T", v)
			} else if m["nested"] != true {
				t.Errorf("expected nested=true, got %v", m["nested"])
			}
		default:
			t.Errorf("unexpected key %q", k)
		}
	}
	if count != 2 {
		t.Errorf("expected 2 entries, got %d", count)
	}
}

// --- SessionAdapter tests ---

func TestSessionAdapter_BasicFields(t *testing.T) {
	now := time.Now()
	sess := &internal.Session{
		Key:       "sess-123",
		UpdatedAt: now,
	}
	store := newMockStore()
	adapter := NewSessionAdapter(sess, store, "lango-agent")

	if adapter.ID() != "sess-123" {
		t.Errorf("expected ID 'sess-123', got %q", adapter.ID())
	}
	if adapter.AppName() != "lango" {
		t.Errorf("expected AppName 'lango', got %q", adapter.AppName())
	}
	if adapter.UserID() != "user" {
		t.Errorf("expected UserID 'user', got %q", adapter.UserID())
	}
	if !adapter.LastUpdateTime().Equal(now) {
		t.Errorf("expected LastUpdateTime %v, got %v", now, adapter.LastUpdateTime())
	}
}

// --- EventsAdapter tests ---

func TestEventsAdapter(t *testing.T) {
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
		if event.Timestamp.IsZero() {
			t.Error("expected non-zero timestamp")
		}
	}

	if count != 2 {
		t.Errorf("expected 2 events, got %d", count)
	}
}

func TestEventsAdapter_AuthorMapping(t *testing.T) {
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
		if i < len(expectedAuthors) && evt.Author != expectedAuthors[i] {
			t.Errorf("event %d: expected author %q, got %q", i, expectedAuthors[i], evt.Author)
		}
		i++
	}
	if i != 4 {
		t.Errorf("expected 4 events, got %d", i)
	}
}

func TestEventsAdapter_AuthorMapping_MultiAgent(t *testing.T) {
	now := time.Now()
	sess := &internal.Session{
		History: []internal.Message{
			{Role: "user", Content: "hello", Timestamp: now},
			// Stored author from a previous multi-agent event.
			{Role: "assistant", Content: "hi", Author: "lango-orchestrator", Timestamp: now.Add(time.Second)},
			// No stored author — should fall back to rootAgentName.
			{Role: "assistant", Content: "ok", Timestamp: now.Add(2 * time.Second)},
		},
	}

	adapter := NewSessionAdapter(sess, &mockStore{}, "lango-orchestrator")
	events := adapter.Events()

	expectedAuthors := []string{"user", "lango-orchestrator", "lango-orchestrator"}
	i := 0
	for evt := range events.All() {
		if i < len(expectedAuthors) && evt.Author != expectedAuthors[i] {
			t.Errorf("event %d: expected author %q, got %q", i, expectedAuthors[i], evt.Author)
		}
		i++
	}
	if i != 3 {
		t.Errorf("expected 3 events, got %d", i)
	}
}

func TestEventsAdapter_Truncation(t *testing.T) {
	// Create 150 small messages — all fit within default token budget.
	var msgs []internal.Message
	now := time.Now()
	for i := range 150 {
		msgs = append(msgs, internal.Message{
			Role:      "user",
			Content:   "msg",
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}

	sess := &internal.Session{History: msgs}
	adapter := NewSessionAdapter(sess, &mockStore{}, "lango-agent")
	events := adapter.Events()

	// All 150 small messages should fit within the default token budget.
	if events.Len() != 150 {
		t.Errorf("expected Len=150, got %d", events.Len())
	}

	// Count events from All()
	count := 0
	for range events.All() {
		count++
	}
	if count != 150 {
		t.Errorf("expected 150 events from All(), got %d", count)
	}

	// With an explicit small budget, messages should be truncated.
	budgetEvents := adapter.EventsWithTokenBudget(30)
	if budgetEvents.Len() >= 150 {
		t.Errorf("expected truncation with small budget, got %d", budgetEvents.Len())
	}
}

func TestEventsAdapter_WithToolCalls(t *testing.T) {
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
		if evt.LLMResponse.Content == nil {
			t.Fatal("expected non-nil content")
		}
		hasFunctionCall := false
		for _, p := range evt.LLMResponse.Content.Parts {
			if p.FunctionCall != nil {
				hasFunctionCall = true
				if p.FunctionCall.Name != "exec" {
					t.Errorf("expected function name 'exec', got %q", p.FunctionCall.Name)
				}
				if p.FunctionCall.Args["command"] != "ls" {
					t.Errorf("expected arg command='ls', got %v", p.FunctionCall.Args["command"])
				}
			}
		}
		if !hasFunctionCall {
			t.Error("expected a FunctionCall part in event")
		}
	}
	if count != 1 {
		t.Errorf("expected 1 event, got %d", count)
	}
}

func TestEventsAdapter_EmptyHistory(t *testing.T) {
	sess := &internal.Session{}
	adapter := NewSessionAdapter(sess, &mockStore{}, "lango-agent")
	events := adapter.Events()

	if events.Len() != 0 {
		t.Errorf("expected Len=0, got %d", events.Len())
	}

	count := 0
	for range events.All() {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 events, got %d", count)
	}
}

func TestEventsAdapter_At(t *testing.T) {
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
	if evt0 == nil {
		t.Fatal("expected non-nil event at index 0")
	}
	if evt0.LLMResponse.Content.Parts[0].Text != "first" {
		t.Errorf("expected 'first', got %q", evt0.LLMResponse.Content.Parts[0].Text)
	}

	evt2 := events.At(2)
	if evt2 == nil {
		t.Fatal("expected non-nil event at index 2")
	}
	if evt2.LLMResponse.Content.Parts[0].Text != "third" {
		t.Errorf("expected 'third', got %q", evt2.LLMResponse.Content.Parts[0].Text)
	}
}

// --- Token-Budget Truncation tests ---

func TestEventsAdapter_TokenBudgetTruncation(t *testing.T) {
	t.Run("includes all messages within budget", func(t *testing.T) {
		var msgs []internal.Message
		for range 5 {
			msgs = append(msgs, internal.Message{
				Role:      "user",
				Content:   "short",
				Timestamp: time.Now(),
			})
		}
		adapter := &EventsAdapter{
			history:     msgs,
			tokenBudget: 10000,
		}
		if adapter.Len() != 5 {
			t.Errorf("expected 5, got %d", adapter.Len())
		}
	})

	t.Run("truncates when budget exceeded", func(t *testing.T) {
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
		if resultLen >= 20 {
			t.Errorf("expected truncation, got %d", resultLen)
		}
		if resultLen < 1 {
			t.Error("expected at least 1 message")
		}
	})

	t.Run("always includes at least one message", func(t *testing.T) {
		msgs := []internal.Message{{
			Role:      "user",
			Content:   string(make([]byte, 40000)), // huge message
			Timestamp: time.Now(),
		}}
		adapter := &EventsAdapter{
			history:     msgs,
			tokenBudget: 10,
		}
		if adapter.Len() != 1 {
			t.Errorf("expected 1 message, got %d", adapter.Len())
		}
	})

	t.Run("empty history", func(t *testing.T) {
		adapter := &EventsAdapter{
			history:     nil,
			tokenBudget: 100,
		}
		if adapter.Len() != 0 {
			t.Errorf("expected 0, got %d", adapter.Len())
		}
	})

	t.Run("preserves most recent messages", func(t *testing.T) {
		var msgs []internal.Message
		for i := range 10 {
			content := ""
			for range 40 {
				content += "x"
			}
			msgs = append(msgs, internal.Message{
				Role:      "user",
				Content:   content,          // ~10 tokens + 4 overhead = 14 tokens each
				Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			})
		}
		// Budget for ~2 messages: 28 tokens
		adapter := &EventsAdapter{
			history:     msgs,
			tokenBudget: 30,
		}
		truncated := adapter.truncatedHistory()
		if len(truncated) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(truncated))
		}
		// Should be the last 2 messages
		if truncated[0].Content != msgs[8].Content {
			t.Error("expected 9th message (index 8)")
		}
		if truncated[1].Content != msgs[9].Content {
			t.Error("expected 10th message (index 9)")
		}
	})
}

func TestEventsAdapter_DefaultTokenBudget(t *testing.T) {
	var msgs []internal.Message
	for range 150 {
		msgs = append(msgs, internal.Message{
			Role:      "user",
			Content:   "msg",
			Timestamp: time.Now(),
		})
	}
	// tokenBudget=0 means use DefaultTokenBudget
	adapter := &EventsAdapter{
		history:     msgs,
		tokenBudget: 0,
	}
	// With DefaultTokenBudget (32000) and tiny messages (~1 token each),
	// all 150 messages should fit within the budget.
	if adapter.Len() != 150 {
		t.Errorf("expected all 150 messages within default budget, got %d", adapter.Len())
	}
}

// --- FunctionResponse reconstruction tests ---

func TestEventsAdapter_FunctionResponseReconstruction(t *testing.T) {
	now := time.Now()

	t.Run("new format with ToolCalls metadata", func(t *testing.T) {
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

		if len(events) != 3 {
			t.Fatalf("expected 3 events, got %d", len(events))
		}

		// Verify assistant event has FunctionCall with ID
		assistantEvt := events[1]
		if assistantEvt.LLMResponse.Content.Role != "assistant" {
			t.Errorf("expected role 'assistant', got %q", assistantEvt.LLMResponse.Content.Role)
		}
		var fc *genai.FunctionCall
		for _, p := range assistantEvt.LLMResponse.Content.Parts {
			if p.FunctionCall != nil {
				fc = p.FunctionCall
			}
		}
		if fc == nil {
			t.Fatal("expected FunctionCall part in assistant event")
		}
		if fc.ID != "adk-abc-123" {
			t.Errorf("expected FunctionCall.ID 'adk-abc-123', got %q", fc.ID)
		}
		if fc.Name != "exec" {
			t.Errorf("expected FunctionCall.Name 'exec', got %q", fc.Name)
		}

		// Verify tool event has FunctionResponse
		toolEvt := events[2]
		if toolEvt.LLMResponse.Content.Role != "function" {
			t.Errorf("expected role 'function', got %q", toolEvt.LLMResponse.Content.Role)
		}
		var fr *genai.FunctionResponse
		for _, p := range toolEvt.LLMResponse.Content.Parts {
			if p.FunctionResponse != nil {
				fr = p.FunctionResponse
			}
		}
		if fr == nil {
			t.Fatal("expected FunctionResponse part in tool event")
		}
		if fr.ID != "adk-abc-123" {
			t.Errorf("expected FunctionResponse.ID 'adk-abc-123', got %q", fr.ID)
		}
		if fr.Name != "exec" {
			t.Errorf("expected FunctionResponse.Name 'exec', got %q", fr.Name)
		}
		if fr.Response["result"] != "file.txt" {
			t.Errorf("expected response result 'file.txt', got %v", fr.Response["result"])
		}
	})

	t.Run("legacy format without ToolCalls on tool message", func(t *testing.T) {
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

		if len(events) != 3 {
			t.Fatalf("expected 3 events, got %d", len(events))
		}

		// Verify tool event has FunctionResponse reconstructed from legacy
		toolEvt := events[2]
		if toolEvt.LLMResponse.Content.Role != "function" {
			t.Errorf("expected role 'function', got %q", toolEvt.LLMResponse.Content.Role)
		}
		var fr *genai.FunctionResponse
		for _, p := range toolEvt.LLMResponse.Content.Parts {
			if p.FunctionResponse != nil {
				fr = p.FunctionResponse
			}
		}
		if fr == nil {
			t.Fatal("expected FunctionResponse part in legacy tool event")
		}
		if fr.ID != "call_exec" {
			t.Errorf("expected FunctionResponse.ID 'call_exec', got %q", fr.ID)
		}
		if fr.Name != "exec" {
			t.Errorf("expected FunctionResponse.Name 'exec', got %q", fr.Name)
		}
	})

	t.Run("tool message without preceding assistant ToolCalls falls back to text", func(t *testing.T) {
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

		if len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}

		toolEvt := events[1]
		// Should fall back to text since no context to reconstruct FunctionResponse
		hasText := false
		for _, p := range toolEvt.LLMResponse.Content.Parts {
			if p.Text != "" {
				hasText = true
			}
		}
		if !hasText {
			t.Error("expected text part in tool event without FunctionResponse context")
		}
	})
}

func TestEventsAdapter_TruncationSequenceSafety(t *testing.T) {
	t.Run("skips leading tool message after truncation", func(t *testing.T) {
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
			if first.Role == "tool" || first.Role == "function" {
				t.Error("truncated history should not start with tool/function message")
			}
		}
	})

	t.Run("does not skip trailing FunctionCall without truncation", func(t *testing.T) {
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

		if len(truncated) != 2 {
			t.Errorf("expected 2 messages (no truncation), got %d", len(truncated))
		}
	})
}

// --- SessionServiceAdapter tests ---

func TestSessionServiceAdapter_Create(t *testing.T) {
	store := newMockStore()
	service := NewSessionServiceAdapter(store, "lango-agent")

	resp, err := service.Create(context.Background(), &session.CreateRequest{
		SessionID: "new-session",
		State: map[string]any{
			"key": "value",
		},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.Session.ID() != "new-session" {
		t.Errorf("expected session ID 'new-session', got %q", resp.Session.ID())
	}

	// Verify state was set
	val, err := resp.Session.State().Get("key")
	if err != nil {
		t.Fatalf("Get state failed: %v", err)
	}
	if val != "value" {
		t.Errorf("expected 'value', got %v", val)
	}
}

func TestSessionServiceAdapter_Get(t *testing.T) {
	store := newMockStore()
	store.Create(&internal.Session{
		Key:      "existing",
		Metadata: map[string]string{"foo": "bar"},
	})

	service := NewSessionServiceAdapter(store, "lango-agent")

	resp, err := service.Get(context.Background(), &session.GetRequest{
		SessionID: "existing",
	})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if resp.Session.ID() != "existing" {
		t.Errorf("expected session ID 'existing', got %q", resp.Session.ID())
	}
}

func TestSessionServiceAdapter_GetAutoCreate(t *testing.T) {
	store := newMockStore()
	service := NewSessionServiceAdapter(store, "lango-agent")

	// Get on a nonexistent session should auto-create it
	resp, err := service.Get(context.Background(), &session.GetRequest{
		SessionID: "auto-created",
	})
	if err != nil {
		t.Fatalf("expected auto-create, got error: %v", err)
	}
	if resp.Session.ID() != "auto-created" {
		t.Fatalf("expected session ID 'auto-created', got %q", resp.Session.ID())
	}

	// Verify session now exists in store
	sess, err := store.Get("auto-created")
	if err != nil {
		t.Fatalf("expected session in store, got error: %v", err)
	}
	if sess.Key != "auto-created" {
		t.Fatalf("expected key 'auto-created', got %q", sess.Key)
	}
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
func (m *uniqueMockStore) Delete(key string) error              { return nil }
func (m *uniqueMockStore) AppendMessage(string, internal.Message) error { return nil }
func (m *uniqueMockStore) Close() error                         { return nil }
func (m *uniqueMockStore) GetSalt(string) ([]byte, error)       { return nil, nil }
func (m *uniqueMockStore) SetSalt(string, []byte) error         { return nil }

func TestSessionServiceAdapter_GetAutoCreate_Concurrent(t *testing.T) {
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
		if err != nil {
			t.Errorf("goroutine %d failed: %v", i, err)
		}
	}
}

func TestSessionServiceAdapter_Delete(t *testing.T) {
	store := newMockStore()
	store.Create(&internal.Session{Key: "to-delete"})

	service := NewSessionServiceAdapter(store, "lango-agent")

	err := service.Delete(context.Background(), &session.DeleteRequest{
		SessionID: "to-delete",
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	s, _ := store.Get("to-delete")
	if s != nil {
		t.Error("expected session to be deleted")
	}
}

func TestSessionServiceAdapter_List(t *testing.T) {
	store := newMockStore()
	service := NewSessionServiceAdapter(store, "lango-agent")

	resp, err := service.List(context.Background(), &session.ListRequest{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	// Currently returns empty
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestSessionServiceAdapter_AppendEvent_UserMessage(t *testing.T) {
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
	// Simulate user content via LLMResponse structure
	// (ADK events always carry LLMResponse)
	// For user message, content role is "user"
	// We need to import genai for this
	// Since the test is in the adk package, we can use genai directly

	err := service.AppendEvent(context.Background(), adapter, evt)
	if err != nil {
		t.Fatalf("AppendEvent failed: %v", err)
	}

	// Verify message was appended
	updated, _ := store.Get("sess-1")
	if len(updated.History) != 1 {
		t.Fatalf("expected 1 message in history, got %d", len(updated.History))
	}
	if updated.History[0].Role != "user" {
		t.Errorf("expected role 'user', got %q", updated.History[0].Role)
	}
}

// --- convertMessages tests ---

func TestConvertMessages_RoleMapping(t *testing.T) {
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
			msgs, err := convertMessages([]*genai.Content{{
				Role:  tt.give,
				Parts: []*genai.Part{{Text: "test"}},
			}})
			if err != nil {
				t.Fatalf("convertMessages failed: %v", err)
			}
			if len(msgs) != 1 {
				t.Fatalf("expected 1 message, got %d", len(msgs))
			}
			if msgs[0].Role != tt.want {
				t.Errorf("expected role %q, got %q", tt.want, msgs[0].Role)
			}
		})
	}
}

func TestConvertMessages_TextContent(t *testing.T) {
	msgs, err := convertMessages([]*genai.Content{{
		Role:  "user",
		Parts: []*genai.Part{{Text: "hello world"}},
	}})
	if err != nil {
		t.Fatalf("convertMessages failed: %v", err)
	}
	if msgs[0].Content != "hello world" {
		t.Errorf("expected 'hello world', got %q", msgs[0].Content)
	}
}

func TestConvertMessages_FunctionCall(t *testing.T) {
	msgs, err := convertMessages([]*genai.Content{{
		Role: "model",
		Parts: []*genai.Part{{
			FunctionCall: &genai.FunctionCall{
				Name: "exec",
				Args: map[string]any{"cmd": "ls"},
			},
		}},
	}})
	if err != nil {
		t.Fatalf("convertMessages failed: %v", err)
	}
	if len(msgs[0].ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(msgs[0].ToolCalls))
	}
	if msgs[0].ToolCalls[0].Name != "exec" {
		t.Errorf("expected tool name 'exec', got %q", msgs[0].ToolCalls[0].Name)
	}
}

func TestConvertMessages_FunctionResponse(t *testing.T) {
	msgs, err := convertMessages([]*genai.Content{{
		Role: "function",
		Parts: []*genai.Part{{
			FunctionResponse: &genai.FunctionResponse{
				Name:     "exec",
				Response: map[string]any{"output": "file.txt"},
			},
		}},
	}})
	if err != nil {
		t.Fatalf("convertMessages failed: %v", err)
	}
	if msgs[0].Role != "tool" {
		t.Errorf("expected role 'tool', got %q", msgs[0].Role)
	}
	if msgs[0].Content == "" {
		t.Error("expected non-empty content from function response")
	}
	if msgs[0].Metadata == nil || msgs[0].Metadata["tool_call_id"] != "exec" {
		t.Errorf("expected tool_call_id metadata, got %v", msgs[0].Metadata)
	}
}

func TestConvertMessages_Empty(t *testing.T) {
	msgs, err := convertMessages(nil)
	if err != nil {
		t.Fatalf("convertMessages failed: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages, got %d", len(msgs))
	}
}

func TestConvertTools_NilConfig(t *testing.T) {
	tools, err := convertTools(nil)
	if err != nil {
		t.Fatalf("convertTools(nil) failed: %v", err)
	}
	if len(tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(tools))
	}
}

func TestConvertTools_NilTools(t *testing.T) {
	cfg := &genai.GenerateContentConfig{}
	tools, err := convertTools(cfg)
	if err != nil {
		t.Fatalf("convertTools failed: %v", err)
	}
	if len(tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(tools))
	}
}

func TestConvertTools_WithFunctionDeclarations(t *testing.T) {
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
	if err != nil {
		t.Fatalf("convertTools failed: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Name != "test_tool" {
		t.Errorf("expected tool name 'test_tool', got %q", tools[0].Name)
	}
	if tools[0].Description != "A test tool" {
		t.Errorf("expected description 'A test tool', got %q", tools[0].Description)
	}
}
