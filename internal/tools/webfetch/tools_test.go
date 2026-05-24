package webfetch

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTools_RequireURLParameter(t *testing.T) {
	t.Parallel()

	result, err := findWebfetchTool(t, BuildTools(), "web_fetch").Handler(
		context.Background(),
		map[string]interface{}{},
	)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "missing url parameter")
}

func findWebfetchTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
