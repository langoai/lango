package adk

import (
	"context"
	"iter"
	"testing"

	"github.com/langoai/lango/internal/provider"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

type mockProvider struct {
	id         string
	events     []provider.StreamEvent
	err        error
	lastParams provider.GenerateParams
}

func (m *mockProvider) ID() string { return m.id }

func (m *mockProvider) Generate(_ context.Context, params provider.GenerateParams) (iter.Seq2[provider.StreamEvent, error], error) {
	m.lastParams = params
	if m.err != nil {
		return nil, m.err
	}
	return func(yield func(provider.StreamEvent, error) bool) {
		for _, evt := range m.events {
			if !yield(evt, nil) {
				return
			}
		}
	}, nil
}

func (m *mockProvider) ListModels(_ context.Context) ([]provider.ModelInfo, error) {
	return nil, nil
}

func TestModelAdapter_Name(t *testing.T) {
	p := &mockProvider{id: "test-provider"}
	adapter := NewModelAdapter(p, "test-model")

	if adapter.Name() != "test-model" {
		t.Errorf("expected 'test-model', got %q", adapter.Name())
	}
}

func TestModelAdapter_GenerateContent_TextDelta(t *testing.T) {
	p := &mockProvider{
		id: "test",
		events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "Hello "},
			{Type: provider.StreamEventPlainText, Text: "world"},
			{Type: provider.StreamEventDone},
		},
	}
	adapter := NewModelAdapter(p, "test-model")

	req := &model.LLMRequest{Model: "test-model"}
	seq := adapter.GenerateContent(context.Background(), req, true)

	var responses []*model.LLMResponse
	for resp, err := range seq {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		responses = append(responses, resp)
	}

	if len(responses) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(responses))
	}

	// First two should be partial text
	if !responses[0].Partial {
		t.Error("expected first response to be partial")
	}
	if responses[0].Content.Parts[0].Text != "Hello " {
		t.Errorf("expected 'Hello ', got %q", responses[0].Content.Parts[0].Text)
	}

	// Last should be turn complete
	if !responses[2].TurnComplete {
		t.Error("expected last response to be turn complete")
	}
	if responses[2].Partial {
		t.Error("expected last response to not be partial")
	}
}

func TestModelAdapter_GenerateContent_ProviderError(t *testing.T) {
	p := &mockProvider{
		id:  "test",
		err: context.DeadlineExceeded,
	}
	adapter := NewModelAdapter(p, "test-model")

	req := &model.LLMRequest{Model: "test-model"}
	seq := adapter.GenerateContent(context.Background(), req, false)

	for _, err := range seq {
		if err == nil {
			t.Fatal("expected error from provider")
		}
		return // Only check first yield
	}
	t.Fatal("expected at least one yield")
}

func TestModelAdapter_GenerateContent_ToolCall(t *testing.T) {
	p := &mockProvider{
		id: "test",
		events: []provider.StreamEvent{
			{
				Type: provider.StreamEventToolCall,
				ToolCall: &provider.ToolCall{
					ID:        "call_1",
					Name:      "exec",
					Arguments: `{"command":"ls"}`,
				},
			},
			{Type: provider.StreamEventDone},
		},
	}
	adapter := NewModelAdapter(p, "test-model")

	req := &model.LLMRequest{Model: "test-model"}
	seq := adapter.GenerateContent(context.Background(), req, false)

	var responses []*model.LLMResponse
	for resp, err := range seq {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		responses = append(responses, resp)
	}

	// Non-streaming mode accumulates all events into a single response.
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}

	resp := responses[0]
	if !resp.TurnComplete {
		t.Error("expected response to be turn complete")
	}
	if resp.Partial {
		t.Error("expected response to not be partial")
	}

	// Should have the function call part.
	hasFuncCall := false
	for _, p := range resp.Content.Parts {
		if p.FunctionCall != nil {
			hasFuncCall = true
			if p.FunctionCall.Name != "exec" {
				t.Errorf("expected function name 'exec', got %q", p.FunctionCall.Name)
			}
			if p.FunctionCall.Args["command"] != "ls" {
				t.Errorf("expected arg command='ls', got %v", p.FunctionCall.Args["command"])
			}
		}
	}
	if !hasFuncCall {
		t.Error("expected a FunctionCall part")
	}
}

func TestModelAdapter_GenerateContent_StreamError(t *testing.T) {
	p := &mockProvider{
		id: "test",
		events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "partial"},
			{Type: provider.StreamEventError, Error: context.Canceled},
		},
	}
	adapter := NewModelAdapter(p, "test-model")

	req := &model.LLMRequest{Model: "test-model"}
	seq := adapter.GenerateContent(context.Background(), req, true)

	gotError := false
	for _, err := range seq {
		if err != nil {
			gotError = true
			break
		}
	}
	if !gotError {
		t.Error("expected error event to propagate")
	}
}

func TestModelAdapter_GenerateContent_SystemInstruction(t *testing.T) {
	p := &mockProvider{
		id: "test",
		events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "response"},
			{Type: provider.StreamEventDone},
		},
	}
	adapter := NewModelAdapter(p, "test-model")

	req := &model.LLMRequest{
		Model: "test-model",
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "hello"}}},
		},
		Config: &genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{
					{Text: "You are a helpful assistant."},
					{Text: "Always be concise."},
				},
			},
		},
	}
	seq := adapter.GenerateContent(context.Background(), req, false)

	for _, err := range seq {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Verify system message is prepended to messages
	msgs := p.lastParams.Messages
	if len(msgs) < 2 {
		t.Fatalf("expected at least 2 messages (system + user), got %d", len(msgs))
	}
	if msgs[0].Role != "system" {
		t.Errorf("expected first message role 'system', got %q", msgs[0].Role)
	}
	if msgs[0].Content != "You are a helpful assistant.\nAlways be concise." {
		t.Errorf("unexpected system content: %q", msgs[0].Content)
	}
	if msgs[1].Role != "user" {
		t.Errorf("expected second message role 'user', got %q", msgs[1].Role)
	}
}

func TestModelAdapter_GenerateContent_NoSystemInstruction(t *testing.T) {
	p := &mockProvider{
		id: "test",
		events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "response"},
			{Type: provider.StreamEventDone},
		},
	}
	adapter := NewModelAdapter(p, "test-model")

	req := &model.LLMRequest{
		Model: "test-model",
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "hello"}}},
		},
	}
	seq := adapter.GenerateContent(context.Background(), req, false)

	for _, err := range seq {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Without system instruction, only the user message should be present
	msgs := p.lastParams.Messages
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Role != "user" {
		t.Errorf("expected role 'user', got %q", msgs[0].Role)
	}
}
