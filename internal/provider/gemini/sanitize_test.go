package gemini

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

func TestSanitizeContents(t *testing.T) {
	tests := []struct {
		give string
		// input builds the content sequence for this test case.
		input func() []*genai.Content
		// wantRoles is the expected role sequence after sanitization.
		wantRoles []string
		// wantPartCounts is the expected number of parts per content entry.
		wantPartCounts []int
	}{
		{
			give: "empty input",
			input: func() []*genai.Content {
				return nil
			},
			wantRoles:      nil,
			wantPartCounts: nil,
		},
		{
			give: "already valid user-model sequence",
			input: func() []*genai.Content {
				return []*genai.Content{
					{Role: "user", Parts: []*genai.Part{{Text: "hello"}}},
					{Role: "model", Parts: []*genai.Part{{Text: "hi"}}},
				}
			},
			wantRoles:      []string{"user", "model"},
			wantPartCounts: []int{1, 1},
		},
		{
			give: "consecutive model turns are merged",
			input: func() []*genai.Content {
				return []*genai.Content{
					{Role: "user", Parts: []*genai.Part{{Text: "hello"}}},
					{Role: "model", Parts: []*genai.Part{{Text: "part1"}}},
					{Role: "model", Parts: []*genai.Part{{Text: "part2"}}},
				}
			},
			wantRoles:      []string{"user", "model"},
			wantPartCounts: []int{1, 2},
		},
		{
			give: "consecutive user turns are merged",
			input: func() []*genai.Content {
				return []*genai.Content{
					{Role: "user", Parts: []*genai.Part{{Text: "msg1"}}},
					{Role: "user", Parts: []*genai.Part{{Text: "msg2"}}},
					{Role: "model", Parts: []*genai.Part{{Text: "reply"}}},
				}
			},
			wantRoles:      []string{"user", "model"},
			wantPartCounts: []int{2, 1},
		},
		{
			give: "model turn at start gets user prepended",
			input: func() []*genai.Content {
				return []*genai.Content{
					{Role: "model", Parts: []*genai.Part{{Text: "I am a model"}}},
				}
			},
			wantRoles:      []string{"user", "model"},
			wantPartCounts: []int{1, 1},
		},
		{
			give: "orphaned FunctionResponse at start is removed",
			input: func() []*genai.Content {
				return []*genai.Content{
					{Role: "user", Parts: []*genai.Part{
						{FunctionResponse: &genai.FunctionResponse{Name: "tool_a", Response: map[string]any{"r": "ok"}}},
					}},
					{Role: "user", Parts: []*genai.Part{{Text: "hello"}}},
					{Role: "model", Parts: []*genai.Part{{Text: "hi"}}},
				}
			},
			// After orphan removal, the two user turns should be merged.
			wantRoles:      []string{"user", "model"},
			wantPartCounts: []int{1, 1},
		},
		{
			give: "FunctionCall without FunctionResponse gets synthetic response",
			input: func() []*genai.Content {
				return []*genai.Content{
					{Role: "user", Parts: []*genai.Part{{Text: "do something"}}},
					{Role: "model", Parts: []*genai.Part{
						{FunctionCall: &genai.FunctionCall{Name: "tool_a", Args: map[string]any{"x": 1}}},
					}},
					// No FunctionResponse follows — next is user.
					{Role: "user", Parts: []*genai.Part{{Text: "next"}}},
				}
			},
			// synthetic FunctionResponse inserted between model and user, then merged with user.
			wantRoles:      []string{"user", "model", "user"},
			wantPartCounts: []int{1, 1, 2}, // synthetic response (1 part) + "next" (1 part) merged
		},
		{
			give: "valid FunctionCall-FunctionResponse pair passes through",
			input: func() []*genai.Content {
				return []*genai.Content{
					{Role: "user", Parts: []*genai.Part{{Text: "call tool"}}},
					{Role: "model", Parts: []*genai.Part{
						{FunctionCall: &genai.FunctionCall{Name: "tool_a", Args: map[string]any{}}},
					}},
					{Role: "user", Parts: []*genai.Part{
						{FunctionResponse: &genai.FunctionResponse{Name: "tool_a", Response: map[string]any{"ok": true}}},
					}},
					{Role: "model", Parts: []*genai.Part{{Text: "done"}}},
				}
			},
			wantRoles:      []string{"user", "model", "user", "model"},
			wantPartCounts: []int{1, 1, 1, 1},
		},
		{
			give: "complex mixed sequence with multiple issues",
			input: func() []*genai.Content {
				return []*genai.Content{
					// Orphaned FunctionResponse at start
					{Role: "user", Parts: []*genai.Part{
						{FunctionResponse: &genai.FunctionResponse{Name: "stale", Response: map[string]any{}}},
					}},
					// Consecutive model turns
					{Role: "model", Parts: []*genai.Part{{Text: "thinking"}}},
					{Role: "model", Parts: []*genai.Part{
						{FunctionCall: &genai.FunctionCall{Name: "search", Args: map[string]any{"q": "test"}}},
					}},
					// FunctionResponse present
					{Role: "user", Parts: []*genai.Part{
						{FunctionResponse: &genai.FunctionResponse{Name: "search", Response: map[string]any{"result": "found"}}},
					}},
					{Role: "model", Parts: []*genai.Part{{Text: "result"}}},
				}
			},
			// orphan removed → starts with model → user prepended → models merged →
			// sequence: user(synthetic), model(thinking+FunctionCall), user(FunctionResponse), model(result)
			wantRoles:      []string{"user", "model", "user", "model"},
			wantPartCounts: []int{1, 2, 1, 1},
		},
		{
			give: "multiple FunctionCalls in single model turn get paired responses",
			input: func() []*genai.Content {
				return []*genai.Content{
					{Role: "user", Parts: []*genai.Part{{Text: "multi-tool"}}},
					{Role: "model", Parts: []*genai.Part{
						{FunctionCall: &genai.FunctionCall{Name: "tool_a", Args: map[string]any{}}},
						{FunctionCall: &genai.FunctionCall{Name: "tool_b", Args: map[string]any{}}},
					}},
					// No FunctionResponse follows.
					{Role: "model", Parts: []*genai.Part{{Text: "done"}}},
				}
			},
			// Consecutive models merged → model(2 FC + text "done") = 3 parts.
			// Synthetic user(2 FR) appended after model.
			wantRoles:      []string{"user", "model", "user"},
			wantPartCounts: []int{1, 3, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result := sanitizeContents(tt.input())

			if tt.wantRoles == nil {
				assert.Empty(t, result)
				return
			}

			require.Len(t, result, len(tt.wantRoles), "role count mismatch")

			for i, c := range result {
				assert.Equal(t, tt.wantRoles[i], c.Role, "role mismatch at index %d", i)
				assert.Len(t, c.Parts, tt.wantPartCounts[i], "part count mismatch at index %d (role=%s)", i, c.Role)
			}

			// Verify invariants on the sanitized output.
			assertNoConsecutiveSameRole(t, result)
			assertFunctionCallPairsValid(t, result)
		})
	}
}

// assertNoConsecutiveSameRole verifies that no two adjacent contents share the same role.
func assertNoConsecutiveSameRole(t *testing.T, contents []*genai.Content) {
	t.Helper()
	for i := 1; i < len(contents); i++ {
		assert.NotEqual(t, contents[i-1].Role, contents[i].Role,
			"consecutive same role at index %d-%d: %s", i-1, i, contents[i].Role)
	}
}

// assertFunctionCallPairsValid verifies that every model+FunctionCall is followed
// by a content with FunctionResponse.
func assertFunctionCallPairsValid(t *testing.T, contents []*genai.Content) {
	t.Helper()
	for i, c := range contents {
		if !hasFunctionCallParts(c) {
			continue
		}
		require.Less(t, i+1, len(contents),
			"FunctionCall at index %d has no following content", i)
		assert.True(t, hasFunctionResponseParts(contents[i+1]),
			"FunctionCall at index %d not followed by FunctionResponse (got role=%s)", i, contents[i+1].Role)
	}
}
