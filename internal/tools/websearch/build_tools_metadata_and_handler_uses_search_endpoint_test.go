package websearch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTools_MetadataAndHandlerUsesSearchEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "coverage query", r.URL.Query().Get("q"))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(sampleDDGHTML))
	}))
	defer srv.Close()

	origEndpoint := searchEndpoint
	t.Cleanup(func() { restoreEndpoint(origEndpoint) })
	overrideEndpoint(srv.URL + "/?q=")

	tools := BuildTools()
	require.Len(t, tools, 1)

	tool := tools[0]
	assert.Equal(t, "web_search", tool.Name)
	assert.Equal(t, agent.SafetyLevelSafe, tool.SafetyLevel)
	assert.Equal(t, "web", tool.Capability.Category)
	assert.Equal(t, agent.ActivityQuery, tool.Capability.Activity)
	assert.True(t, tool.Capability.ReadOnly)
	assert.True(t, tool.Capability.ConcurrencySafe)
	assert.Contains(t, tool.Capability.Aliases, "internet_search")
	assert.Contains(t, tool.Capability.SearchHints, "lookup")

	required, ok := tool.Parameters["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "query")

	result, err := tool.Handler(context.Background(), map[string]interface{}{
		"query": "coverage query",
		"limit": float64(2),
	})
	require.NoError(t, err)

	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "coverage query", payload["query"])
	assert.Equal(t, 2, payload["count"])

	results, ok := payload["results"].([]SearchResult)
	require.True(t, ok)
	require.Len(t, results, 2)
	assert.Equal(t, "Example Page One", results[0].Title)
	assert.Equal(t, "https://example.com/page1", results[0].URL)
}
