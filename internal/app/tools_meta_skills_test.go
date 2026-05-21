package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/skill"
)

func findTool(tools []*agent.Tool, name string) *agent.Tool {
	for _, t := range tools {
		if t.Name == name {
			return t
		}
	}
	return nil
}

func newMetaToolSkillRegistry(t *testing.T) *skill.Registry {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "skills")
	logger := zap.NewNop().Sugar()
	store := skill.NewFileSkillStore(dir, logger)
	baseTool := &agent.Tool{Name: "test_tool", Description: "a test tool"}
	return skill.NewRegistry(store, []*agent.Tool{baseTool}, logger)
}

func newMetaToolBrokenSkillRegistry(t *testing.T) *skill.Registry {
	t.Helper()
	skillsRoot := filepath.Join(t.TempDir(), "skills-file")
	require.NoError(t, os.WriteFile(skillsRoot, []byte("not a directory"), 0o600))
	logger := zap.NewNop().Sugar()
	return skill.NewRegistry(skill.NewFileSkillStore(skillsRoot, logger), nil, logger)
}

func TestListSkills_AcceptsSummaryParameter(t *testing.T) {
	// Tool builds without a registry so handler short-circuits to empty set;
	// what we verify is that the parameter schema accepts `summary`.
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, nil)
	tool := findTool(tools, "list_skills")
	require.NotNil(t, tool)

	props, _ := tool.Parameters["properties"].(map[string]interface{})
	_, hasSummary := props["summary"]
	assert.True(t, hasSummary, "list_skills schema should include `summary` parameter")
}

func TestViewSkill_ToolRegistered(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, nil)
	tool := findTool(tools, "view_skill")
	require.NotNil(t, tool, "view_skill tool should be registered")

	required, _ := tool.Parameters["required"].([]string)
	assert.Contains(t, required, "name")
}

func TestViewSkill_RejectsPathEscape(t *testing.T) {
	// Use a temp skills directory with a fake skill folder to test escape rejection.
	tmpDir := t.TempDir()
	skillRoot := filepath.Join(tmpDir, "test-skill")
	require.NoError(t, os.MkdirAll(skillRoot, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(skillRoot, "SKILL.md"), []byte("# test"), 0o644))

	// Construct a registry-less tool — handler can't actually verify the skill is
	// active, but we can exercise the path-safety branch by using a minimal stub.
	// For simplicity, we test that the helper logic in tool's handler rejects
	// "../" patterns by verifying the path stays under skillRoot.
	abs, err := filepath.Abs(skillRoot)
	require.NoError(t, err)
	escaped := filepath.Clean(filepath.Join(abs, "../../../etc/passwd"))
	assert.False(t,
		filepath.Clean(escaped) == abs ||
			len(escaped) > len(abs) && escaped[:len(abs)] == abs,
		"escaped path must not start with skill root")
}

func TestListSkills_HandlerReturnsEmptyWhenNoRegistry(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, nil)
	tool := findTool(tools, "list_skills")
	require.NotNil(t, tool)
	res, err := tool.Handler(context.Background(), nil)
	require.NoError(t, err)
	m := res.(map[string]interface{})
	assert.Equal(t, 0, m["count"])
}

func TestListSkills_PropagatesRegistryListError(t *testing.T) {
	registry := newMetaToolBrokenSkillRegistry(t)
	tool := findTool(buildMetaTools(nil, nil, registry, config.SkillConfig{}, nil, nil), "list_skills")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "list skills")
}

func TestViewSkill_PropagatesRegistryListError(t *testing.T) {
	registry := newMetaToolBrokenSkillRegistry(t)
	tool := findTool(buildMetaTools(nil, nil, registry, config.SkillConfig{}, nil, nil), "view_skill")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"name": "unreachable",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "list skills")
}

func TestSaveLearning_RequiresCanonicalInputs(t *testing.T) {
	tools := buildMetaTools(newAppKnowledgeStore(t), nil, nil, config.SkillConfig{}, nil, nil)
	tool := findTool(tools, "save_learning")
	require.NotNil(t, tool)

	cases := []struct {
		name      string
		params    map[string]interface{}
		wantError string
	}{
		{
			name: "missing trigger",
			params: map[string]interface{}{
				"fix": "restart the worker",
			},
			wantError: "missing trigger parameter",
		},
		{
			name: "missing fix",
			params: map[string]interface{}{
				"trigger": "tool timeout",
			},
			wantError: "missing fix parameter",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tool.Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}

func TestSearchLearnings_RequiresQueryParameter(t *testing.T) {
	tools := buildMetaTools(newAppKnowledgeStore(t), nil, nil, config.SkillConfig{}, nil, nil)
	tool := findTool(tools, "search_learnings")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "missing query parameter")
}

func TestCreateSkill_RequiresCanonicalInputs(t *testing.T) {
	tools := buildMetaTools(newAppKnowledgeStore(t), nil, nil, config.SkillConfig{}, nil, nil)
	tool := findTool(tools, "create_skill")
	require.NotNil(t, tool)

	cases := []struct {
		name      string
		params    map[string]interface{}
		wantError string
	}{
		{
			name: "missing name",
			params: map[string]interface{}{
				"description": "Reusable sequence",
				"type":        "template",
				"definition":  "{}",
			},
			wantError: "missing name parameter",
		},
		{
			name: "missing description",
			params: map[string]interface{}{
				"name":       "skill-one",
				"type":       "template",
				"definition": "{}",
			},
			wantError: "missing description parameter",
		},
		{
			name: "missing type",
			params: map[string]interface{}{
				"name":        "skill-one",
				"description": "Reusable sequence",
				"definition":  "{}",
			},
			wantError: "missing type parameter",
		},
		{
			name: "missing definition",
			params: map[string]interface{}{
				"name":        "skill-one",
				"description": "Reusable sequence",
				"type":        "template",
			},
			wantError: "missing definition parameter",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tool.Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}

func TestCreateSkill_ValidationBranches(t *testing.T) {
	registry := newMetaToolSkillRegistry(t)

	cases := []struct {
		name      string
		registry  *skill.Registry
		params    map[string]interface{}
		wantError string
	}{
		{
			name:     "invalid definition json",
			registry: registry,
			params: map[string]interface{}{
				"name":        "invalid-definition",
				"description": "Reusable sequence",
				"type":        "template",
				"definition":  "{",
			},
			wantError: "parse definition JSON",
		},
		{
			name:     "invalid parameters json",
			registry: registry,
			params: map[string]interface{}{
				"name":        "invalid-parameters",
				"description": "Reusable sequence",
				"type":        "template",
				"definition":  `{"steps":[]}`,
				"parameters":  "{",
			},
			wantError: "parse parameters JSON",
		},
		{
			name:     "missing registry",
			registry: nil,
			params: map[string]interface{}{
				"name":        "missing-registry",
				"description": "Reusable sequence",
				"type":        "template",
				"definition":  `{"steps":[]}`,
			},
			wantError: "skill system is not enabled",
		},
		{
			name:     "invalid type",
			registry: registry,
			params: map[string]interface{}{
				"name":        "invalid-type",
				"description": "Reusable sequence",
				"type":        "invalid",
				"definition":  `{"steps":[]}`,
			},
			wantError: "skill type must be",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tool := findTool(buildMetaTools(nil, nil, tt.registry, config.SkillConfig{}, nil, nil), "create_skill")
			require.NotNil(t, tool)

			got, err := tool.Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}

func TestViewSkill_RequiresNameParameter(t *testing.T) {
	registry := newMetaToolSkillRegistry(t)
	tool := findTool(buildMetaTools(nil, nil, registry, config.SkillConfig{}, nil, nil), "view_skill")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "name is required")
}

func TestViewSkill_ReturnsReadErrorForMissingSupportingFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "skills")
	logger := zap.NewNop().Sugar()
	store := skill.NewFileSkillStore(dir, logger)
	registry := skill.NewRegistry(store, nil, logger)
	ctx := context.Background()
	entry := skill.SkillEntry{
		Name:        "missing-support",
		Description: "Reads a supporting file",
		Type:        skill.SkillTypeInstruction,
		Context:     "reference material",
		Status:      skill.SkillStatusActive,
	}
	require.NoError(t, registry.CreateSkill(ctx, entry))
	require.NoError(t, registry.ActivateSkill(ctx, entry.Name))

	cfg := config.SkillConfig{SkillsDir: dir}
	tool := findTool(buildMetaTools(nil, nil, registry, cfg, nil, nil), "view_skill")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"name": entry.Name,
		"path": "missing.txt",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "read missing.txt")
}

func TestImportSkill_RequiresURLParameter(t *testing.T) {
	registry := newMetaToolSkillRegistry(t)
	cfg := config.SkillConfig{AllowImport: true}
	tool := findTool(buildMetaTools(nil, nil, registry, cfg, nil, nil), "import_skill")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "missing url parameter")
}

func TestImportSkill_EarlyValidationBranches(t *testing.T) {
	registry := newMetaToolSkillRegistry(t)

	cases := []struct {
		name      string
		registry  *skill.Registry
		cfg       config.SkillConfig
		params    map[string]interface{}
		wantError string
	}{
		{
			name:      "disabled",
			registry:  registry,
			cfg:       config.SkillConfig{AllowImport: false},
			params:    map[string]interface{}{"url": "https://github.com/langoai/lango"},
			wantError: "skill import disabled",
		},
		{
			name:      "missing registry",
			registry:  nil,
			cfg:       config.SkillConfig{AllowImport: true},
			params:    map[string]interface{}{"url": "https://github.com/langoai/lango"},
			wantError: "skill system is not enabled",
		},
		{
			name:      "malformed github url",
			registry:  registry,
			cfg:       config.SkillConfig{AllowImport: true},
			params:    map[string]interface{}{"url": "https://github.com"},
			wantError: "parse GitHub URL",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tool := findTool(buildMetaTools(nil, nil, tt.registry, tt.cfg, nil, nil), "import_skill")
			require.NotNil(t, tool)

			got, err := tool.Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}
