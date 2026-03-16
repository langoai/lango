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

func intPtr(v int) *int { return &v }

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

// --- toolCallAccumulator tests ---

func TestToolCallAccumulator_SingleComplete(t *testing.T) {
	t.Parallel()
	var acc toolCallAccumulator
	acc.add(&provider.ToolCall{
		Index:     intPtr(0),
		ID:        "call_1",
		Name:      "exec",
		Arguments: `{"command":"ls"}`,
	})

	parts := acc.done()
	require.Len(t, parts, 1)
	assert.Equal(t, "exec", parts[0].FunctionCall.Name)
	assert.Equal(t, "call_1", parts[0].FunctionCall.ID)
	assert.Equal(t, "ls", parts[0].FunctionCall.Args["command"])
}

func TestToolCallAccumulator_OpenAIStreaming(t *testing.T) {
	t.Parallel()
	var acc toolCallAccumulator

	// First chunk: Index=0, ID+Name present, partial args
	acc.add(&provider.ToolCall{
		Index:     intPtr(0),
		ID:        "call_abc",
		Name:      "exec",
		Arguments: `{"comma`,
	})
	// Second chunk: Index=0, only args continuation
	acc.add(&provider.ToolCall{
		Index:     intPtr(0),
		Arguments: `nd":"ls"}`,
	})

	parts := acc.done()
	require.Len(t, parts, 1)
	assert.Equal(t, "exec", parts[0].FunctionCall.Name)
	assert.Equal(t, "call_abc", parts[0].FunctionCall.ID)
	assert.Equal(t, "ls", parts[0].FunctionCall.Args["command"])
}

func TestToolCallAccumulator_OpenAIMultipleCalls(t *testing.T) {
	t.Parallel()
	var acc toolCallAccumulator

	// Interleaved chunks for two different tool calls
	acc.add(&provider.ToolCall{Index: intPtr(0), ID: "c1", Name: "exec", Arguments: `{"a`})
	acc.add(&provider.ToolCall{Index: intPtr(1), ID: "c2", Name: "read", Arguments: `{"p`})
	acc.add(&provider.ToolCall{Index: intPtr(0), Arguments: `":"1"}`})
	acc.add(&provider.ToolCall{Index: intPtr(1), Arguments: `":"2"}`})

	parts := acc.done()
	require.Len(t, parts, 2)

	// Sorted by index
	assert.Equal(t, "exec", parts[0].FunctionCall.Name)
	assert.Equal(t, "c1", parts[0].FunctionCall.ID)
	assert.Equal(t, "1", parts[0].FunctionCall.Args["a"])

	assert.Equal(t, "read", parts[1].FunctionCall.Name)
	assert.Equal(t, "c2", parts[1].FunctionCall.ID)
	assert.Equal(t, "2", parts[1].FunctionCall.Args["p"])
}

func TestToolCallAccumulator_AnthropicStreaming(t *testing.T) {
	t.Parallel()
	var acc toolCallAccumulator

	// Anthropic start: ID+Name, no Index
	acc.add(&provider.ToolCall{ID: "tool_1", Name: "exec"})
	// Anthropic delta: only args, no Index/ID/Name
	acc.add(&provider.ToolCall{Arguments: `{"command":"ls"}`})

	parts := acc.done()
	require.Len(t, parts, 1)
	assert.Equal(t, "exec", parts[0].FunctionCall.Name)
	assert.Equal(t, "tool_1", parts[0].FunctionCall.ID)
	assert.Equal(t, "ls", parts[0].FunctionCall.Args["command"])
}

func TestToolCallAccumulator_AnthropicMultipleCalls(t *testing.T) {
	t.Parallel()
	var acc toolCallAccumulator

	// First tool
	acc.add(&provider.ToolCall{ID: "tool_1", Name: "exec"})
	acc.add(&provider.ToolCall{Arguments: `{"a":"1"}`})
	// Second tool
	acc.add(&provider.ToolCall{ID: "tool_2", Name: "read"})
	acc.add(&provider.ToolCall{Arguments: `{"b":"2"}`})

	parts := acc.done()
	require.Len(t, parts, 2)
	assert.Equal(t, "exec", parts[0].FunctionCall.Name)
	assert.Equal(t, "1", parts[0].FunctionCall.Args["a"])
	assert.Equal(t, "read", parts[1].FunctionCall.Name)
	assert.Equal(t, "2", parts[1].FunctionCall.Args["b"])
}

func TestToolCallAccumulator_OrphanDeltaDropped(t *testing.T) {
	t.Parallel()
	var acc toolCallAccumulator

	// Delta with no preceding start — should be dropped
	acc.add(&provider.ToolCall{Arguments: `{"x":"y"}`})

	parts := acc.done()
	assert.Empty(t, parts)
}

func TestToolCallAccumulator_EmptyNameDropped(t *testing.T) {
	t.Parallel()
	var acc toolCallAccumulator

	// Entry with Index but no Name ever set
	acc.add(&provider.ToolCall{Index: intPtr(0), ID: "call_x", Arguments: `{"a":"b"}`})

	parts := acc.done()
	assert.Empty(t, parts)
}

func TestToolCallAccumulator_IDPreserved(t *testing.T) {
	t.Parallel()
	var acc toolCallAccumulator

	acc.add(&provider.ToolCall{Index: intPtr(0), ID: "call_custom_id", Name: "my_tool", Arguments: `{}`})

	parts := acc.done()
	require.Len(t, parts, 1)
	assert.Equal(t, "call_custom_id", parts[0].FunctionCall.ID)
}

func TestGenerateContent_StreamingToolCallRegression(t *testing.T) {
	t.Parallel()

	// Simulate OpenAI streaming pattern: first chunk has Name+ID, subsequent
	// chunks only have partial arguments. Previously each chunk was stored as
	// a separate FunctionCall part, causing empty-name parts in the session.
	p := &mockProvider{
		id: "test",
		events: []provider.StreamEvent{
			{
				Type: provider.StreamEventToolCall,
				ToolCall: &provider.ToolCall{
					Index:     intPtr(0),
					ID:        "call_abc",
					Name:      "exec",
					Arguments: `{"comma`,
				},
			},
			{
				Type: provider.StreamEventToolCall,
				ToolCall: &provider.ToolCall{
					Index:     intPtr(0),
					Arguments: `nd":"l`,
				},
			},
			{
				Type: provider.StreamEventToolCall,
				ToolCall: &provider.ToolCall{
					Index:     intPtr(0),
					Arguments: `s"}`,
				},
			},
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

	// Expect: 1 partial yield (first chunk with Name) + 1 final done
	require.Len(t, responses, 2, "expected 1 partial + 1 done response")

	// First response: partial tool call yield (has Name)
	partial := responses[0]
	require.NotNil(t, partial.Content)
	require.Len(t, partial.Content.Parts, 1)
	assert.Equal(t, "exec", partial.Content.Parts[0].FunctionCall.Name)
	assert.Equal(t, "call_abc", partial.Content.Parts[0].FunctionCall.ID)

	// Final response: complete, assembled tool call
	final := responses[1]
	assert.True(t, final.TurnComplete)
	assert.False(t, final.Partial)

	// Verify no empty-name parts
	for _, p := range final.Content.Parts {
		if p.FunctionCall != nil {
			assert.NotEmpty(t, p.FunctionCall.Name, "final response must not have empty function call name")
		}
	}

	// Verify assembled result
	require.Len(t, final.Content.Parts, 1)
	fc := final.Content.Parts[0].FunctionCall
	assert.Equal(t, "exec", fc.Name)
	assert.Equal(t, "call_abc", fc.ID)
	assert.Equal(t, "ls", fc.Args["command"])
}

func TestConvertMessages_EmptyFunctionCallName(t *testing.T) {
	t.Parallel()

	contents := []*genai.Content{
		{
			Role: "model",
			Parts: []*genai.Part{
				{FunctionCall: &genai.FunctionCall{ID: "call_1", Name: "valid", Args: map[string]any{"a": "b"}}},
				{FunctionCall: &genai.FunctionCall{ID: "call_2", Name: "", Args: map[string]any{"c": "d"}}},
			},
		},
	}

	msgs, err := convertMessages(contents)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	// Only the valid tool call should be present
	assert.Len(t, msgs[0].ToolCalls, 1)
	assert.Equal(t, "valid", msgs[0].ToolCalls[0].Name)
}

func TestConvertMessages_OrphanedFunctionCall_NoRepair(t *testing.T) {
	t.Parallel()

	// Assistant FunctionCall followed by user message without tool response.
	// convertMessages no longer repairs orphans — that is provider-specific.
	contents := []*genai.Content{
		{
			Role: "model",
			Parts: []*genai.Part{{
				FunctionCall: &genai.FunctionCall{
					ID:   "call_orphan",
					Name: "exec",
					Args: map[string]any{"cmd": "ls"},
				},
			}},
		},
		{
			Role:  "user",
			Parts: []*genai.Part{{Text: "retry please"}},
		},
	}

	msgs, err := convertMessages(contents)
	require.NoError(t, err)

	// Should pass through as-is: assistant + user = 2 messages (no synthetic injection)
	require.Len(t, msgs, 2, "expected no synthetic tool response from shared convertMessages")
	assert.Equal(t, "assistant", msgs[0].Role)
	assert.Equal(t, "user", msgs[1].Role)
}

func TestConvertMessages_MatchedFunctionCall(t *testing.T) {
	t.Parallel()

	// Assistant FunctionCall with matching tool response — no injection needed
	contents := []*genai.Content{
		{
			Role: "model",
			Parts: []*genai.Part{{
				FunctionCall: &genai.FunctionCall{
					ID:   "call_matched",
					Name: "exec",
					Args: map[string]any{"cmd": "ls"},
				},
			}},
		},
		{
			Role: "function",
			Parts: []*genai.Part{{
				FunctionResponse: &genai.FunctionResponse{
					ID:       "call_matched",
					Name:     "exec",
					Response: map[string]any{"output": "file.txt"},
				},
			}},
		},
		{
			Role:  "user",
			Parts: []*genai.Part{{Text: "thanks"}},
		},
	}

	msgs, err := convertMessages(contents)
	require.NoError(t, err)

	// Should remain 3 messages — no injection
	require.Len(t, msgs, 3, "expected no synthetic injection for matched call")
	assert.Equal(t, "assistant", msgs[0].Role)
	assert.Equal(t, "tool", msgs[1].Role)
	assert.Equal(t, "call_matched", msgs[1].Metadata["tool_call_id"])
	assert.Equal(t, "exec", msgs[1].Metadata["tool_call_name"])
	assert.Equal(t, "user", msgs[2].Role)
}

func TestConvertMessages_PendingFunctionCallNotTouched(t *testing.T) {
	t.Parallel()

	// Assistant FunctionCall at end of history — pending, should not be touched
	contents := []*genai.Content{
		{
			Role:  "user",
			Parts: []*genai.Part{{Text: "run ls"}},
		},
		{
			Role: "model",
			Parts: []*genai.Part{{
				FunctionCall: &genai.FunctionCall{
					ID:   "call_pending",
					Name: "exec",
					Args: map[string]any{"cmd": "ls"},
				},
			}},
		},
	}

	msgs, err := convertMessages(contents)
	require.NoError(t, err)

	// Should remain 2 messages — pending call at end is untouched
	require.Len(t, msgs, 2, "expected pending FunctionCall at end to be untouched")
	assert.Equal(t, "user", msgs[0].Role)
	assert.Equal(t, "assistant", msgs[1].Role)
}

// TestRepairOrphanedFunctionCalls_PartialResponse moved to openai_test.go as
// repairOrphanedToolCalls is now an OpenAI-specific private helper.

func TestConvertTools_EmptyName(t *testing.T) {
	t.Parallel()

	cfg := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{
			{
				FunctionDeclarations: []*genai.FunctionDeclaration{
					{Name: "valid_tool", Description: "A valid tool"},
					{Name: "", Description: "Invalid tool with empty name"},
				},
			},
		},
	}

	tools, err := convertTools(cfg)
	require.NoError(t, err)
	assert.Len(t, tools, 1)
	assert.Equal(t, "valid_tool", tools[0].Name)
}
