package adk

import (
	"context"
	"testing"
	"time"

	internal "github.com/langowarny/lango/internal/session"
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

	if err := svc.AppendEvent(context.Background(), adapter, evt); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}

	// Verify in-memory history was updated
	if len(adapter.sess.History) != 1 {
		t.Fatalf("expected 1 message in history, got %d", len(adapter.sess.History))
	}
	if adapter.sess.History[0].Role != "user" {
		t.Errorf("expected role 'user', got %q", adapter.sess.History[0].Role)
	}
	if adapter.sess.History[0].Content != "hello" {
		t.Errorf("expected content 'hello', got %q", adapter.sess.History[0].Content)
	}

	// Events() should now return the message
	events := adapter.Events()
	if events.Len() != 1 {
		t.Errorf("expected Events().Len() == 1, got %d", events.Len())
	}
}

func TestAppendEvent_MultipleEvents_AccumulateHistory(t *testing.T) {
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
	if err := svc.AppendEvent(context.Background(), adapter, newTestEvent("user", "user", "hello")); err != nil {
		t.Fatalf("AppendEvent user: %v", err)
	}

	// Append assistant message
	if err := svc.AppendEvent(context.Background(), adapter, newTestEvent("lango-agent", "model", "hi there")); err != nil {
		t.Fatalf("AppendEvent assistant: %v", err)
	}

	// Verify both messages in in-memory history
	if len(adapter.sess.History) != 2 {
		t.Fatalf("expected 2 messages in history, got %d", len(adapter.sess.History))
	}
	if adapter.sess.History[0].Role != "user" {
		t.Errorf("expected first role 'user', got %q", adapter.sess.History[0].Role)
	}
	if adapter.sess.History[1].Role != "assistant" {
		t.Errorf("expected second role 'assistant', got %q", adapter.sess.History[1].Role)
	}

	// Events() should see both messages
	events := adapter.Events()
	if events.Len() != 2 {
		t.Errorf("expected Events().Len() == 2, got %d", events.Len())
	}
}

func TestAppendEvent_StateDelta_SkipsHistory(t *testing.T) {
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

	if err := svc.AppendEvent(context.Background(), adapter, evt); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}

	// State-delta-only events should not append to history
	if len(adapter.sess.History) != 0 {
		t.Errorf("expected 0 messages for state-delta event, got %d", len(adapter.sess.History))
	}
}

func TestAppendEvent_DBAndMemoryBothUpdated(t *testing.T) {
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
	if err := svc.AppendEvent(context.Background(), adapter, evt); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}

	// Verify DB store has the message
	dbMsgs := store.messages["test-session"]
	if len(dbMsgs) != 1 {
		t.Fatalf("expected 1 message in DB store, got %d", len(dbMsgs))
	}
	if dbMsgs[0].Content != "hello" {
		t.Errorf("expected DB content 'hello', got %q", dbMsgs[0].Content)
	}

	// Verify in-memory history also has the message
	if len(adapter.sess.History) != 1 {
		t.Fatalf("expected 1 message in memory, got %d", len(adapter.sess.History))
	}
}

func TestAppendEvent_PreservesAuthor(t *testing.T) {
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
	if err := svc.AppendEvent(context.Background(), adapter, evt); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}

	// Verify author was preserved in in-memory history
	if len(adapter.sess.History) != 1 {
		t.Fatalf("expected 1 message, got %d", len(adapter.sess.History))
	}
	if adapter.sess.History[0].Author != "lango-orchestrator" {
		t.Errorf("expected author 'lango-orchestrator', got %q", adapter.sess.History[0].Author)
	}

	// Verify author was preserved in DB store
	dbMsgs := store.messages["test-session"]
	if len(dbMsgs) != 1 {
		t.Fatalf("expected 1 DB message, got %d", len(dbMsgs))
	}
	if dbMsgs[0].Author != "lango-orchestrator" {
		t.Errorf("expected DB author 'lango-orchestrator', got %q", dbMsgs[0].Author)
	}
}

// Verify the LLMResponse field is unused in model import (for compile check)
var _ = model.LLMResponse{}
