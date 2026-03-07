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

func TestFormatContent_MultipleTextParts(t *testing.T) {
	t.Parallel()

	parts := []sdkmcp.Content{
		&sdkmcp.TextContent{Text: "line1"},
		&sdkmcp.TextContent{Text: "line2"},
	}
	result := formatContent(parts, 0)
	assert.Equal(t, "line1\nline2", result)
}

func TestFormatContent_ImageContent(t *testing.T) {
	t.Parallel()

	img := &sdkmcp.ImageContent{
		MIMEType: "image/png",
		Data:     []byte("iVBORw0KGgo="),
	}
	result := formatContent([]sdkmcp.Content{img}, 0)
	assert.Contains(t, result, "[Image: image/png")
}

func TestFormatContent_AudioContent(t *testing.T) {
	t.Parallel()

	audio := &sdkmcp.AudioContent{
		MIMEType: "audio/mp3",
		Data:     []byte("AAAA"),
	}
	result := formatContent([]sdkmcp.Content{audio}, 0)
	assert.Contains(t, result, "[Audio: audio/mp3]")
}

func TestFormatContent_ZeroMaxTokens(t *testing.T) {
	t.Parallel()

	text := &sdkmcp.TextContent{Text: "Hello World, this is a long text"}
	result := formatContent([]sdkmcp.Content{text}, 0)
	assert.Equal(t, "Hello World, this is a long text", result)
}

func TestExtractText_SingleText(t *testing.T) {
	t.Parallel()

	content := []sdkmcp.Content{
		&sdkmcp.TextContent{Text: "error message"},
	}
	assert.Equal(t, "error message", extractText(content))
}

func TestExtractText_MultipleText(t *testing.T) {
	t.Parallel()

	content := []sdkmcp.Content{
		&sdkmcp.TextContent{Text: "line 1"},
		&sdkmcp.TextContent{Text: "line 2"},
	}
	assert.Equal(t, "line 1\nline 2", extractText(content))
}

func TestExtractText_Empty(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "unknown error", extractText(nil))
	assert.Equal(t, "unknown error", extractText([]sdkmcp.Content{}))
}

func TestExtractText_NonTextContent(t *testing.T) {
	t.Parallel()

	content := []sdkmcp.Content{
		&sdkmcp.ImageContent{MIMEType: "image/png", Data: []byte("data")},
	}
	assert.Equal(t, "unknown error", extractText(content))
}

func TestBuildParams_EnumField(t *testing.T) {
	t.Parallel()

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"color": map[string]any{
				"type":        "string",
				"description": "Favorite color",
				"enum":        []any{"red", "green", "blue"},
			},
		},
	}

	params := buildParams(schema)
	assert.Len(t, params, 1)
	colorDef := params["color"].(map[string]interface{})
	assert.Equal(t, "string", colorDef["type"])
	assert.Equal(t, "Favorite color", colorDef["description"])
	assert.NotNil(t, colorDef["enum"])
}

func TestBuildParams_PropertyWithoutType(t *testing.T) {
	t.Parallel()

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"field": map[string]any{
				"description": "A field without explicit type",
			},
		},
	}

	params := buildParams(schema)
	assert.Len(t, params, 1)
	fieldDef := params["field"].(map[string]interface{})
	assert.Equal(t, "string", fieldDef["type"])
}

func TestBuildParams_NonMapProperty(t *testing.T) {
	t.Parallel()

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"weird": "just a string",
		},
	}

	params := buildParams(schema)
	assert.Len(t, params, 1)
	weirdDef := params["weird"].(map[string]interface{})
	assert.Equal(t, "string", weirdDef["type"])
}

func TestBuildParams_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	// Test a struct type that requires JSON round-trip
	type schemaStruct struct {
		Type       string         `json:"type"`
		Properties map[string]any `json:"properties"`
	}
	schema := schemaStruct{
		Type: "object",
		Properties: map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "The name",
			},
		},
	}

	params := buildParams(schema)
	assert.Len(t, params, 1)
	assert.Contains(t, params, "name")
}
