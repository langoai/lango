package agentmemory

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	tools := BuildTools(nil)

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name:    "save requires key",
			tool:    "memory_agent_save",
			params:  map[string]interface{}{"content": "remember this"},
			wantErr: "missing key parameter",
		},
		{
			name:    "save requires content",
			tool:    "memory_agent_save",
			params:  map[string]interface{}{"key": "fact-1"},
			wantErr: "missing content parameter",
		},
		{
			name:    "recall requires query",
			tool:    "memory_agent_recall",
			params:  map[string]interface{}{},
			wantErr: "missing query parameter",
		},
		{
			name:    "forget requires key",
			tool:    "memory_agent_forget",
			params:  map[string]interface{}{},
			wantErr: "missing key parameter",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := findAgentMemoryTool(t, tools, tc.tool).Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func findAgentMemoryTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
