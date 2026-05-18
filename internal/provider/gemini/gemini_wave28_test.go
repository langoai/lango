package gemini

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"

	"github.com/langoai/lango/internal/provider"
)

func TestWave28NewProviderPreservesIdentityAndModel(t *testing.T) {
	t.Parallel()

	p, err := NewProvider(context.Background(), "gemini-primary", "test-api-key", "gemini-2.5-flash")

	require.NoError(t, err)
	require.NotNil(t, p)
	require.NotNil(t, p.client)
	assert.Equal(t, "gemini-primary", p.ID())
	assert.Equal(t, "gemini-2.5-flash", p.model)
}

func TestWave28GenerateRejectsMismatchedModelBeforeNetwork(t *testing.T) {
	t.Parallel()

	p := &GeminiProvider{
		id:    "gemini-local",
		model: "gpt-5.3-codex",
	}

	stream, err := p.Generate(context.Background(), provider.GenerateParams{
		Messages: []provider.Message{{Role: "user", Content: "hello"}},
	})

	require.Error(t, err)
	assert.Nil(t, stream)
	assert.ErrorIs(t, err, provider.ErrModelProviderMismatch)
	assert.Contains(t, err.Error(), "gemini provider")
}

func TestWave28GenerateRejectsUnmarshalableToolSchemaBeforeNetwork(t *testing.T) {
	t.Parallel()

	p := &GeminiProvider{
		id:    "gemini-local",
		model: "gemini-2.5-flash",
	}

	stream, err := p.Generate(context.Background(), provider.GenerateParams{
		Messages: []provider.Message{{Role: "user", Content: "run search"}},
		Tools: []provider.Tool{
			{
				Name:        "search",
				Description: "Search local test data.",
				Parameters: map[string]interface{}{
					"type":       string(genai.TypeObject),
					"properties": func() {},
				},
			},
		},
	})

	require.Error(t, err)
	assert.Nil(t, stream)
	assert.Contains(t, err.Error(), "convert tool schema")
}

func TestWave28BuildGenerateContentRequestAssemblesValidRequestWithoutNetwork(t *testing.T) {
	t.Parallel()

	p := &GeminiProvider{
		id:    "gemini-local",
		model: "gemini-2.5-flash",
	}

	model, contents, conf, err := p.buildGenerateContentRequest(provider.GenerateParams{
		Model:       "gemini",
		Temperature: 0.25,
		MaxTokens:   321,
		Messages: []provider.Message{
			{Role: "system", Content: "Follow Wave 28 policy."},
			{Role: "user", Content: "Search for coverage."},
			{
				Role:    "assistant",
				Content: "Calling a tool.",
				ToolCalls: []provider.ToolCall{
					{ID: "call-1", Name: "search", Arguments: `{"query":"coverage"}`},
				},
			},
			{
				Role:    "tool",
				Content: `{"matches":2}`,
				Metadata: map[string]interface{}{
					"tool_call_id":   "call-1",
					"tool_call_name": "search",
				},
			},
		},
		Tools: []provider.Tool{
			{
				Name:        "search",
				Description: "Search local test data.",
				Parameters: map[string]interface{}{
					"type":     string(genai.TypeObject),
					"required": []interface{}{"query"},
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        string(genai.TypeString),
							"description": "Search query.",
						},
					},
				},
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "gemini-3-flash-preview", model)
	require.NotNil(t, conf)
	require.NotNil(t, conf.Temperature)
	assert.Equal(t, float32(0.25), *conf.Temperature)
	assert.Equal(t, int32(321), conf.MaxOutputTokens)

	require.NotNil(t, conf.SystemInstruction)
	require.Len(t, conf.SystemInstruction.Parts, 1)
	assert.Equal(t, "Follow Wave 28 policy.", conf.SystemInstruction.Parts[0].Text)

	require.Len(t, conf.Tools, 1)
	require.Len(t, conf.Tools[0].FunctionDeclarations, 1)
	decl := conf.Tools[0].FunctionDeclarations[0]
	assert.Equal(t, "search", decl.Name)
	assert.Equal(t, "Search local test data.", decl.Description)
	require.NotNil(t, decl.Parameters)
	assert.Equal(t, genai.TypeObject, decl.Parameters.Type)
	assert.ElementsMatch(t, []string{"query"}, decl.Parameters.Required)
	require.Contains(t, decl.Parameters.Properties, "query")
	assert.Equal(t, genai.TypeString, decl.Parameters.Properties["query"].Type)

	require.Len(t, contents, 3)
	assert.Equal(t, "user", contents[0].Role)
	assert.Equal(t, "Search for coverage.", contents[0].Parts[0].Text)
	assert.Equal(t, "model", contents[1].Role)
	assert.Equal(t, "Calling a tool.", contents[1].Parts[0].Text)
	require.NotNil(t, contents[1].Parts[1].FunctionCall)
	assert.Equal(t, "call-1", contents[1].Parts[1].FunctionCall.ID)
	assert.Equal(t, "search", contents[1].Parts[1].FunctionCall.Name)
	assert.Equal(t, map[string]interface{}{"query": "coverage"}, contents[1].Parts[1].FunctionCall.Args)
	assert.Equal(t, "user", contents[2].Role)
	require.NotNil(t, contents[2].Parts[0].FunctionResponse)
	assert.Equal(t, "call-1", contents[2].Parts[0].FunctionResponse.ID)
	assert.Equal(t, "search", contents[2].Parts[0].FunctionResponse.Name)
	assert.Equal(t, map[string]interface{}{"matches": float64(2)}, contents[2].Parts[0].FunctionResponse.Response)
}

func TestWave28ListModelsReturnsCanceledContextErrorBeforeNetwork(t *testing.T) {
	t.Parallel()

	p, err := NewProvider(context.Background(), "gemini-primary", "test-api-key", "gemini-2.5-flash")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	models, err := p.ListModels(ctx)

	require.Error(t, err)
	assert.Nil(t, models)
	assert.Contains(t, err.Error(), "list gemini models")
	assert.True(t,
		errors.Is(err, context.Canceled) || strings.Contains(err.Error(), context.Canceled.Error()),
		"expected context cancellation in error chain, got %v", err,
	)
}

func TestWave28ConvertSchemaMapsNestedJSONSchema(t *testing.T) {
	t.Parallel()

	schema, err := convertSchema(map[string]interface{}{
		"type":        string(genai.TypeObject),
		"description": "Search parameters.",
		"required":    []interface{}{"query", "limit"},
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        string(genai.TypeString),
				"description": "Search query.",
			},
			"limit": map[string]interface{}{
				"type":    string(genai.TypeInteger),
				"minimum": 1,
				"maximum": 10,
			},
			"filters": map[string]interface{}{
				"type": string(genai.TypeArray),
				"items": map[string]interface{}{
					"type": string(genai.TypeString),
				},
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, schema)
	assert.Equal(t, genai.TypeObject, schema.Type)
	assert.Equal(t, "Search parameters.", schema.Description)
	assert.ElementsMatch(t, []string{"query", "limit"}, schema.Required)

	require.Contains(t, schema.Properties, "query")
	assert.Equal(t, genai.TypeString, schema.Properties["query"].Type)
	assert.Equal(t, "Search query.", schema.Properties["query"].Description)

	require.Contains(t, schema.Properties, "limit")
	assert.Equal(t, genai.TypeInteger, schema.Properties["limit"].Type)
	require.NotNil(t, schema.Properties["limit"].Minimum)
	require.NotNil(t, schema.Properties["limit"].Maximum)
	assert.Equal(t, float64(1), *schema.Properties["limit"].Minimum)
	assert.Equal(t, float64(10), *schema.Properties["limit"].Maximum)

	require.Contains(t, schema.Properties, "filters")
	require.NotNil(t, schema.Properties["filters"].Items)
	assert.Equal(t, genai.TypeArray, schema.Properties["filters"].Type)
	assert.Equal(t, genai.TypeString, schema.Properties["filters"].Items.Type)
}

func TestWave28ConvertSchemaReportsMarshalErrors(t *testing.T) {
	t.Parallel()

	schema, err := convertSchema(map[string]interface{}{
		"type": string(genai.TypeObject),
		"bad":  make(chan struct{}),
	})

	require.Error(t, err)
	assert.Nil(t, schema)
	assert.Contains(t, err.Error(), "unsupported type")
}

func TestWave28HelperMatchingCoversIDAndNameFallbacks(t *testing.T) {
	t.Parallel()

	contents := []*genai.Content{
		{
			Role: "model",
			Parts: []*genai.Part{
				{FunctionCall: &genai.FunctionCall{ID: "call-id-only", Args: map[string]interface{}{}}},
				{FunctionCall: &genai.FunctionCall{Name: "call-name-only", Args: map[string]interface{}{}}},
			},
		},
		{
			Role: "user",
			Parts: []*genai.Part{
				{
					FunctionResponse: &genai.FunctionResponse{
						ID:       "call-id-only",
						Response: map[string]interface{}{"ok": true},
					},
				},
				{
					FunctionResponse: &genai.FunctionResponse{
						Name:     "call-name-only",
						Response: map[string]interface{}{"ok": true},
					},
				},
				{
					FunctionResponse: &genai.FunctionResponse{
						ID:       "orphan",
						Name:     "missing",
						Response: map[string]interface{}{"ok": false},
					},
				},
			},
		},
	}

	filtered := dropOrphanedFunctionResponses(contents)

	require.Len(t, filtered, 2)
	require.Len(t, filtered[1].Parts, 2)
	assert.Equal(t, "call-id-only", filtered[1].Parts[0].FunctionResponse.ID)
	assert.Equal(t, "call-name-only", filtered[1].Parts[1].FunctionResponse.Name)
}
