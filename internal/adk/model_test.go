package adk

import (
	"context"
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	t.Parallel()

	p := &mockProvider{id: "test-provider"}
	adapter := NewModelAdapter(p, "test-model")

	assert.Equal(t, "test-model", adapter.Name())
}

func TestModelAdapter_GenerateContent_TextDelta(t *testing.T) {
	t.Parallel()

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
		require.NoError(t, err)
		responses = append(responses, resp)
	}

	require.Len(t, responses, 3)

	// First two should be partial text
	assert.True(t, responses[0].Partial, "expected first response to be partial")
	assert.Equal(t, "Hello ", responses[0].Content.Parts[0].Text)

	// Last should be turn complete
	assert.True(t, responses[2].TurnComplete, "expected last response to be turn complete")
	assert.False(t, responses[2].Partial, "expected last response to not be partial")
}

func TestModelAdapter_GenerateContent_ProviderError(t *testing.T) {
	t.Parallel()

	p := &mockProvider{
		id:  "test",
		err: context.DeadlineExceeded,
	}
	adapter := NewModelAdapter(p, "test-model")

	req := &model.LLMRequest{Model: "test-model"}
	seq := adapter.GenerateContent(context.Background(), req, false)

	for _, err := range seq {
		require.Error(t, err, "expected error from provider")
		return // Only check first yield
	}
	t.Fatal("expected at least one yield")
}

func TestModelAdapter_GenerateContent_ToolCall(t *testing.T) {
	t.Parallel()

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
		require.NoError(t, err)
		responses = append(responses, resp)
	}

	// Non-streaming mode accumulates all events into a single response.
	require.Len(t, responses, 1)

	resp := responses[0]
	assert.True(t, resp.TurnComplete, "expected response to be turn complete")
	assert.False(t, resp.Partial, "expected response to not be partial")

	// Should have the function call part.
	hasFuncCall := false
	for _, p := range resp.Content.Parts {
		if p.FunctionCall != nil {
			hasFuncCall = true
			assert.Equal(t, "exec", p.FunctionCall.Name)
			assert.Equal(t, "ls", p.FunctionCall.Args["command"])
		}
	}
	assert.True(t, hasFuncCall, "expected a FunctionCall part")
}

func TestModelAdapter_GenerateContent_StreamError(t *testing.T) {
	t.Parallel()

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
	assert.True(t, gotError, "expected error event to propagate")
}

func TestModelAdapter_GenerateContent_SystemInstruction(t *testing.T) {
	t.Parallel()

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
		require.NoError(t, err)
	}

	// Verify system message is prepended to messages
	msgs := p.lastParams.Messages
	require.GreaterOrEqual(t, len(msgs), 2, "expected at least 2 messages (system + user)")
	assert.Equal(t, "system", string(msgs[0].Role))
	assert.Equal(t, "You are a helpful assistant.\nAlways be concise.", msgs[0].Content)
	assert.Equal(t, "user", string(msgs[1].Role))
}

func TestModelAdapter_GenerateContent_NoSystemInstruction(t *testing.T) {
	t.Parallel()

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
		require.NoError(t, err)
	}

	// Without system instruction, only the user message should be present
	msgs := p.lastParams.Messages
	require.Len(t, msgs, 1)
	assert.Equal(t, "user", string(msgs[0].Role))
}
