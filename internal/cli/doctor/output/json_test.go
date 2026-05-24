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

func TestJSONRenderer_Render_SanitizesDisplayText(t *testing.T) {
	t.Parallel()

	renderer := &JSONRenderer{}
	out, err := renderer.Render(checks.Summary{
		Results: []checks.Result{{
			Name:      "Multi-\x1b[31mAgent\n",
			Status:    checks.StatusWarn,
			Message:   "Recent\x1b[31m failures\nfound",
			Details:   "Inspect\x1b[31m traces\nnow",
			FixAction: "Run\x1b[31m doctor --fix\n",
			TraceFailures: []checks.TraceFailure{{
				TraceID:    "trace-\x1b[31m1\n",
				Outcome:    "tool_\x1b[31merror\n",
				ErrorCode:  "E\x1b[31m003\n",
				CauseClass: "tool_\x1b[31mnot_found\n",
				Summary:    "[E003]\x1b[31m tool_not_found\n",
			}},
		}},
		Warnings: 1,
	})
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	results := decoded["results"].([]any)
	result := results[0].(map[string]any)
	assert.Equal(t, "Multi-Agent", result["name"])
	assert.Equal(t, "Recent failures found", result["message"])
	assert.Equal(t, "Inspect traces now", result["details"])
	assert.Equal(t, "Run doctor --fix", result["fixAction"])

	failures := result["traceFailures"].([]any)
	failure := failures[0].(map[string]any)
	assert.Equal(t, "trace-1", failure["traceId"])
	assert.Equal(t, "tool_error", failure["outcome"])
	assert.Equal(t, "E003", failure["errorCode"])
	assert.Equal(t, "tool_not_found", failure["causeClass"])
	assert.Equal(t, "[E003] tool_not_found", failure["summary"])
}

func TestJSONRenderer_Render_ProducesDecodablePrettyJSON(t *testing.T) {
	t.Parallel()

	renderer := &JSONRenderer{}
	out, err := renderer.Render(checks.Summary{
		Results: []checks.Result{{
			Name:    "Config",
			Status:  checks.StatusPass,
			Message: "ok",
		}},
		Passed: 1,
	})
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, out, "\n  \"results\":")
}
