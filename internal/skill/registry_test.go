package skill

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

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
			err := registry.CreateSkill(ctx, tt.entry)
			if err == nil {
				t.Fatalf("CreateSkill(%q) = nil, want error containing %q", tt.give, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestRegistry_LoadSkills_AllTools(t *testing.T) {
	registry := newTestRegistry(t)
	ctx := context.Background()

	// Before loading any skills, AllTools should return only the base tool.
	toolsBefore := registry.AllTools()
	if len(toolsBefore) != 1 {
		t.Fatalf("AllTools before load: len = %d, want 1", len(toolsBefore))
	}
	if toolsBefore[0].Name != "test_tool" {
		t.Errorf("base tool name = %q, want %q", toolsBefore[0].Name, "test_tool")
	}

	// Create and activate a skill.
	err := registry.CreateSkill(ctx, SkillEntry{
		Name:        "my_skill",
		Description: "does stuff",
		Type:        "template",
		Definition:  map[string]interface{}{"template": "Hello {{.Name}}"},
	})
	if err != nil {
		t.Fatalf("CreateSkill: %v", err)
	}

	err = registry.ActivateSkill(ctx, "my_skill")
	if err != nil {
		t.Fatalf("ActivateSkill: %v", err)
	}

	// After activation (which calls LoadSkills internally), AllTools should include both.
	toolsAfter := registry.AllTools()
	if len(toolsAfter) != 2 {
		t.Fatalf("AllTools after load: len = %d, want 2", len(toolsAfter))
	}

	names := make(map[string]bool, len(toolsAfter))
	for _, tool := range toolsAfter {
		names[tool.Name] = true
	}
	if !names["test_tool"] {
		t.Error("AllTools missing base tool 'test_tool'")
	}
	if !names["skill_my_skill"] {
		t.Error("AllTools missing loaded skill 'skill_my_skill'")
	}
}

func TestRegistry_LoadedSkills(t *testing.T) {
	registry := newTestRegistry(t)
	ctx := context.Background()

	// Before loading any skills, LoadedSkills should return empty (no base tools).
	loaded := registry.LoadedSkills()
	if len(loaded) != 0 {
		t.Fatalf("LoadedSkills before load: len = %d, want 0", len(loaded))
	}

	// Create and activate a skill.
	err := registry.CreateSkill(ctx, SkillEntry{
		Name:        "loaded_skill",
		Description: "test loaded",
		Type:        "template",
		Definition:  map[string]interface{}{"template": "Hi"},
	})
	if err != nil {
		t.Fatalf("CreateSkill: %v", err)
	}

	err = registry.ActivateSkill(ctx, "loaded_skill")
	if err != nil {
		t.Fatalf("ActivateSkill: %v", err)
	}

	// After activation, LoadedSkills should return only the dynamic skill.
	loaded = registry.LoadedSkills()
	if len(loaded) != 1 {
		t.Fatalf("LoadedSkills after load: len = %d, want 1", len(loaded))
	}
	if loaded[0].Name != "skill_loaded_skill" {
		t.Errorf("loaded tool name = %q, want %q", loaded[0].Name, "skill_loaded_skill")
	}

	// AllTools should still include both base and loaded.
	all := registry.AllTools()
	if len(all) != 2 {
		t.Fatalf("AllTools: len = %d, want 2", len(all))
	}
}

func TestRegistry_ActivateSkill(t *testing.T) {
	registry := newTestRegistry(t)
	ctx := context.Background()

	// Create a skill first.
	err := registry.CreateSkill(ctx, SkillEntry{
		Name:        "activate_me",
		Description: "a skill to activate",
		Type:        "composite",
		Definition: map[string]interface{}{
			"steps": []interface{}{
				map[string]interface{}{"tool": "read", "params": map[string]interface{}{"path": "/tmp"}},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateSkill: %v", err)
	}

	// Before activation, GetSkillTool should return false.
	_, found := registry.GetSkillTool("activate_me")
	if found {
		t.Error("GetSkillTool returned true before activation, want false")
	}

	// Activate the skill.
	err = registry.ActivateSkill(ctx, "activate_me")
	if err != nil {
		t.Fatalf("ActivateSkill: %v", err)
	}

	// After activation, GetSkillTool should return the tool.
	tool, found := registry.GetSkillTool("activate_me")
	if !found {
		t.Fatal("GetSkillTool returned false after activation, want true")
	}
	if tool.Name != "skill_activate_me" {
		t.Errorf("tool.Name = %q, want %q", tool.Name, "skill_activate_me")
	}
}

func TestRegistry_GetSkillTool(t *testing.T) {
	registry := newTestRegistry(t)
	ctx := context.Background()

	t.Run("skill_ prefix naming", func(t *testing.T) {
		err := registry.CreateSkill(ctx, SkillEntry{
			Name:        "prefixed",
			Description: "test prefix",
			Type:        "template",
			Definition:  map[string]interface{}{"template": "test"},
		})
		if err != nil {
			t.Fatalf("CreateSkill: %v", err)
		}

		err = registry.ActivateSkill(ctx, "prefixed")
		if err != nil {
			t.Fatalf("ActivateSkill: %v", err)
		}

		tool, found := registry.GetSkillTool("prefixed")
		if !found {
			t.Fatal("GetSkillTool returned false, want true")
		}
		if !strings.HasPrefix(tool.Name, "skill_") {
			t.Errorf("tool.Name = %q, want prefix %q", tool.Name, "skill_")
		}
		if tool.Name != "skill_prefixed" {
			t.Errorf("tool.Name = %q, want %q", tool.Name, "skill_prefixed")
		}
	})

	t.Run("non-existent skill returns false", func(t *testing.T) {
		_, found := registry.GetSkillTool("does_not_exist")
		if found {
			t.Error("GetSkillTool returned true for non-existent skill, want false")
		}
	})
}

func TestRegistry_InstructionSkillAsTool(t *testing.T) {
	registry := newTestRegistry(t)
	ctx := context.Background()

	// Create an instruction skill.
	err := registry.CreateSkill(ctx, SkillEntry{
		Name:        "obsidian-ref",
		Description: "Obsidian Markdown reference guide",
		Type:        "instruction",
		Definition:  map[string]interface{}{"content": "# Obsidian\n\nUse wikilinks."},
		Source:      "https://github.com/owner/repo",
	})
	if err != nil {
		t.Fatalf("CreateSkill: %v", err)
	}

	err = registry.ActivateSkill(ctx, "obsidian-ref")
	if err != nil {
		t.Fatalf("ActivateSkill: %v", err)
	}

	// Verify tool is registered.
	tool, found := registry.GetSkillTool("obsidian-ref")
	if !found {
		t.Fatal("GetSkillTool returned false for instruction skill")
	}
	if tool.Name != "skill_obsidian-ref" {
		t.Errorf("tool.Name = %q, want %q", tool.Name, "skill_obsidian-ref")
	}
}

func TestRegistry_InstructionTool_ReturnsContent(t *testing.T) {
	registry := newTestRegistry(t)
	ctx := context.Background()

	err := registry.CreateSkill(ctx, SkillEntry{
		Name:        "my-guide",
		Description: "My guide",
		Type:        "instruction",
		Definition:  map[string]interface{}{"content": "Guide content here."},
		Source:      "https://example.com/guide",
	})
	if err != nil {
		t.Fatalf("CreateSkill: %v", err)
	}

	err = registry.ActivateSkill(ctx, "my-guide")
	if err != nil {
		t.Fatalf("ActivateSkill: %v", err)
	}

	tool, found := registry.GetSkillTool("my-guide")
	if !found {
		t.Fatal("GetSkillTool returned false")
	}

	// Call the handler.
	result, err := tool.Handler(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Handler: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("result type = %T, want map[string]interface{}", result)
	}

	if resultMap["content"] != "Guide content here." {
		t.Errorf("content = %q, want %q", resultMap["content"], "Guide content here.")
	}
	if resultMap["source"] != "https://example.com/guide" {
		t.Errorf("source = %q, want %q", resultMap["source"], "https://example.com/guide")
	}
	if resultMap["type"] != "instruction" {
		t.Errorf("type = %q, want %q", resultMap["type"], "instruction")
	}
}

func TestRegistry_InstructionTool_Description(t *testing.T) {
	registry := newTestRegistry(t)
	ctx := context.Background()

	t.Run("custom description preserved", func(t *testing.T) {
		err := registry.CreateSkill(ctx, SkillEntry{
			Name:        "custom-desc",
			Description: "Use this when working with Obsidian Markdown syntax",
			Type:        "instruction",
			Definition:  map[string]interface{}{"content": "content"},
		})
		if err != nil {
			t.Fatalf("CreateSkill: %v", err)
		}
		err = registry.ActivateSkill(ctx, "custom-desc")
		if err != nil {
			t.Fatalf("ActivateSkill: %v", err)
		}

		tool, _ := registry.GetSkillTool("custom-desc")
		if tool.Description != "Use this when working with Obsidian Markdown syntax" {
			t.Errorf("Description = %q, want original", tool.Description)
		}
	})

	t.Run("empty description gets default", func(t *testing.T) {
		err := registry.CreateSkill(ctx, SkillEntry{
			Name:       "no-desc",
			Type:       "instruction",
			Definition: map[string]interface{}{"content": "content"},
		})
		if err != nil {
			t.Fatalf("CreateSkill: %v", err)
		}
		err = registry.ActivateSkill(ctx, "no-desc")
		if err != nil {
			t.Fatalf("ActivateSkill: %v", err)
		}

		tool, _ := registry.GetSkillTool("no-desc")
		if tool.Description != "Reference guide for no-desc" {
			t.Errorf("Description = %q, want default", tool.Description)
		}
	})
}

func TestRegistry_ListActiveSkills(t *testing.T) {
	registry := newTestRegistry(t)
	ctx := context.Background()

	// Create and activate a skill.
	err := registry.CreateSkill(ctx, SkillEntry{
		Name:        "listable",
		Description: "a listable skill",
		Type:        "script",
		Status:      "active",
		Definition:  map[string]interface{}{"script": "echo hi"},
	})
	if err != nil {
		t.Fatalf("CreateSkill: %v", err)
	}

	err = registry.ActivateSkill(ctx, "listable")
	if err != nil {
		t.Fatalf("ActivateSkill: %v", err)
	}

	skills, err := registry.ListActiveSkills(ctx)
	if err != nil {
		t.Fatalf("ListActiveSkills: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("len(skills) = %d, want 1", len(skills))
	}
	if skills[0].Name != "listable" {
		t.Errorf("skills[0].Name = %q, want %q", skills[0].Name, "listable")
	}
}
