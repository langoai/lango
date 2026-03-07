package skill

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/agent"
)

func newTestRegistry(t *testing.T) *Registry {
	dir := filepath.Join(t.TempDir(), "skills")
	logger := zap.NewNop().Sugar()
	store := NewFileSkillStore(dir, logger)
	baseTool := &agent.Tool{Name: "test_tool", Description: "a test tool"}
	return NewRegistry(store, []*agent.Tool{baseTool}, logger)
}

func TestRegistry_CreateSkill_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    string
		entry   SkillEntry
		wantErr string
	}{
		{
			give:    "empty name",
			entry:   SkillEntry{Name: "", Type: "composite", Definition: map[string]interface{}{"steps": []interface{}{}}},
			wantErr: "skill name is required",
		},
		{
			give:    "invalid type",
			entry:   SkillEntry{Name: "foo", Type: "unknown", Definition: map[string]interface{}{"steps": []interface{}{}}},
			wantErr: "skill type must be composite, script, template, or instruction",
		},
		{
			give:    "empty definition",
			entry:   SkillEntry{Name: "foo", Type: "composite", Definition: map[string]interface{}{}},
			wantErr: "skill definition is required",
		},
		{
			give: "dangerous script",
			entry: SkillEntry{
				Name: "danger",
				Type: "script",
				Definition: map[string]interface{}{
					"script": "rm -rf /",
				},
			},
			wantErr: "dangerous pattern",
		},
	}

	registry := newTestRegistry(t)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			err := registry.CreateSkill(ctx, tt.entry)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestRegistry_LoadSkills_AllTools(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)
	ctx := context.Background()

	// Before loading any skills, AllTools should return only the base tool.
	toolsBefore := registry.AllTools()
	require.Len(t, toolsBefore, 1)
	assert.Equal(t, "test_tool", toolsBefore[0].Name)

	// Create and activate a skill.
	require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
		Name:        "my_skill",
		Description: "does stuff",
		Type:        "template",
		Definition:  map[string]interface{}{"template": "Hello {{.Name}}"},
	}))

	require.NoError(t, registry.ActivateSkill(ctx, "my_skill"))

	// After activation (which calls LoadSkills internally), AllTools should include both.
	toolsAfter := registry.AllTools()
	require.Len(t, toolsAfter, 2)

	names := make(map[string]bool, len(toolsAfter))
	for _, tool := range toolsAfter {
		names[tool.Name] = true
	}
	assert.True(t, names["test_tool"], "AllTools missing base tool 'test_tool'")
	assert.True(t, names["skill_my_skill"], "AllTools missing loaded skill 'skill_my_skill'")
}

func TestRegistry_LoadedSkills(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)
	ctx := context.Background()

	// Before loading any skills, LoadedSkills should return empty (no base tools).
	loaded := registry.LoadedSkills()
	require.Empty(t, loaded)

	// Create and activate a skill.
	require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
		Name:        "loaded_skill",
		Description: "test loaded",
		Type:        "template",
		Definition:  map[string]interface{}{"template": "Hi"},
	}))

	require.NoError(t, registry.ActivateSkill(ctx, "loaded_skill"))

	// After activation, LoadedSkills should return only the dynamic skill.
	loaded = registry.LoadedSkills()
	require.Len(t, loaded, 1)
	assert.Equal(t, "skill_loaded_skill", loaded[0].Name)

	// AllTools should still include both base and loaded.
	all := registry.AllTools()
	require.Len(t, all, 2)
}

func TestRegistry_ActivateSkill(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)
	ctx := context.Background()

	// Create a skill first.
	require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
		Name:        "activate_me",
		Description: "a skill to activate",
		Type:        "composite",
		Definition: map[string]interface{}{
			"steps": []interface{}{
				map[string]interface{}{"tool": "read", "params": map[string]interface{}{"path": "/tmp"}},
			},
		},
	}))

	// Before activation, GetSkillTool should return false.
	_, found := registry.GetSkillTool("activate_me")
	assert.False(t, found)

	// Activate the skill.
	require.NoError(t, registry.ActivateSkill(ctx, "activate_me"))

	// After activation, GetSkillTool should return the tool.
	tool, found := registry.GetSkillTool("activate_me")
	require.True(t, found)
	assert.Equal(t, "skill_activate_me", tool.Name)
}

func TestRegistry_GetSkillTool(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)
	ctx := context.Background()

	t.Run("skill_ prefix naming", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
			Name:        "prefixed",
			Description: "test prefix",
			Type:        "template",
			Definition:  map[string]interface{}{"template": "test"},
		}))

		require.NoError(t, registry.ActivateSkill(ctx, "prefixed"))

		tool, found := registry.GetSkillTool("prefixed")
		require.True(t, found)
		assert.True(t, strings.HasPrefix(tool.Name, "skill_"))
		assert.Equal(t, "skill_prefixed", tool.Name)
	})

	t.Run("non-existent skill returns false", func(t *testing.T) {
		t.Parallel()

		_, found := registry.GetSkillTool("does_not_exist")
		assert.False(t, found)
	})
}

func TestRegistry_InstructionSkillAsTool(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)
	ctx := context.Background()

	// Create an instruction skill.
	require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
		Name:        "obsidian-ref",
		Description: "Obsidian Markdown reference guide",
		Type:        "instruction",
		Definition:  map[string]interface{}{"content": "# Obsidian\n\nUse wikilinks."},
		Source:      "https://github.com/owner/repo",
	}))

	require.NoError(t, registry.ActivateSkill(ctx, "obsidian-ref"))

	// Verify tool is registered.
	tool, found := registry.GetSkillTool("obsidian-ref")
	require.True(t, found)
	assert.Equal(t, "skill_obsidian-ref", tool.Name)
}

func TestRegistry_InstructionTool_ReturnsContent(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)
	ctx := context.Background()

	require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
		Name:        "my-guide",
		Description: "My guide",
		Type:        "instruction",
		Definition:  map[string]interface{}{"content": "Guide content here."},
		Source:      "https://example.com/guide",
	}))

	require.NoError(t, registry.ActivateSkill(ctx, "my-guide"))

	tool, found := registry.GetSkillTool("my-guide")
	require.True(t, found)

	// Call the handler.
	result, err := tool.Handler(ctx, map[string]interface{}{})
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "result type = %T, want map[string]interface{}", result)

	assert.Equal(t, "Guide content here.", resultMap["content"])
	assert.Equal(t, "https://example.com/guide", resultMap["source"])
	assert.Equal(t, "instruction", resultMap["type"])
}

func TestRegistry_InstructionTool_Description(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)
	ctx := context.Background()

	t.Run("custom description preserved", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
			Name:        "custom-desc",
			Description: "Use this when working with Obsidian Markdown syntax",
			Type:        "instruction",
			Definition:  map[string]interface{}{"content": "content"},
		}))
		require.NoError(t, registry.ActivateSkill(ctx, "custom-desc"))

		tool, _ := registry.GetSkillTool("custom-desc")
		assert.Equal(t, "Use this when working with Obsidian Markdown syntax", tool.Description)
	})

	t.Run("empty description gets default", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
			Name:       "no-desc",
			Type:       "instruction",
			Definition: map[string]interface{}{"content": "content"},
		}))
		require.NoError(t, registry.ActivateSkill(ctx, "no-desc"))

		tool, _ := registry.GetSkillTool("no-desc")
		assert.Equal(t, "Reference guide for no-desc", tool.Description)
	})
}

func TestRegistry_ListActiveSkills(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)
	ctx := context.Background()

	// Create and activate a skill.
	require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
		Name:        "listable",
		Description: "a listable skill",
		Type:        "script",
		Status:      "active",
		Definition:  map[string]interface{}{"script": "echo hi"},
	}))

	require.NoError(t, registry.ActivateSkill(ctx, "listable"))

	skills, err := registry.ListActiveSkills(ctx)
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "listable", skills[0].Name)
}
