package websearch

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTools_RequireQueryParameter(t *testing.T) {
	t.Parallel()

	result, err := findWebsearchTool(t, BuildTools(), "web_search").Handler(
		context.Background(),
		map[string]interface{}{},
	)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "missing query parameter")
}

func findWebsearchTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
