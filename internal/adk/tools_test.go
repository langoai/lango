package adk

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

func TestAdaptTool_ParameterDef(t *testing.T) {
	t.Parallel()

	tool := &agent.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]interface{}{
			"command": agent.ParameterDef{
				Type:        "string",
				Description: "The command to run",
				Required:    true,
			},
			"timeout": agent.ParameterDef{
				Type:        "integer",
				Description: "Timeout in seconds",
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}

	adkTool, err := AdaptTool(tool)
	require.NoError(t, err)
	require.NotNil(t, adkTool)
}

func TestAdaptTool_MapParams(t *testing.T) {
	t.Parallel()

	tool := &agent.Tool{
		Name:        "map_tool",
		Description: "A tool with map params",
		Parameters: map[string]interface{}{
			"arg": map[string]interface{}{
				"type":        "string",
				"description": "An argument",
				"required":    true,
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "result", nil
		},
	}

	adkTool, err := AdaptTool(tool)
	require.NoError(t, err)
	require.NotNil(t, adkTool)
}

func TestAdaptTool_FallbackParams(t *testing.T) {
	t.Parallel()

	// Test with an unknown param type (not ParameterDef, not map)
	tool := &agent.Tool{
		Name:        "fallback_tool",
		Description: "A tool with fallback params",
		Parameters: map[string]interface{}{
			"arg": "just a string", // Neither ParameterDef nor map
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	}

	adkTool, err := AdaptTool(tool)
	require.NoError(t, err)
	require.NotNil(t, adkTool)
}

func TestAdaptTool_NoParams(t *testing.T) {
	t.Parallel()

	tool := &agent.Tool{
		Name:        "no_params_tool",
		Description: "A tool with no params",
		Parameters:  map[string]interface{}{},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "done", nil
		},
	}

	adkTool, err := AdaptTool(tool)
	require.NoError(t, err)
	require.NotNil(t, adkTool)
}

func TestAdaptTool_WithEnum(t *testing.T) {
	t.Parallel()

	tool := &agent.Tool{
		Name:        "enum_tool",
		Description: "A tool with enum param",
		Parameters: map[string]interface{}{
			"action": agent.ParameterDef{
				Type:        "string",
				Description: "Action to take",
				Required:    true,
				Enum:        []string{"start", "stop", "restart"},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return params["action"], nil
		},
	}

	adkTool, err := AdaptTool(tool)
	require.NoError(t, err)
	require.NotNil(t, adkTool)
}

func TestBuildInputSchema_AdditionalPropertiesFalse(t *testing.T) {
	t.Parallel()

	tool := &agent.Tool{
		Name: "test",
		Parameters: agent.Schema().
			Str("command", "The command").
			Required("command").
			Build(),
	}

	schema := buildInputSchema(tool)
	require.NotNil(t, schema.AdditionalProperties, "expected AdditionalProperties to be set")

	// The "false schema" pattern: {not: {}} serializes to JSON false.
	require.NotNil(t, schema.AdditionalProperties.Not, "expected Not field for false schema")
}

func TestBuildInputSchema_SchemaBuilder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     string
		tool     *agent.Tool
		wantKeys []string
		wantReq  []string
	}{
		{
			give: "single required string param",
			tool: &agent.Tool{
				Name: "exec",
				Parameters: agent.Schema().
					Str("command", "The shell command to execute").
					Required("command").
					Build(),
			},
			wantKeys: []string{"command"},
			wantReq:  []string{"command"},
		},
		{
			give: "multiple params with mixed required",
			tool: &agent.Tool{
				Name: "read_file",
				Parameters: agent.Schema().
					Str("path", "File path").
					Int("offset", "Start line").
					Int("limit", "Number of lines").
					Required("path").
					Build(),
			},
			wantKeys: []string{"limit", "offset", "path"},
			wantReq:  []string{"path"},
		},
		{
			give: "enum param",
			tool: &agent.Tool{
				Name: "browser",
				Parameters: agent.Schema().
					Enum("action", "Browser action", "click", "type", "navigate").
					Str("selector", "CSS selector").
					Required("action").
					Build(),
			},
			wantKeys: []string{"action", "selector"},
			wantReq:  []string{"action"},
		},
		{
			give: "no required fields",
			tool: &agent.Tool{
				Name:       "status",
				Parameters: agent.Schema().Bool("verbose", "Verbose output").Build(),
			},
			wantKeys: []string{"verbose"},
			wantReq:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			schema := buildInputSchema(tt.tool)
			require.Equal(t, "object", schema.Type)

			var gotKeys []string
			for k := range schema.Properties {
				gotKeys = append(gotKeys, k)
			}
			sort.Strings(gotKeys)
			assert.Equal(t, tt.wantKeys, gotKeys)

			if tt.wantReq != nil {
				assert.Equal(t, tt.wantReq, schema.Required)
			} else {
				assert.Empty(t, schema.Required)
			}
		})
	}
}

func TestBuildInputSchema_SchemaBuilder_PropertyTypes(t *testing.T) {
	t.Parallel()

	tool := &agent.Tool{
		Name: "multi_type",
		Parameters: agent.Schema().
			Str("name", "The name").
			Int("count", "The count").
			Bool("verbose", "Verbose mode").
			Enum("action", "The action", "start", "stop").
			Required("name", "action").
			Build(),
	}

	schema := buildInputSchema(tool)

	assert.Equal(t, "string", schema.Properties["name"].Type)
	assert.Equal(t, "The name", schema.Properties["name"].Description)

	assert.Equal(t, "integer", schema.Properties["count"].Type)
	assert.Equal(t, "The count", schema.Properties["count"].Description)

	assert.Equal(t, "boolean", schema.Properties["verbose"].Type)
	assert.Equal(t, "Verbose mode", schema.Properties["verbose"].Description)

	assert.Equal(t, "string", schema.Properties["action"].Type)
	assert.Equal(t, "The action", schema.Properties["action"].Description)
	assert.Equal(t, []interface{}{"start", "stop"}, schema.Properties["action"].Enum)

	assert.Equal(t, []string{"name", "action"}, schema.Required)
}

func TestBuildInputSchema_FlatParameterDef(t *testing.T) {
	t.Parallel()

	tool := &agent.Tool{
		Name: "flat_pd",
		Parameters: map[string]interface{}{
			"command": agent.ParameterDef{
				Type:        "string",
				Description: "The command",
				Required:    true,
			},
			"timeout": agent.ParameterDef{
				Type:        "integer",
				Description: "Timeout in seconds",
			},
		},
	}

	schema := buildInputSchema(tool)
	assert.Equal(t, "object", schema.Type)
	assert.Equal(t, "string", schema.Properties["command"].Type)
	assert.Equal(t, "integer", schema.Properties["timeout"].Type)
	assert.Contains(t, schema.Required, "command")
	assert.NotContains(t, schema.Required, "timeout")
}

func TestBuildInputSchema_FlatMap(t *testing.T) {
	t.Parallel()

	tool := &agent.Tool{
		Name: "flat_map",
		Parameters: map[string]interface{}{
			"arg": map[string]interface{}{
				"type":        "string",
				"description": "An argument",
				"required":    true,
			},
		},
	}

	schema := buildInputSchema(tool)
	assert.Equal(t, "object", schema.Type)
	assert.Equal(t, "string", schema.Properties["arg"].Type)
	assert.Equal(t, "An argument", schema.Properties["arg"].Description)
	assert.Contains(t, schema.Required, "arg")
}

func TestBuildInputSchema_ArrayWithItems(t *testing.T) {
	t.Parallel()

	tool := &agent.Tool{
		Name: "array_tool",
		Parameters: map[string]interface{}{
			"args": map[string]interface{}{
				"type":        "array",
				"description": "Command arguments",
				"items":       map[string]interface{}{"type": "string"},
				"required":    true,
			},
		},
	}

	schema := buildInputSchema(tool)
	assert.Equal(t, "object", schema.Type)

	argsProp := schema.Properties["args"]
	require.NotNil(t, argsProp)
	assert.Equal(t, "array", argsProp.Type)
	assert.Equal(t, "Command arguments", argsProp.Description)
	require.NotNil(t, argsProp.Items, "array schema must have Items")
	assert.Equal(t, "string", argsProp.Items.Type)
	assert.Contains(t, schema.Required, "args")
}

func TestBuildInputSchema_ArrayWithObjectItems(t *testing.T) {
	t.Parallel()

	tool := &agent.Tool{
		Name: "escrow_tool",
		Parameters: map[string]interface{}{
			"milestones": map[string]interface{}{
				"type":        "array",
				"description": "Escrow milestones",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Milestone name",
						},
						"amount": map[string]interface{}{
							"type":        "integer",
							"description": "Amount in cents",
						},
					},
					"required": []string{"name", "amount"},
				},
			},
		},
	}

	schema := buildInputSchema(tool)
	milestones := schema.Properties["milestones"]
	require.NotNil(t, milestones)
	assert.Equal(t, "array", milestones.Type)

	items := milestones.Items
	require.NotNil(t, items, "array schema must have Items")
	assert.Equal(t, "object", items.Type)
	require.NotNil(t, items.Properties)
	assert.Equal(t, "string", items.Properties["name"].Type)
	assert.Equal(t, "integer", items.Properties["amount"].Type)
	assert.Equal(t, []string{"name", "amount"}, items.Required)
}

func TestBuildInputSchema_ArrayInFullSchemaFormat(t *testing.T) {
	t.Parallel()

	tool := &agent.Tool{
		Name: "schema_builder_array",
		Parameters: agent.Schema().
			Array("tags", "string", "Tag list").
			Required("tags").
			Build(),
	}

	schema := buildInputSchema(tool)
	assert.Equal(t, "object", schema.Type)

	tagsProp := schema.Properties["tags"]
	require.NotNil(t, tagsProp)
	assert.Equal(t, "array", tagsProp.Type)
	assert.Equal(t, "Tag list", tagsProp.Description)
	require.NotNil(t, tagsProp.Items, "array schema must have Items")
	assert.Equal(t, "string", tagsProp.Items.Type)
	assert.Equal(t, []string{"tags"}, schema.Required)
}

func TestAdaptTool_SchemaBuilder(t *testing.T) {
	t.Parallel()

	tool := &agent.Tool{
		Name:        "exec",
		Description: "Execute a shell command",
		Parameters: agent.Schema().
			Str("command", "The shell command to execute").
			Required("command").
			Build(),
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return params["command"], nil
		},
	}

	adkTool, err := AdaptTool(tool)
	require.NoError(t, err)
	require.NotNil(t, adkTool)
}
