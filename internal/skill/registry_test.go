package skill

import (
	"context"
	"os"
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
			wantErr: "skill type must be composite, script, template, instruction, or fork",
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

func TestRegistry_ForkSkillAsTool(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)
	ctx := context.Background()

	t.Run("fork skill creates tool with delegation handler", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
			Name:        "deploy-fork",
			Description: "Delegates deployment to deployer agent",
			Type:        SkillTypeFork,
			Agent:       "deployer",
			Definition: map[string]interface{}{
				"instruction": "Deploy to production",
			},
			AllowedTools: []string{"bash"},
		}))

		require.NoError(t, registry.ActivateSkill(ctx, "deploy-fork"))

		tool, found := registry.GetSkillTool("deploy-fork")
		require.True(t, found)
		assert.Equal(t, "skill_deploy-fork", tool.Name)
		assert.Equal(t, "Delegates deployment to deployer agent", tool.Description)

		// Call the handler and verify delegation text.
		result, err := tool.Handler(ctx, map[string]interface{}{})
		require.NoError(t, err)

		got, ok := result.(string)
		require.True(t, ok, "result is %T, want string", result)
		assert.Contains(t, got, "deployer")
		assert.Contains(t, got, "Deploy to production")
	})

	t.Run("fork skill default description mentions agent", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
			Name: "auto-desc-fork",
			Type: SkillTypeFork,
			Definition: map[string]interface{}{
				"instruction": "Do something",
			},
		}))

		require.NoError(t, registry.ActivateSkill(ctx, "auto-desc-fork"))

		tool, found := registry.GetSkillTool("auto-desc-fork")
		require.True(t, found)
		assert.Contains(t, tool.Description, "operator")
	})
}

func TestRegistry_CreateSkill_ForkValidation(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)
	ctx := context.Background()

	t.Run("fork type accepted", func(t *testing.T) {
		t.Parallel()

		err := registry.CreateSkill(ctx, SkillEntry{
			Name: "valid-fork",
			Type: SkillTypeFork,
			Definition: map[string]interface{}{
				"instruction": "Do the thing",
			},
		})
		assert.NoError(t, err)
	})

	t.Run("fork with empty definition rejected", func(t *testing.T) {
		t.Parallel()

		err := registry.CreateSkill(ctx, SkillEntry{
			Name:       "empty-def-fork",
			Type:       SkillTypeFork,
			Definition: map[string]interface{}{},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skill definition is required")
	})
}

func TestRegistry_LoadProjectSkills(t *testing.T) {
	t.Parallel()

	t.Run("loads project-local skills", func(t *testing.T) {
		t.Parallel()

		registry := newTestRegistry(t)
		ctx := context.Background()

		// Set up a project directory with a skill.
		projectRoot := t.TempDir()
		skillDir := filepath.Join(projectRoot, ".lango", "skills", "proj-tool")
		require.NoError(t, os.MkdirAll(skillDir, 0o700))
		skillMD := []byte("---\nname: proj-tool\ndescription: Project tool\ntype: instruction\nstatus: active\n---\n\nProject tool reference.\n")
		require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), skillMD, 0o644))

		require.NoError(t, registry.LoadProjectSkills(ctx, projectRoot))

		// Verify project skill is available.
		tool, found := registry.GetSkillTool("proj-tool")
		require.True(t, found)
		assert.Equal(t, "skill_proj-tool", tool.Name)

		// AllTools should include base + project skill.
		all := registry.AllTools()
		require.Len(t, all, 2) // 1 base + 1 project
	})

	t.Run("global skill wins on name conflict", func(t *testing.T) {
		t.Parallel()

		registry := newTestRegistry(t)
		ctx := context.Background()

		// Create and activate a global skill.
		require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
			Name:        "conflict-skill",
			Description: "Global version",
			Type:        "template",
			Definition:  map[string]interface{}{"template": "global"},
		}))
		require.NoError(t, registry.ActivateSkill(ctx, "conflict-skill"))

		// Set up a project skill with the same name.
		projectRoot := t.TempDir()
		skillDir := filepath.Join(projectRoot, ".lango", "skills", "conflict-skill")
		require.NoError(t, os.MkdirAll(skillDir, 0o700))
		skillMD := []byte("---\nname: conflict-skill\ndescription: Project version\ntype: instruction\nstatus: active\n---\n\nProject version content.\n")
		require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), skillMD, 0o644))

		require.NoError(t, registry.LoadProjectSkills(ctx, projectRoot))

		// Should still only have the global version — project-local was skipped.
		tool, found := registry.GetSkillTool("conflict-skill")
		require.True(t, found)
		assert.Equal(t, "Global version", tool.Description)

		// Total loaded skills should be 1 (only the global one).
		loaded := registry.LoadedSkills()
		require.Len(t, loaded, 1)
	})

	t.Run("missing project directory is not an error", func(t *testing.T) {
		t.Parallel()

		registry := newTestRegistry(t)
		ctx := context.Background()

		projectRoot := filepath.Join(t.TempDir(), "nonexistent")
		err := registry.LoadProjectSkills(ctx, projectRoot)
		require.NoError(t, err)

		// No skills should have been added.
		loaded := registry.LoadedSkills()
		require.Empty(t, loaded)
	})

	t.Run("empty project skills directory", func(t *testing.T) {
		t.Parallel()

		registry := newTestRegistry(t)
		ctx := context.Background()

		projectRoot := t.TempDir()
		skillsDir := filepath.Join(projectRoot, ".lango", "skills")
		require.NoError(t, os.MkdirAll(skillsDir, 0o700))

		err := registry.LoadProjectSkills(ctx, projectRoot)
		require.NoError(t, err)

		loaded := registry.LoadedSkills()
		require.Empty(t, loaded)
	})
}

func TestSkillToTool_Capability_InstructionSkill(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)
	ctx := context.Background()

	require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
		Name:        "cap-instruction",
		Description: "Instruction ref",
		Type:        "instruction",
		Definition:  map[string]interface{}{"content": "Reference content"},
	}))
	require.NoError(t, registry.ActivateSkill(ctx, "cap-instruction"))

	tool, found := registry.GetSkillTool("cap-instruction")
	require.True(t, found)

	assert.Equal(t, "skill", tool.Capability.Category)
	assert.Equal(t, agent.ActivityRead, tool.Capability.Activity)
	assert.True(t, tool.Capability.ReadOnly)
	assert.Contains(t, tool.Capability.SearchHints, "cap-instruction")
}

func TestSkillToTool_Capability_ExecuteSkills(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)
	ctx := context.Background()

	tests := []struct {
		give  string
		entry SkillEntry
	}{
		{
			give: "composite skill",
			entry: SkillEntry{
				Name:        "cap-composite",
				Description: "Composite",
				Type:        "composite",
				Definition: map[string]interface{}{
					"steps": []interface{}{
						map[string]interface{}{"tool": "read", "params": map[string]interface{}{"path": "/tmp"}},
					},
				},
			},
		},
		{
			give: "template skill",
			entry: SkillEntry{
				Name:        "cap-template",
				Description: "Template",
				Type:        "template",
				Definition:  map[string]interface{}{"template": "Hello"},
			},
		},
		{
			give: "script skill",
			entry: SkillEntry{
				Name:        "cap-script",
				Description: "Script",
				Type:        "script",
				Definition:  map[string]interface{}{"script": "echo hi"},
			},
		},
		{
			give: "fork skill",
			entry: SkillEntry{
				Name:        "cap-fork",
				Description: "Fork",
				Type:        SkillTypeFork,
				Agent:       "deployer",
				Definition: map[string]interface{}{
					"instruction": "Deploy",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			require.NoError(t, registry.CreateSkill(ctx, tt.entry))
			require.NoError(t, registry.ActivateSkill(ctx, tt.entry.Name))

			tool, found := registry.GetSkillTool(tt.entry.Name)
			require.True(t, found)

			assert.Equal(t, "skill", tool.Capability.Category)
			assert.Equal(t, agent.ActivityExecute, tool.Capability.Activity)
			assert.False(t, tool.Capability.ReadOnly)
			assert.Contains(t, tool.Capability.SearchHints, tt.entry.Name)
		})
	}
}

func TestSkillToTool_Capability_SearchHints_IncludesAllowedTools(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)
	ctx := context.Background()

	require.NoError(t, registry.CreateSkill(ctx, SkillEntry{
		Name:        "hints-skill",
		Description: "Skill with allowed tools",
		Type:        SkillTypeFork,
		Agent:       "ops",
		Definition: map[string]interface{}{
			"instruction": "Do something",
		},
		AllowedTools: []string{"bash", "read_file"},
	}))
	require.NoError(t, registry.ActivateSkill(ctx, "hints-skill"))

	tool, found := registry.GetSkillTool("hints-skill")
	require.True(t, found)

	assert.Contains(t, tool.Capability.SearchHints, "hints-skill")
	assert.Contains(t, tool.Capability.SearchHints, "bash")
	assert.Contains(t, tool.Capability.SearchHints, "read_file")
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
