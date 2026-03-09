package skill

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCompositeSkill(t *testing.T) {
	t.Parallel()

	t.Run("basic fields and steps conversion", func(t *testing.T) {
		t.Parallel()

		steps := []SkillStep{
			{Tool: "read", Params: map[string]interface{}{"path": "/tmp"}},
			{Tool: "write", Params: map[string]interface{}{"path": "/out"}},
		}
		got := BuildCompositeSkill("my-skill", "does things", steps, nil)

		assert.Equal(t, "my-skill", got.Name)
		assert.Equal(t, "does things", got.Description)
		assert.Equal(t, SkillTypeComposite, got.Type)
		assert.True(t, got.RequiresApproval)

		stepDefs, ok := got.Definition["steps"].([]interface{})
		require.True(t, ok, "Definition[\"steps\"] is %T, want []interface{}", got.Definition["steps"])
		require.Len(t, stepDefs, 2)

		first, ok := stepDefs[0].(map[string]interface{})
		require.True(t, ok, "stepDefs[0] is %T, want map[string]interface{}", stepDefs[0])
		assert.Equal(t, "read", first["tool"])
	})

	t.Run("nil params leaves Parameters nil", func(t *testing.T) {
		t.Parallel()

		got := BuildCompositeSkill("s", "d", nil, nil)
		assert.Nil(t, got.Parameters)
	})

	t.Run("non-nil params sets Parameters", func(t *testing.T) {
		t.Parallel()

		params := map[string]interface{}{"key": "value"}
		got := BuildCompositeSkill("s", "d", nil, params)
		require.NotNil(t, got.Parameters)
		assert.Equal(t, "value", got.Parameters["key"])
	})
}

func TestBuildScriptSkill(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		giveScript string
		giveParams map[string]interface{}
	}{
		{
			give:       "with params",
			giveScript: "echo hello",
			giveParams: map[string]interface{}{"env": "prod"},
		},
		{
			give:       "without params",
			giveScript: "ls -la",
			giveParams: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got := BuildScriptSkill("run", "runs script", tt.giveScript, tt.giveParams)

			assert.Equal(t, SkillTypeScript, got.Type)
			assert.True(t, got.RequiresApproval)

			script, ok := got.Definition["script"].(string)
			require.True(t, ok, "Definition[\"script\"] is %T, want string", got.Definition["script"])
			assert.Equal(t, tt.giveScript, script)

			if tt.giveParams != nil {
				assert.NotNil(t, got.Parameters)
			} else {
				assert.Nil(t, got.Parameters)
			}
		})
	}
}

func TestBuildTemplateSkill(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give         string
		giveTemplate string
		giveParams   map[string]interface{}
	}{
		{
			give:         "with params",
			giveTemplate: "Hello {{.Name}}",
			giveParams:   map[string]interface{}{"name": "string"},
		},
		{
			give:         "without params",
			giveTemplate: "static template",
			giveParams:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got := BuildTemplateSkill("tmpl", "renders template", tt.giveTemplate, tt.giveParams)

			assert.Equal(t, SkillTypeTemplate, got.Type)
			assert.True(t, got.RequiresApproval)

			tmpl, ok := got.Definition["template"].(string)
			require.True(t, ok, "Definition[\"template\"] is %T, want string", got.Definition["template"])
			assert.Equal(t, tt.giveTemplate, tmpl)

			if tt.giveParams != nil {
				assert.NotNil(t, got.Parameters)
			} else {
				assert.Nil(t, got.Parameters)
			}
		})
	}
}
