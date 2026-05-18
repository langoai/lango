package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/skill"
)

func TestWave24BuildMetaToolsWithRuntimes_ComposesAvailableRuntimeTools(t *testing.T) {
	t.Parallel()

	tools := buildMetaToolsWithRuntimes(
		nil,
		nil,
		nil,
		config.SkillConfig{},
		config.DefaultConfig(),
		receipts.NewStore(),
		nil,
		&fakeSettlementExecutionRuntime{},
		nil,
		nil,
		nil,
		nil,
	)

	assert.NotNil(t, findTool(tools, "execute_settlement"))
	assert.NotNil(t, findTool(tools, "execute_partial_settlement"))
	assert.NotNil(t, findTool(tools, "list_dead_lettered_post_adjudication_executions"))
	assert.NotNil(t, findTool(tools, "get_post_adjudication_execution_status"))
	assert.NotNil(t, findTool(tools, "retry_post_adjudication_execution"))
	assert.Nil(t, findTool(tools, "hold_escrow_for_dispute"))
	assert.Nil(t, findTool(tools, "release_escrow_settlement"))
	assert.Nil(t, findTool(tools, "refund_escrow_settlement"))
	assert.Nil(t, findTool(tools, "execute_escrow_recommendation"))
	assertNoDuplicateNames(t, tools)
}

func TestWave24BuildMetaToolsWithRuntimes_ReceiptBackedHandlersRejectMissingStore(t *testing.T) {
	t.Parallel()

	tools := buildMetaToolsWithRuntimes(
		nil,
		nil,
		nil,
		config.SkillConfig{},
		config.DefaultConfig(),
		nil,
		nil,
		&fakeSettlementExecutionRuntime{},
		&fakePartialSettlementExecutionRuntime{},
		&fakeDisputeHoldRuntime{},
		&fakeEscrowReleaseRuntime{},
		&fakeEscrowRefundRuntime{},
	)

	for _, name := range []string{
		"create_dispute_ready_receipt",
		"open_knowledge_exchange_transaction",
		"select_knowledge_exchange_path",
		"approve_upfront_payment",
		"apply_settlement_progression",
		"adjudicate_escrow_dispute",
		"execute_settlement",
		"execute_partial_settlement",
		"hold_escrow_for_dispute",
		"release_escrow_settlement",
		"refund_escrow_settlement",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tool := findTool(tools, name)
			require.NotNil(t, tool)

			got, err := tool.Handler(context.Background(), map[string]interface{}{})
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, "receipts store dependency is not configured")
		})
	}
}

func TestWave24RetryPostAdjudicationExecution_DependencyDisabledBeforeParams(t *testing.T) {
	t.Parallel()

	tool := findTool(
		buildMetaToolsWithRuntimes(
			nil,
			nil,
			nil,
			config.SkillConfig{},
			config.DefaultConfig(),
			receipts.NewStore(),
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		),
		"retry_post_adjudication_execution",
	)
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.EqualError(t, err, "background manager is not configured")
}

func TestWave24ListSkills_FiltersBySessionModeAndReturnsSummary(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	registry := newMetaToolSkillRegistry(t)
	createWave24InstructionSkill(t, registry, "wave24-allowed", "Allowed skill")
	createWave24InstructionSkill(t, registry, "wave24-hidden", "Hidden skill")

	cfg := config.DefaultConfig()
	cfg.Modes = map[string]config.SessionMode{
		"wave24": {
			Name:   "wave24",
			Skills: []string{"wave24-allowed"},
		},
	}

	tool := findTool(buildMetaTools(nil, nil, registry, config.SkillConfig{}, cfg, nil), "list_skills")
	require.NotNil(t, tool)

	got, err := tool.Handler(session.WithModeName(ctx, "wave24"), map[string]interface{}{
		"summary": true,
	})
	require.NoError(t, err)

	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, payload["count"])

	skills, ok := payload["skills"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, skills, 1)
	assert.Equal(t, "wave24-allowed", skills[0]["name"])
	assert.Equal(t, "Allowed skill", skills[0]["description"])
	assert.Equal(t, "Use wave24-allowed for deterministic meta-tool tests.", skills[0]["when_to_use"])
}

func TestWave24ViewSkill_ReturnsContextFallbackAndRejectsPathEscape(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fallbackRegistry := newMetaToolSkillRegistry(t)
	createWave24InstructionSkill(t, fallbackRegistry, "wave24-fallback", "Fallback skill")
	fallbackTool := findTool(
		buildMetaTools(nil, nil, fallbackRegistry, config.SkillConfig{}, nil, nil),
		"view_skill",
	)
	require.NotNil(t, fallbackTool)

	got, err := fallbackTool.Handler(ctx, map[string]interface{}{"name": "wave24-fallback"})
	require.NoError(t, err)
	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "wave24-fallback", payload["name"])
	assert.Equal(t, "Context for wave24-fallback.", payload["content"])
	assert.Equal(t, "skills directory not configured; returning skill Context only", payload["note"])

	skillsDir := t.TempDir()
	registry := newWave24SkillRegistry(t, skillsDir)
	createWave24InstructionSkill(t, registry, "wave24-files", "File skill")
	require.NoError(t, registry.Store().SaveResource(ctx, "wave24-files", "docs/note.txt", []byte("note body")))

	fileTool := findTool(
		buildMetaTools(
			nil,
			nil,
			registry,
			config.SkillConfig{SkillsDir: skillsDir},
			nil,
			nil,
		),
		"view_skill",
	)
	require.NotNil(t, fileTool)

	fileResult, err := fileTool.Handler(ctx, map[string]interface{}{
		"name": "wave24-files",
		"path": filepath.Join("docs", "note.txt"),
	})
	require.NoError(t, err)
	filePayload, ok := fileResult.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "wave24-files", filePayload["name"])
	assert.Equal(t, "note body", filePayload["content"])

	escapedResult, err := fileTool.Handler(ctx, map[string]interface{}{
		"name": "wave24-files",
		"path": filepath.Join("..", "outside.txt"),
	})
	require.Error(t, err)
	assert.Nil(t, escapedResult)
	assert.ErrorContains(t, err, "is outside the skill directory")

	outsideDir := t.TempDir()
	outsidePath := filepath.Join(outsideDir, "outside.txt")
	require.NoError(t, os.WriteFile(outsidePath, []byte("outside body"), 0o600))
	linkPath := filepath.Join(skillsDir, "wave24-files", "docs", "outside-link.txt")
	if err := os.Symlink(outsidePath, linkPath); err != nil {
		t.Skipf("symlinks are not available in this environment: %v", err)
	}
	symlinkResult, err := fileTool.Handler(ctx, map[string]interface{}{
		"name": "wave24-files",
		"path": filepath.Join("docs", "outside-link.txt"),
	})
	require.Error(t, err)
	assert.Nil(t, symlinkResult)
	assert.ErrorContains(t, err, "escapes the skill directory")
}

func TestWave24LearningCleanup_DeletesSingleLearningByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newAppKnowledgeStore(t)
	require.NoError(t, store.SaveLearning(ctx, "", knowledge.LearningEntry{
		Trigger:      "wave24 tool failure",
		ErrorPattern: "timeout",
		Diagnosis:    "worker stalled",
		Fix:          "retry with bounded timeout",
		Category:     entlearning.CategoryGeneral,
	}))

	entries, total, err := store.ListLearnings(ctx, "", 0, time.Time{}, 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, entries, 1)

	tool := findTool(buildMetaTools(store, nil, nil, config.SkillConfig{}, nil, nil), "learning_cleanup")
	require.NotNil(t, tool)

	dryRunResult, err := tool.Handler(ctx, map[string]interface{}{
		"id":      entries[0].ID.String(),
		"dry_run": true,
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"would_delete": 1, "dry_run": true}, dryRunResult)

	_, err = store.GetLearning(ctx, entries[0].ID)
	require.NoError(t, err)

	deleteResult, err := tool.Handler(ctx, map[string]interface{}{
		"id":      entries[0].ID.String(),
		"dry_run": false,
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"deleted": 1, "dry_run": false}, deleteResult)

	_, err = store.GetLearning(ctx, entries[0].ID)
	require.Error(t, err)
	assert.ErrorContains(t, err, "learning not found")
}

func newWave24SkillRegistry(t *testing.T, skillsDir string) *skill.Registry {
	t.Helper()
	require.NoError(t, os.MkdirAll(skillsDir, 0o700))
	logger := zap.NewNop().Sugar()
	store := skill.NewFileSkillStore(skillsDir, logger)
	return skill.NewRegistry(store, []*agent.Tool{{Name: "wave24_base_tool"}}, logger)
}

func createWave24InstructionSkill(t *testing.T, registry *skill.Registry, name, description string) {
	t.Helper()
	err := registry.CreateSkill(context.Background(), skill.SkillEntry{
		Name:        name,
		Description: description,
		Type:        skill.SkillTypeInstruction,
		Status:      skill.SkillStatusActive,
		WhenToUse:   "Use " + name + " for deterministic meta-tool tests.",
		Context:     "Context for " + name + ".",
	})
	require.NoError(t, err)
}
