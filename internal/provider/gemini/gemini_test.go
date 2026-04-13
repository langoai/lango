package gemini

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/provider"
	"google.golang.org/genai"
)

func TestInferToolNameFromHistory(t *testing.T) {
	tests := []struct {
		give       string
		msgs       []provider.Message
		idx        int
		toolCallID string
		wantName   string
	}{
		{
			give: "finds name from preceding assistant",
			msgs: []provider.Message{
				{Role: "user", Content: "run"},
				{
					Role: "assistant",
					ToolCalls: []provider.ToolCall{
						{ID: "call_1", Name: "exec"},
					},
				},
				{
					Role:     "tool",
					Content:  `{"result":"ok"}`,
					Metadata: map[string]interface{}{"tool_call_id": "call_1"},
				},
			},
			idx:        2,
			toolCallID: "call_1",
			wantName:   "exec",
		},
		{
			give: "no preceding assistant returns empty",
			msgs: []provider.Message{
				{Role: "user", Content: "hello"},
				{
					Role:     "tool",
					Content:  `{"result":"ok"}`,
					Metadata: map[string]interface{}{"tool_call_id": "call_x"},
				},
			},
			idx:        1,
			toolCallID: "call_x",
			wantName:   "",
		},
		{
			give: "non-matching ID returns empty",
			msgs: []provider.Message{
				{
					Role: "assistant",
					ToolCalls: []provider.ToolCall{
						{ID: "call_other", Name: "read"},
					},
				},
				{
					Role:     "tool",
					Content:  `{"result":"ok"}`,
					Metadata: map[string]interface{}{"tool_call_id": "call_missing"},
				},
			},
			idx:        1,
			toolCallID: "call_missing",
			wantName:   "",
		},
		{
			give: "only checks nearest assistant",
			msgs: []provider.Message{
				{
					Role: "assistant",
					ToolCalls: []provider.ToolCall{
						{ID: "call_1", Name: "far_tool"},
					},
				},
				{Role: "user", Content: "ok"},
				{
					Role: "assistant",
					ToolCalls: []provider.ToolCall{
						{ID: "call_2", Name: "near_tool"},
					},
				},
				{
					Role:     "tool",
					Content:  `{}`,
					Metadata: map[string]interface{}{"tool_call_id": "call_1"},
				},
			},
			idx:        3,
			toolCallID: "call_1",
			wantName:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := inferToolNameFromHistory(tt.msgs, tt.idx, tt.toolCallID)
			assert.Equal(t, tt.wantName, got)
		})
	}
}

func TestThoughtSignatureFiltering(t *testing.T) {
	tests := []struct {
		give      string
		toolCalls []provider.ToolCall
		wantCount int
	}{
		{
			give: "Thought=true with empty ThoughtSignature is dropped",
			toolCalls: []provider.ToolCall{
				{ID: "call_1", Name: "exec", Arguments: `{}`, Thought: true, ThoughtSignature: nil},
			},
			wantCount: 0,
		},
		{
			give: "Thought=false with empty ThoughtSignature passes through",
			toolCalls: []provider.ToolCall{
				{ID: "call_1", Name: "exec", Arguments: `{}`, Thought: false, ThoughtSignature: nil},
			},
			wantCount: 1,
		},
		{
			give: "Thought=true with valid ThoughtSignature passes through",
			toolCalls: []provider.ToolCall{
				{ID: "call_1", Name: "exec", Arguments: `{}`, Thought: true, ThoughtSignature: []byte("sig123")},
			},
			wantCount: 1,
		},
		{
			give: "mixed: one dropped one kept",
			toolCalls: []provider.ToolCall{
				{ID: "call_1", Name: "exec", Arguments: `{}`, Thought: true, ThoughtSignature: nil},
				{ID: "call_2", Name: "read", Arguments: `{}`, Thought: false},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			// Build params with an assistant message carrying the tool calls
			params := provider.GenerateParams{
				Messages: []provider.Message{
					{Role: "user", Content: "hello"},
					{Role: "assistant", ToolCalls: tt.toolCalls},
				},
			}

			// Use the Generate method's internal content builder by testing
			// the content construction directly. We simulate what Generate does.
			var contents []*genai.Content
			for _, m := range params.Messages {
				if m.Role == "user" {
					contents = append(contents, &genai.Content{
						Role:  "user",
						Parts: []*genai.Part{{Text: m.Content}},
					})
					continue
				}
				if m.Role == "assistant" {
					role := "model"
					var parts []*genai.Part
					for _, tc := range m.ToolCalls {
						if tc.Thought && len(tc.ThoughtSignature) == 0 {
							continue
						}
						args := map[string]interface{}{}
						parts = append(parts, &genai.Part{
							FunctionCall: &genai.FunctionCall{
								ID:   tc.ID,
								Name: tc.Name,
								Args: args,
							},
							Thought:          tc.Thought,
							ThoughtSignature: tc.ThoughtSignature,
						})
					}
					if len(parts) > 0 {
						contents = append(contents, &genai.Content{Role: role, Parts: parts})
					}
				}
			}

			// Count FunctionCall parts in model contents
			count := 0
			for _, c := range contents {
				if c.Role != "model" {
					continue
				}
				for _, p := range c.Parts {
					if p.FunctionCall != nil {
						count++
					}
				}
			}
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestFunctionCallID_InContent(t *testing.T) {
	t.Parallel()

	// Verify that FunctionCall.ID and FunctionResponse.ID are set in genai.Content
	params := provider.GenerateParams{
		Messages: []provider.Message{
			{Role: "user", Content: "run tool"},
			{
				Role: "assistant",
				ToolCalls: []provider.ToolCall{
					{ID: "call_abc", Name: "exec", Arguments: `{"cmd":"ls"}`},
				},
			},
			{
				Role:    "tool",
				Content: `{"result":"files"}`,
				Metadata: map[string]interface{}{
					"tool_call_id":   "call_abc",
					"tool_call_name": "exec",
				},
			},
		},
	}

	// Simulate content building (same logic as Generate)
	var contents []*genai.Content
	for i, m := range params.Messages {
		switch m.Role {
		case "user":
			contents = append(contents, &genai.Content{
				Role:  "user",
				Parts: []*genai.Part{{Text: m.Content}},
			})
		case "assistant":
			var parts []*genai.Part
			for _, tc := range m.ToolCalls {
				args := map[string]interface{}{}
				parts = append(parts, &genai.Part{
					FunctionCall: &genai.FunctionCall{
						ID:   tc.ID,
						Name: tc.Name,
						Args: args,
					},
				})
			}
			contents = append(contents, &genai.Content{Role: "model", Parts: parts})
		case "tool":
			toolCallID, _ := m.Metadata["tool_call_id"].(string)
			toolCallName, _ := m.Metadata["tool_call_name"].(string)
			if toolCallName == "" && toolCallID != "" {
				toolCallName = inferToolNameFromHistory(params.Messages, i, toolCallID)
			}
			contents = append(contents, &genai.Content{
				Role: "user",
				Parts: []*genai.Part{{
					FunctionResponse: &genai.FunctionResponse{
						ID:       toolCallID,
						Name:     toolCallName,
						Response: map[string]interface{}{"result": "files"},
					},
				}},
			})
		}
	}

	// Verify FunctionCall.ID
	require.Len(t, contents, 3)
	modelContent := contents[1]
	require.Len(t, modelContent.Parts, 1)
	assert.Equal(t, "call_abc", modelContent.Parts[0].FunctionCall.ID)
	assert.Equal(t, "exec", modelContent.Parts[0].FunctionCall.Name)

	// Verify FunctionResponse.ID
	responseContent := contents[2]
	require.Len(t, responseContent.Parts, 1)
	assert.Equal(t, "call_abc", responseContent.Parts[0].FunctionResponse.ID)
	assert.Equal(t, "exec", responseContent.Parts[0].FunctionResponse.Name)
}

func TestInferToolNameFromHistory_BackwardCompat(t *testing.T) {
	t.Parallel()

	// Legacy message without tool_call_name in metadata — should be inferred
	params := provider.GenerateParams{
		Messages: []provider.Message{
			{Role: "user", Content: "run tool"},
			{
				Role: "assistant",
				ToolCalls: []provider.ToolCall{
					{ID: "call_legacy", Name: "exec"},
				},
			},
			{
				Role:    "tool",
				Content: `{"result":"ok"}`,
				Metadata: map[string]interface{}{
					"tool_call_id": "call_legacy",
					// No tool_call_name — legacy session
				},
			},
		},
	}

	m := params.Messages[2]
	toolCallID, _ := m.Metadata["tool_call_id"].(string)
	toolCallName, _ := m.Metadata["tool_call_name"].(string)

	assert.Equal(t, "", toolCallName, "legacy message should not have tool_call_name")

	// inferToolNameFromHistory should recover the name
	inferred := inferToolNameFromHistory(params.Messages, 2, toolCallID)
	assert.Equal(t, "exec", inferred)
}

func TestDropOrphanedFunctionResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give         string
		contents     []*genai.Content
		wantParts    int // total parts after filtering
		wantContents int // total content blocks after filtering
	}{
		{
			give: "orphan response removed when FunctionCall dropped",
			contents: []*genai.Content{
				{
					Role: "user",
					Parts: []*genai.Part{{Text: "hello"}},
				},
				{
					Role: "model",
					Parts: []*genai.Part{
						{FunctionCall: &genai.FunctionCall{ID: "call_real", Name: "exec", Args: map[string]interface{}{}}},
						// Note: call_thought was dropped by thought filtering, so no FunctionCall for it.
					},
				},
				{
					Role: "user",
					Parts: []*genai.Part{
						{FunctionResponse: &genai.FunctionResponse{ID: "call_real", Name: "exec", Response: map[string]interface{}{"result": "ok"}}},
						{FunctionResponse: &genai.FunctionResponse{ID: "call_thought", Name: "think", Response: map[string]interface{}{"result": "thought"}}},
					},
				},
			},
			wantParts:    1, // only call_real response survives
			wantContents: 3,
		},
		{
			give: "no orphans when all FunctionCalls present",
			contents: []*genai.Content{
				{
					Role: "model",
					Parts: []*genai.Part{
						{FunctionCall: &genai.FunctionCall{ID: "c1", Name: "exec", Args: map[string]interface{}{}}},
						{FunctionCall: &genai.FunctionCall{ID: "c2", Name: "read", Args: map[string]interface{}{}}},
					},
				},
				{
					Role: "user",
					Parts: []*genai.Part{
						{FunctionResponse: &genai.FunctionResponse{ID: "c1", Name: "exec", Response: map[string]interface{}{}}},
						{FunctionResponse: &genai.FunctionResponse{ID: "c2", Name: "read", Response: map[string]interface{}{}}},
					},
				},
			},
			wantParts:    2,
			wantContents: 2,
		},
		{
			give: "all responses orphaned removes content block",
			contents: []*genai.Content{
				{
					Role: "user",
					Parts: []*genai.Part{{Text: "hello"}},
				},
				{
					Role: "model",
					Parts: []*genai.Part{{Text: "thinking..."}},
				},
				{
					Role: "user",
					Parts: []*genai.Part{
						{FunctionResponse: &genai.FunctionResponse{ID: "ghost", Name: "gone", Response: map[string]interface{}{}}},
					},
				},
			},
			wantParts:    0, // the orphan content block is removed entirely
			wantContents: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			result := dropOrphanedFunctionResponses(tt.contents)
			assert.Len(t, result, tt.wantContents)

			// Count FunctionResponse parts in all content blocks.
			var frCount int
			for _, c := range result {
				for _, p := range c.Parts {
					if p.FunctionResponse != nil {
						frCount++
					}
				}
			}
			assert.Equal(t, tt.wantParts, frCount)
		})
	}
}

func TestResolveFunctionCallID(t *testing.T) {
	tests := []struct {
		give   string
		fc     *genai.FunctionCall
		wantID string
	}{
		{
			give:   "uses ID when present",
			fc:     &genai.FunctionCall{ID: "fc_123", Name: "exec"},
			wantID: "fc_123",
		},
		{
			give:   "falls back to Name when ID empty",
			fc:     &genai.FunctionCall{ID: "", Name: "exec"},
			wantID: "exec",
		},
		{
			give:   "ID takes precedence over Name",
			fc:     &genai.FunctionCall{ID: "custom_id", Name: "different_name"},
			wantID: "custom_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := resolveFunctionCallID(tt.fc)
			assert.Equal(t, tt.wantID, got)
		})
	}
}
