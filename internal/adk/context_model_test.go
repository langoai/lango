package adk

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/adk/model"
	"google.golang.org/genai"

	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/prompt"
	"github.com/langoai/lango/internal/provider"
	"github.com/langoai/lango/internal/session"
)

// mockMemoryProvider records calls and returns canned data.
type mockMemoryProvider struct {
	lastSessionKey string
	observations   []memory.Observation
	reflections    []memory.Reflection
}

func (m *mockMemoryProvider) ListObservations(_ context.Context, sessionKey string) ([]memory.Observation, error) {
	m.lastSessionKey = sessionKey
	return m.observations, nil
}

func (m *mockMemoryProvider) ListReflections(_ context.Context, sessionKey string) ([]memory.Reflection, error) {
	m.lastSessionKey = sessionKey
	return m.reflections, nil
}

func (m *mockMemoryProvider) ListRecentReflections(_ context.Context, sessionKey string, _ int) ([]memory.Reflection, error) {
	m.lastSessionKey = sessionKey
	return m.reflections, nil
}

func (m *mockMemoryProvider) ListRecentObservations(_ context.Context, sessionKey string, _ int) ([]memory.Observation, error) {
	m.lastSessionKey = sessionKey
	return m.observations, nil
}

// Compile-time check.
var _ MemoryProvider = (*mockMemoryProvider)(nil)

func newTestContextAdapter(t *testing.T, mp MemoryProvider) *ContextAwareModelAdapter {
	t.Helper()
	p := &mockProvider{
		id: "test",
		events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "ok"},
			{Type: provider.StreamEventDone},
		},
	}
	inner := NewModelAdapter(p, "test-model")
	builder := prompt.DefaultBuilder()
	logger := zap.NewNop().Sugar()
	adapter := NewContextAwareModelAdapter(inner, nil, builder, logger)
	if mp != nil {
		adapter.WithMemory(mp)
		adapter.WithMemoryLimits(3, 5)
	}
	return adapter
}

func TestGenerateContent_SessionKeyFromContext(t *testing.T) {
	mp := &mockMemoryProvider{
		observations: []memory.Observation{{Content: "user prefers dark mode"}},
		reflections:  []memory.Reflection{{Content: "user is a developer"}},
	}
	adapter := newTestContextAdapter(t, mp)

	ctx := session.WithSessionKey(context.Background(), "telegram:123:456")
	req := &model.LLMRequest{
		Model: "test-model",
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "hello"}}},
		},
	}

	seq := adapter.GenerateContent(ctx, req, false)
	for _, err := range seq {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if mp.lastSessionKey != "telegram:123:456" {
		t.Errorf("want session key %q passed to memory provider, got %q",
			"telegram:123:456", mp.lastSessionKey)
	}
}

func TestGenerateContent_NoSessionKey_SkipsMemory(t *testing.T) {
	mp := &mockMemoryProvider{
		observations: []memory.Observation{{Content: "should not appear"}},
	}
	adapter := newTestContextAdapter(t, mp)

	// No session key in context.
	ctx := context.Background()
	req := &model.LLMRequest{
		Model: "test-model",
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "hello"}}},
		},
	}

	seq := adapter.GenerateContent(ctx, req, false)
	for _, err := range seq {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Memory provider should not have been called.
	if mp.lastSessionKey != "" {
		t.Errorf("memory provider should not be called without session key, got %q", mp.lastSessionKey)
	}
}

func TestGenerateContent_SessionKey_UpdatesRuntimeAdapter(t *testing.T) {
	adapter := newTestContextAdapter(t, nil)
	ra := NewRuntimeContextAdapter(2, false, false, true)
	adapter.WithRuntimeAdapter(ra)

	ctx := session.WithSessionKey(context.Background(), "discord:guild:chan")
	req := &model.LLMRequest{
		Model: "test-model",
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "hello"}}},
		},
	}

	seq := adapter.GenerateContent(ctx, req, false)
	for _, err := range seq {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	rc := ra.GetRuntimeContext()
	if rc.SessionKey != "discord:guild:chan" {
		t.Errorf("want runtime session key %q, got %q", "discord:guild:chan", rc.SessionKey)
	}
	if rc.ChannelType != "discord" {
		t.Errorf("want channel type %q, got %q", "discord", rc.ChannelType)
	}
}

func TestGenerateContent_MemoryInjectedIntoPrompt(t *testing.T) {
	mp := &mockMemoryProvider{
		observations: []memory.Observation{{Content: "user prefers Go"}},
		reflections:  []memory.Reflection{{Content: "experienced developer"}},
	}
	p := &mockProvider{
		id: "test",
		events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "ok"},
			{Type: provider.StreamEventDone},
		},
	}
	inner := NewModelAdapter(p, "test-model")
	builder := prompt.DefaultBuilder()
	logger := zap.NewNop().Sugar()
	adapter := NewContextAwareModelAdapter(inner, nil, builder, logger)
	adapter.WithMemory(mp)
	adapter.WithMemoryLimits(3, 5)

	ctx := session.WithSessionKey(context.Background(), "test:session:1")
	req := &model.LLMRequest{
		Model: "test-model",
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "hello"}}},
		},
	}

	seq := adapter.GenerateContent(ctx, req, false)
	for _, err := range seq {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Verify system instruction was augmented with memory content.
	msgs := p.lastParams.Messages
	if len(msgs) < 2 {
		t.Fatalf("expected at least 2 messages (system + user), got %d", len(msgs))
	}

	systemMsg := msgs[0]
	if systemMsg.Role != "system" {
		t.Fatalf("expected first message to be system, got %q", systemMsg.Role)
	}

	// The system prompt should contain memory sections.
	if !containsSubstring(systemMsg.Content, "Conversation Memory") {
		t.Error("system prompt should contain 'Conversation Memory' section")
	}
	if !containsSubstring(systemMsg.Content, "user prefers Go") {
		t.Error("system prompt should contain observation content")
	}
	if !containsSubstring(systemMsg.Content, "experienced developer") {
		t.Error("system prompt should contain reflection content")
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && contains(s, sub))
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
