package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/skill"
)

func TestBuildMetaToolsEarlyKnowledgeHandlers_ReturnHistoryAndEmptySearch(t *testing.T) {
	ctx := context.Background()
	store := newAppKnowledgeStore(t)
	require.NoError(t, store.SaveKnowledge(ctx, "", knowledge.KnowledgeEntry{
		Key:         "meta-early-knowledge-history",
		Category:    entknowledge.CategoryFact,
		Content:     "first version",
		SourceClass: "private-confidential",
		AssetLabel:  "meta-early-knowledge-history",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "", knowledge.KnowledgeEntry{
		Key:         "meta-early-knowledge-history",
		Category:    entknowledge.CategoryFact,
		Content:     "second version",
		SourceClass: "private-confidential",
		AssetLabel:  "meta-early-knowledge-history",
	}))

	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, nil, nil)
	historyTool := findTool(tools, "get_knowledge_history")
	searchTool := findTool(tools, "search_knowledge")
	require.NotNil(t, historyTool)
	require.NotNil(t, searchTool)

	historyResult, err := historyTool.Handler(ctx, map[string]interface{}{
		"key": "meta-early-knowledge-history",
	})
	require.NoError(t, err)
	historyPayload, ok := historyResult.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "meta-early-knowledge-history", historyPayload["key"])
	versions, ok := historyPayload["versions"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, versions, 2)
	assert.Equal(t, 2, versions[0]["version"])
	assert.Equal(t, "second version", versions[0]["content"])

	searchResult, err := searchTool.Handler(ctx, map[string]interface{}{
		"query":    "zzzzunmatchedzzzz",
		"category": "rule",
	})
	require.NoError(t, err)
	searchPayload, ok := searchResult.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, searchPayload["count"])
}

func TestBuildMetaToolsEarlyKnowledgeHandlers_PropagateStoreErrors(t *testing.T) {
	store := newAppKnowledgeStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cases := []struct {
		name      string
		toolName  string
		params    map[string]interface{}
		wantError string
	}{
		{
			name:     "save knowledge",
			toolName: "save_knowledge",
			params: map[string]interface{}{
				"key":      "meta-early-canceled-save",
				"category": "fact",
				"content":  "store should reject canceled context",
			},
			wantError: "save knowledge",
		},
		{
			name:     "history",
			toolName: "get_knowledge_history",
			params: map[string]interface{}{
				"key": "meta-early-canceled-history",
			},
			wantError: "get knowledge history",
		},
		{
			name:     "search knowledge",
			toolName: "search_knowledge",
			params: map[string]interface{}{
				"query": "meta early canceled search",
			},
			wantError: "search knowledge",
		},
		{
			name:     "save learning",
			toolName: "save_learning",
			params: map[string]interface{}{
				"trigger": "meta early canceled learning",
				"fix":     "retry later",
			},
			wantError: "save learning",
		},
		{
			name:     "search learnings",
			toolName: "search_learnings",
			params: map[string]interface{}{
				"query": "meta early canceled learning",
			},
			wantError: "search learnings",
		},
		{
			name:      "learning stats",
			toolName:  "learning_stats",
			params:    map[string]interface{}{},
			wantError: "get learning stats",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tool := findTool(tools, tt.toolName)
			require.NotNil(t, tool)

			got, err := tool.Handler(ctx, tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}

func TestBuildMetaToolsEarlyLearningHandlers_SaveSearchStatsAndCleanupBulk(t *testing.T) {
	ctx := context.Background()
	store := newAppKnowledgeStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, nil, nil)

	saveTool := findTool(tools, "save_learning")
	searchTool := findTool(tools, "search_learnings")
	statsTool := findTool(tools, "learning_stats")
	cleanupTool := findTool(tools, "learning_cleanup")
	require.NotNil(t, saveTool)
	require.NotNil(t, searchTool)
	require.NotNil(t, statsTool)
	require.NotNil(t, cleanupTool)

	saveResult, err := saveTool.Handler(ctx, map[string]interface{}{
		"trigger":       "meta early learning timeout",
		"error_pattern": "deadline exceeded",
		"diagnosis":     "worker waited too long",
		"fix":           "retry with bounded timeout",
		"category":      string(entlearning.CategoryToolError),
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"status":  "saved",
		"message": "Learning for 'meta early learning timeout' saved successfully",
	}, saveResult)

	searchResult, err := searchTool.Handler(ctx, map[string]interface{}{
		"query":    "deadline",
		"category": string(entlearning.CategoryToolError),
	})
	require.NoError(t, err)
	searchPayload, ok := searchResult.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, searchPayload["count"])

	statsResult, err := statsTool.Handler(ctx, map[string]interface{}{})
	require.NoError(t, err)
	stats, ok := statsResult.(*knowledge.LearningStats)
	require.True(t, ok)
	assert.Equal(t, 1, stats.TotalCount)

	require.NoError(t, store.SaveLearning(ctx, "meta-early-high-confidence-session", knowledge.LearningEntry{
		Trigger:  "meta early high confidence cleanup survivor",
		Category: entlearning.CategoryToolError,
	}))
	highConfidenceEntries, err := store.SearchLearningEntities(ctx, "meta early high confidence cleanup survivor", 1)
	require.NoError(t, err)
	require.Len(t, highConfidenceEntries, 1)
	require.NoError(t, store.BoostLearningConfidence(ctx, highConfidenceEntries[0].ID, 1, 0.5))

	dryRunResult, err := cleanupTool.Handler(ctx, map[string]interface{}{
		"category":       string(entlearning.CategoryToolError),
		"max_confidence": float64(0.5),
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"would_delete": 1, "dry_run": true}, dryRunResult)

	ageDryRunResult, err := cleanupTool.Handler(ctx, map[string]interface{}{
		"category":        string(entlearning.CategoryToolError),
		"older_than_days": float64(30),
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"would_delete": 0, "dry_run": true}, ageDryRunResult)

	deleteResult, err := cleanupTool.Handler(ctx, map[string]interface{}{
		"category":       string(entlearning.CategoryToolError),
		"max_confidence": float64(0.5),
		"dry_run":        false,
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"deleted": 1, "dry_run": false}, deleteResult)
}

func TestBuildMetaToolsEarlyLearningCleanup_RejectsInvalidID(t *testing.T) {
	tool := findTool(
		buildMetaTools(newAppKnowledgeStore(t), nil, nil, config.SkillConfig{}, nil, nil),
		"learning_cleanup",
	)
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"id": "not-a-uuid",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "invalid id")
}

func TestBuildMetaToolsEarlySkillHandlers_CreateListAndViewSkill(t *testing.T) {
	ctx := context.Background()
	store := newAppKnowledgeStore(t)
	skillsDir := t.TempDir()
	registry := newBuildMetaToolsWithRuntimesComposesAvailableRuntimeToolsSkillRegistry(t, skillsDir)
	tools := buildMetaTools(
		store,
		nil,
		registry,
		config.SkillConfig{SkillsDir: skillsDir},
		nil,
		nil,
	)

	createTool := findTool(tools, "create_skill")
	listTool := findTool(tools, "list_skills")
	viewTool := findTool(tools, "view_skill")
	require.NotNil(t, createTool)
	require.NotNil(t, listTool)
	require.NotNil(t, viewTool)

	got, err := createTool.Handler(ctx, map[string]interface{}{
		"name":        "meta-early-created-skill",
		"description": "Created through the meta tool",
		"type":        string(skill.SkillTypeComposite),
		"definition":  `{"steps":["inspect","report"]}`,
		"parameters":  `{"type":"object"}`,
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"status":  "active",
		"name":    "meta-early-created-skill",
		"message": "Skill 'meta-early-created-skill' created and activated",
	}, got)

	listResult, err := listTool.Handler(ctx, map[string]interface{}{})
	require.NoError(t, err)
	listPayload, ok := listResult.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, listPayload["count"])
	listed, ok := listPayload["skills"].([]skill.SkillEntry)
	require.True(t, ok)
	require.Len(t, listed, 1)
	assert.Equal(t, "meta-early-created-skill", listed[0].Name)

	viewResult, err := viewTool.Handler(ctx, map[string]interface{}{
		"name": "meta-early-created-skill",
	})
	require.NoError(t, err)
	viewPayload, ok := viewResult.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "meta-early-created-skill", viewPayload["name"])
	assert.Equal(t, "SKILL.md", filepath.Base(viewPayload["path"].(string)))
	assert.Contains(t, viewPayload["path"], "meta-early-created-skill")
	assert.Contains(t, viewPayload["content"], "Created through the meta tool")
}

func TestBuildMetaToolsEarlySkillHandlers_RejectInvalidDefinitionAndUnavailableDependencies(t *testing.T) {
	ctx := context.Background()
	store := newAppKnowledgeStore(t)
	createTool := findTool(buildMetaTools(store, nil, nil, config.SkillConfig{}, nil, nil), "create_skill")
	require.NotNil(t, createTool)

	got, err := createTool.Handler(ctx, map[string]interface{}{
		"name":        "meta-early-invalid-json",
		"description": "Invalid JSON should be rejected before registry access",
		"type":        string(skill.SkillTypeComposite),
		"definition":  `{`,
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "parse definition JSON")

	got, err = createTool.Handler(ctx, map[string]interface{}{
		"name":        "meta-early-no-registry",
		"description": "Registry is required after valid JSON parsing",
		"type":        string(skill.SkillTypeComposite),
		"definition":  `{}`,
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "skill system is not enabled")

	viewTool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, nil), "view_skill")
	require.NotNil(t, viewTool)
	got, err = viewTool.Handler(ctx, map[string]interface{}{
		"name": "meta-early-missing-registry",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "skill registry not available")

	registry := newMetaToolSkillRegistry(t)
	viewTool = findTool(buildMetaTools(nil, nil, registry, config.SkillConfig{}, nil, nil), "view_skill")
	require.NotNil(t, viewTool)
	got, err = viewTool.Handler(ctx, map[string]interface{}{
		"name": "meta-early-inactive",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, `skill "meta-early-inactive" is not active`)

	importTool := findTool(
		buildMetaTools(nil, nil, nil, config.SkillConfig{AllowImport: true}, nil, nil),
		"import_skill",
	)
	require.NotNil(t, importTool)
	got, err = importTool.Handler(ctx, map[string]interface{}{
		"url": "https://example.invalid/SKILL.md",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "skill system is not enabled")
}

func TestBuildMetaToolsEarlyExportability_RejectsEmptySourceKey(t *testing.T) {
	tool := findTool(
		buildMetaTools(newAppKnowledgeStore(t), nil, nil, config.SkillConfig{}, nil, nil),
		"evaluate_exportability",
	)
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"artifact_label": "meta-early-artifact",
		"source_keys":    []interface{}{"meta-source", " "},
		"stage":          "draft",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "source_keys must not contain empty values")
}
