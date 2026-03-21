package openai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/provider"
)

func TestNewProvider(t *testing.T) {
	p := NewProvider("openai", "test-key", "http://localhost:1234")
	if p.ID() != "openai" {
		t.Errorf("expected ID 'openai', got %s", p.ID())
	}
}

func TestOpenAIProvider_ListModels(t *testing.T) {
	// ListModels calls the real API, so we test that it returns an error
	// when the server is not available (invalid base URL)
	p := NewProvider("openai", "test-key", "http://localhost:1/v1")
	_, err := p.ListModels(context.Background())
	if err == nil {
		t.Error("expected error when connecting to unavailable server")
	}
}

func TestConvertParams_EmptyToolNameFiltered(t *testing.T) {
	t.Parallel()

	p := NewProvider("openai", "test-key", "")
	params := provider.GenerateParams{
		Model: "gpt-4",
		Messages: []provider.Message{
			{Role: "user", Content: "hello"},
		},
		Tools: []provider.Tool{
			{Name: "valid_tool", Description: "A valid tool", Parameters: map[string]interface{}{"type": "object"}},
			{Name: "", Description: "Tool with empty name"},
			{Name: "another_tool", Description: "Another valid tool"},
		},
	}

	req, err := p.convertParams(params)
	require.NoError(t, err)
	require.Len(t, req.Tools, 2)
	assert.Equal(t, "valid_tool", req.Tools[0].Function.Name)
	assert.Equal(t, "another_tool", req.Tools[1].Function.Name)
}

func TestConvertParams_EmptyToolCallNameFiltered(t *testing.T) {
	t.Parallel()

	p := NewProvider("openai", "test-key", "")
	params := provider.GenerateParams{
		Model: "gpt-4",
		Messages: []provider.Message{
			{
				Role: "assistant",
				ToolCalls: []provider.ToolCall{
					{ID: "call_1", Name: "valid_call", Arguments: `{"a":"b"}`},
					{ID: "call_2", Name: "", Arguments: `{"c":"d"}`},
				},
			},
		},
	}

	req, err := p.convertParams(params)
	require.NoError(t, err)
	require.Len(t, req.Messages, 1)
	require.Len(t, req.Messages[0].ToolCalls, 1)
	assert.Equal(t, "valid_call", req.Messages[0].ToolCalls[0].Function.Name)
	assert.Equal(t, "call_1", req.Messages[0].ToolCalls[0].ID)
}

func TestConvertParams_ValidToolsUnchanged(t *testing.T) {
	t.Parallel()

	p := NewProvider("openai", "test-key", "")
	params := provider.GenerateParams{
		Model: "gpt-4",
		Messages: []provider.Message{
			{Role: "user", Content: "hello"},
		},
		Tools: []provider.Tool{
			{Name: "tool_a", Description: "Tool A"},
			{Name: "tool_b", Description: "Tool B"},
		},
	}

	req, err := p.convertParams(params)
	require.NoError(t, err)
	require.Len(t, req.Tools, 2)
	assert.Equal(t, "tool_a", req.Tools[0].Function.Name)
	assert.Equal(t, "tool_b", req.Tools[1].Function.Name)
}

// --- canUseStrictMode tests ---

func TestCanUseStrictMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		params map[string]interface{}
		want bool
	}{
		{
			give: "all properties required with additionalProperties false",
			params: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{"type": "string"},
				},
				"required":             []string{"command"},
				"additionalProperties": false,
			},
			want: true,
		},
		{
			give: "optional property exists",
			params: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{"type": "string"},
					"verbose": map[string]interface{}{"type": "boolean"},
				},
				"required":             []string{"command"},
				"additionalProperties": false,
			},
			want: false,
		},
		{
			give: "no additionalProperties field",
			params: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{"type": "string"},
				},
				"required": []string{"command"},
			},
			want: false,
		},
		{
			give: "additionalProperties true",
			params: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{"type": "string"},
				},
				"required":             []string{"command"},
				"additionalProperties": true,
			},
			want: false,
		},
		{
			give: "nil params",
			params: nil,
			want: false,
		},
		{
			give: "no properties with additionalProperties false",
			params: map[string]interface{}{
				"type":                 "object",
				"additionalProperties": false,
			},
			want: true,
		},
		{
			give: "required as []interface{}",
			params: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{"type": "string"},
				},
				"required":             []interface{}{"command"},
				"additionalProperties": false,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := canUseStrictMode(tt.params)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertParams_StrictMode(t *testing.T) {
	t.Parallel()

	p := NewProvider("openai", "test-key", "")
	params := provider.GenerateParams{
		Model: "gpt-4",
		Messages: []provider.Message{
			{Role: "user", Content: "hello"},
		},
		Tools: []provider.Tool{
			{
				Name:        "strict_tool",
				Description: "All params required",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{"type": "string"},
					},
					"required":             []string{"command"},
					"additionalProperties": false,
				},
			},
			{
				Name:        "non_strict_tool",
				Description: "Has optional param",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{"type": "string"},
						"verbose": map[string]interface{}{"type": "boolean"},
					},
					"required":             []string{"command"},
					"additionalProperties": false,
				},
			},
		},
	}

	req, err := p.convertParams(params)
	require.NoError(t, err)
	require.Len(t, req.Tools, 2)

	assert.True(t, req.Tools[0].Function.Strict, "strict_tool should have Strict=true")
	assert.False(t, req.Tools[1].Function.Strict, "non_strict_tool should have Strict=false")
}

// --- repairOrphanedToolCalls tests ---

func TestRepairOrphanedToolCalls_OrphanedCallGetsRepaired(t *testing.T) {
	t.Parallel()

	msgs := []provider.Message{
		{
			Role: "assistant",
			ToolCalls: []provider.ToolCall{
				{ID: "call_orphan", Name: "exec", Arguments: `{"cmd":"ls"}`},
			},
		},
		{
			Role:    "user",
			Content: "retry please",
		},
	}

	result := repairOrphanedToolCalls(msgs)

	require.Len(t, result, 3, "expected synthetic tool response injected")
	assert.Equal(t, "assistant", result[0].Role)
	assert.Equal(t, "tool", result[1].Role)
	assert.Equal(t, "call_orphan", result[1].Metadata["tool_call_id"])
	assert.Contains(t, result[1].Content, "interrupted")
	assert.Equal(t, "user", result[2].Role)
}

func TestRepairOrphanedToolCalls_PartialResponse(t *testing.T) {
	t.Parallel()

	// Assistant with 2 FunctionCalls, but only 1 has a tool response
	msgs := []provider.Message{
		{
			Role: "assistant",
			ToolCalls: []provider.ToolCall{
				{ID: "call_a", Name: "exec", Arguments: `{"cmd":"ls"}`},
				{ID: "call_b", Name: "read", Arguments: `{"path":"foo"}`},
			},
		},
		{
			Role:     "tool",
			Content:  `{"result":"ok"}`,
			Metadata: map[string]interface{}{"tool_call_id": "call_a"},
		},
		{
			Role:    "user",
			Content: "next",
		},
	}

	result := repairOrphanedToolCalls(msgs)

	// Should inject synthetic response for call_b only.
	require.Len(t, result, 4, "expected synthetic response for unanswered call_b")
	assert.Equal(t, "assistant", result[0].Role)
	assert.Equal(t, "tool", result[1].Role)
	assert.Equal(t, "call_b", result[1].Metadata["tool_call_id"])
	assert.Contains(t, result[1].Content, "interrupted")
	assert.Equal(t, "tool", result[2].Role)
	assert.Equal(t, "call_a", result[2].Metadata["tool_call_id"])
	assert.Equal(t, "user", result[3].Role)
}

func TestRepairOrphanedToolCalls_TrailingPendingUntouched(t *testing.T) {
	t.Parallel()

	// Pending call at end of history — should not be touched
	msgs := []provider.Message{
		{Role: "user", Content: "run ls"},
		{
			Role: "assistant",
			ToolCalls: []provider.ToolCall{
				{ID: "call_pending", Name: "exec", Arguments: `{"cmd":"ls"}`},
			},
		},
	}

	result := repairOrphanedToolCalls(msgs)

	require.Len(t, result, 2, "expected pending FunctionCall at end to be untouched")
	assert.Equal(t, "user", result[0].Role)
	assert.Equal(t, "assistant", result[1].Role)
}

func TestRepairOrphanedToolCalls_MatchedCallNoRepair(t *testing.T) {
	t.Parallel()

	msgs := []provider.Message{
		{
			Role: "assistant",
			ToolCalls: []provider.ToolCall{
				{ID: "call_ok", Name: "exec", Arguments: `{"cmd":"ls"}`},
			},
		},
		{
			Role:     "tool",
			Content:  `{"result":"files"}`,
			Metadata: map[string]interface{}{"tool_call_id": "call_ok"},
		},
		{Role: "user", Content: "thanks"},
	}

	result := repairOrphanedToolCalls(msgs)

	require.Len(t, result, 3, "expected no synthetic injection for matched call")
}

func TestConvertParams_ThoughtCallsFiltered(t *testing.T) {
	t.Parallel()

	p := NewProvider("openai", "test-key", "")
	params := provider.GenerateParams{
		Model: "gpt-4",
		Messages: []provider.Message{
			{Role: "user", Content: "hello"},
			{
				Role: "assistant",
				ToolCalls: []provider.ToolCall{
					{ID: "call_thought", Name: "think", Arguments: `{}`, Thought: true, ThoughtSignature: []byte("sig")},
					{ID: "call_real", Name: "exec", Arguments: `{"cmd":"ls"}`},
				},
			},
			{
				Role:     "tool",
				Content:  `{"thought":"internal"}`,
				Metadata: map[string]interface{}{"tool_call_id": "call_thought"},
			},
			{
				Role:     "tool",
				Content:  `{"result":"files"}`,
				Metadata: map[string]interface{}{"tool_call_id": "call_real"},
			},
			{Role: "user", Content: "thanks"},
		},
	}

	req, err := p.convertParams(params)
	require.NoError(t, err)

	// The thought tool call and its response should be filtered out.
	// Expected: user + assistant(1 tool call) + tool(call_real) + user = 4 messages.
	require.Len(t, req.Messages, 4, "thought call and its response should be dropped")

	// Assistant should have only the real tool call.
	assert.Len(t, req.Messages[1].ToolCalls, 1)
	assert.Equal(t, "call_real", req.Messages[1].ToolCalls[0].ID)

	// The remaining tool response should be for call_real.
	assert.Equal(t, "tool", req.Messages[2].Role)
	assert.Equal(t, "call_real", req.Messages[2].ToolCallID)
}

func TestConvertParams_ThoughtCallsAllDropped(t *testing.T) {
	t.Parallel()

	p := NewProvider("openai", "test-key", "")
	params := provider.GenerateParams{
		Model: "gpt-4",
		Messages: []provider.Message{
			{Role: "user", Content: "hello"},
			{
				Role: "assistant",
				ToolCalls: []provider.ToolCall{
					{ID: "call_t1", Name: "think", Arguments: `{}`, Thought: true},
				},
			},
			{
				Role:     "tool",
				Content:  `{}`,
				Metadata: map[string]interface{}{"tool_call_id": "call_t1"},
			},
			{Role: "user", Content: "ok"},
		},
	}

	req, err := p.convertParams(params)
	require.NoError(t, err)

	// Both thought call and response dropped. Assistant with no tool calls remains.
	// Expected: user + assistant(no tool calls) + user = 3 messages.
	require.Len(t, req.Messages, 3)
	assert.Len(t, req.Messages[1].ToolCalls, 0)
}

func TestConvertParams_NoThoughtCallsUnchanged(t *testing.T) {
	t.Parallel()

	p := NewProvider("openai", "test-key", "")
	params := provider.GenerateParams{
		Model: "gpt-4",
		Messages: []provider.Message{
			{Role: "user", Content: "hello"},
			{
				Role: "assistant",
				ToolCalls: []provider.ToolCall{
					{ID: "call_1", Name: "exec", Arguments: `{"cmd":"ls"}`},
				},
			},
			{
				Role:     "tool",
				Content:  `{"result":"files"}`,
				Metadata: map[string]interface{}{"tool_call_id": "call_1"},
			},
			{Role: "user", Content: "thanks"},
		},
	}

	req, err := p.convertParams(params)
	require.NoError(t, err)
	require.Len(t, req.Messages, 4, "no thought calls means no filtering")
	assert.Len(t, req.Messages[1].ToolCalls, 1)
	assert.Equal(t, "call_1", req.Messages[1].ToolCalls[0].ID)
}

func TestConvertParams_RepairIntegration(t *testing.T) {
	t.Parallel()

	p := NewProvider("openai", "test-key", "")
	params := provider.GenerateParams{
		Model: "gpt-4",
		Messages: []provider.Message{
			{
				Role: "assistant",
				ToolCalls: []provider.ToolCall{
					{ID: "call_1", Name: "exec", Arguments: `{"cmd":"ls"}`},
				},
			},
			{Role: "user", Content: "retry"},
		},
	}

	req, err := p.convertParams(params)
	require.NoError(t, err)

	// convertParams should have repaired: assistant + synthetic_tool + user = 3 messages
	require.Len(t, req.Messages, 3)
	assert.Equal(t, "assistant", req.Messages[0].Role)
	assert.Equal(t, "tool", req.Messages[1].Role)
	assert.Equal(t, "call_1", req.Messages[1].ToolCallID)
	assert.Equal(t, "user", req.Messages[2].Role)
}
