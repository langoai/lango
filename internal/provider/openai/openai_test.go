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
