package output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/doctor/checks"
)

func TestJSONRenderer_Render_IncludesStructuredTraceMetadata(t *testing.T) {
	t.Parallel()

	leakCount := 2
	renderer := &JSONRenderer{}
	out, err := renderer.Render(checks.Summary{
		Results: []checks.Result{{
			Name:    "Multi-Agent",
			Status:  checks.StatusWarn,
			Message: "Multi-agent mode enabled with recent failures",
			TraceFailures: []checks.TraceFailure{{
				TraceID:    "trace-1",
				Outcome:    "tool_error",
				ErrorCode:  "E003",
				CauseClass: "tool_not_found",
				Summary:    "[E003] tool_not_found",
			}},
			IsolationLeakCount: &leakCount,
		}},
		Warnings: 1,
	})
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))

	results, ok := decoded["results"].([]any)
	require.True(t, ok)
	require.Len(t, results, 1)

	result, ok := results[0].(map[string]any)
	require.True(t, ok)
	failures, ok := result["traceFailures"].([]any)
	require.True(t, ok)
	require.Len(t, failures, 1)

	failure, ok := failures[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "trace-1", failure["traceId"])
	assert.Equal(t, "tool_error", failure["outcome"])
	assert.Equal(t, "E003", failure["errorCode"])
	assert.Equal(t, "tool_not_found", failure["causeClass"])
	assert.Equal(t, "[E003] tool_not_found", failure["summary"])
	assert.EqualValues(t, 2, result["isolationLeakCount"])
}
