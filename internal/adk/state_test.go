package adk

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	internal "github.com/langowarny/lango/internal/session"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

type mockStore struct {
	sessions map[string]*internal.Session
}

func newMockStore() *mockStore {
	return &mockStore{sessions: make(map[string]*internal.Session)}
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
	s, ok := m.sessions[key]
	if ok {
		s.History = append(s.History, msg)
	}
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

	adapter := NewSessionAdapter(sess, store)
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

	adapter := NewSessionAdapter(sess, store)
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

	adapter := NewSessionAdapter(sess, store)
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

	adapter := NewSessionAdapter(sess, store)
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
	adapter := NewSessionAdapter(sess, store)

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

	adapter := NewSessionAdapter(sess, &mockStore{})
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

	adapter := NewSessionAdapter(sess, &mockStore{})
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

func TestEventsAdapter_Truncation(t *testing.T) {
	// Create 150 messages
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
	adapter := NewSessionAdapter(sess, &mockStore{})
	events := adapter.Events()

	// Len should be capped at 100
	if events.Len() != 100 {
		t.Errorf("expected Len=100, got %d", events.Len())
	}

	// Count events from All()
	count := 0
	for range events.All() {
		count++
	}
	if count != 100 {
		t.Errorf("expected 100 events from All(), got %d", count)
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

	adapter := NewSessionAdapter(sess, &mockStore{})
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
	adapter := NewSessionAdapter(sess, &mockStore{})
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

	adapter := NewSessionAdapter(sess, &mockStore{})
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

func TestEventsAdapter_LegacyFallback(t *testing.T) {
	var msgs []internal.Message
	for range 150 {
		msgs = append(msgs, internal.Message{
			Role:      "user",
			Content:   "msg",
			Timestamp: time.Now(),
		})
	}
	// tokenBudget=0 means legacy mode
	adapter := &EventsAdapter{
		history:     msgs,
		tokenBudget: 0,
	}
	if adapter.Len() != 100 {
		t.Errorf("expected legacy 100-cap, got %d", adapter.Len())
	}
}

// --- SessionServiceAdapter tests ---

func TestSessionServiceAdapter_Create(t *testing.T) {
	store := newMockStore()
	service := NewSessionServiceAdapter(store)

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

	service := NewSessionServiceAdapter(store)

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
	service := NewSessionServiceAdapter(store)

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

func TestSessionServiceAdapter_Delete(t *testing.T) {
	store := newMockStore()
	store.Create(&internal.Session{Key: "to-delete"})

	service := NewSessionServiceAdapter(store)

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
	service := NewSessionServiceAdapter(store)

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

	service := NewSessionServiceAdapter(store)
	adapter := NewSessionAdapter(sess, store)

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
