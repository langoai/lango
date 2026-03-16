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
