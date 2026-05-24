package graph

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
			name:    "traverse requires start node",
			tool:    "graph_traverse",
			params:  map[string]interface{}{},
			wantErr: "missing start_node parameter",
		},
		{
			name:    "query requires subject or object",
			tool:    "graph_query",
			params:  map[string]interface{}{},
			wantErr: "either subject or object is required",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := findGraphTool(t, tools, tc.tool).Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func findGraphTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
