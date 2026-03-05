package mcp

import (
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/agent"
)

func TestBuildParams(t *testing.T) {
	tests := []struct {
		give     any
		wantLen  int
		wantKeys []string
	}{
		{give: nil, wantLen: 0},
		{give: "not a map", wantLen: 0},
		{
			give: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The name",
					},
					"count": map[string]any{
						"type":        "integer",
						"description": "A count",
					},
				},
				"required": []any{"name"},
			},
			wantLen:  2,
			wantKeys: []string{"name", "count"},
		},
		{
			give: map[string]any{
				"type":       "object",
				"properties": "not-a-map",
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		params := buildParams(tt.give)
		assert.Len(t, params, tt.wantLen)
		for _, key := range tt.wantKeys {
			assert.Contains(t, params, key)
		}
	}
}

func TestBuildParams_Required(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"a": map[string]any{"type": "string"},
			"b": map[string]any{"type": "integer"},
		},
		"required": []any{"a"},
	}

	params := buildParams(schema)
	aDef := params["a"].(map[string]interface{})
	bDef := params["b"].(map[string]interface{})

	assert.Equal(t, true, aDef["required"])
	_, hasRequired := bDef["required"]
	assert.False(t, hasRequired)
}

func TestParseSafetyLevel(t *testing.T) {
	tests := []struct {
		give string
		want agent.SafetyLevel
	}{
		{give: "safe", want: agent.SafetyLevelSafe},
		{give: "Safe", want: agent.SafetyLevelSafe},
		{give: "moderate", want: agent.SafetyLevelModerate},
		{give: "dangerous", want: agent.SafetyLevelDangerous},
		{give: "", want: agent.SafetyLevelDangerous},
		{give: "unknown", want: agent.SafetyLevelDangerous},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, parseSafetyLevel(tt.give))
		})
	}
}

func TestFormatContent_Empty(t *testing.T) {
	result := formatContent(nil, 0)
	assert.Empty(t, result)
}

func TestFormatContent_Truncation(t *testing.T) {
	// maxTokens=1 → maxChars=4, so any text longer than 4 chars is truncated
	longText := &sdkmcp.TextContent{Text: "Hello World, this is a long text"}
	result := formatContent([]sdkmcp.Content{longText}, 1)
	assert.Contains(t, result, "... [truncated]")
	assert.True(t, len(result) < len("Hello World, this is a long text")+20)
}

func TestFormatContent_NoTruncation(t *testing.T) {
	text := &sdkmcp.TextContent{Text: "short"}
	result := formatContent([]sdkmcp.Content{text}, 1000)
	assert.Equal(t, "short", result)
}
