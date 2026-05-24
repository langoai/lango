package tooloutput

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	store := NewOutputStore(0)
	tools := BuildTools(store)

	t.Run("requires ref", func(t *testing.T) {
		t.Parallel()

		got, err := findOutputTool(t, tools, "tool_output_get").Handler(context.Background(), map[string]interface{}{})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.EqualError(t, err, "missing ref parameter")
	})

	t.Run("grep requires pattern", func(t *testing.T) {
		t.Parallel()

		ref := store.Store("test-tool", "line one\nline two")
		got, err := findOutputTool(t, tools, "tool_output_get").Handler(context.Background(), map[string]interface{}{
			"ref":  ref,
			"mode": "grep",
		})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.EqualError(t, err, "pattern required for grep mode")
	})
}

func findOutputTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
