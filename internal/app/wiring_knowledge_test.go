package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/search"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/testutil"
)

func TestInitKnowledge_ConfigAndStoreBranches(t *testing.T) {
	t.Parallel()

	disabledCfg := &config.Config{}
	components, status := initKnowledge(
		disabledCfg,
		testutil.NewMockSessionStore(),
		nil,
		eventbus.New(),
		nil,
	)
	require.Nil(t, components)
	require.NotNil(t, status)
	assert.False(t, status.Enabled)
	assert.True(t, status.Healthy)

	enabledCfg := &config.Config{Knowledge: config.KnowledgeConfig{Enabled: true}}
	components, status = initKnowledge(
		enabledCfg,
		testutil.NewMockSessionStore(),
		nil,
		eventbus.New(),
		nil,
	)
	require.Nil(t, components)
	require.NotNil(t, status)
	assert.False(t, status.Enabled)
	assert.False(t, status.Healthy)
	assert.Equal(t, "requires EntStore backend", status.Reason)

	entStore := newWiringKnowledgeEntStore(t)
	components, status = initKnowledge(enabledCfg, entStore, nil, eventbus.New(), nil)
	require.NotNil(t, components)
	require.NotNil(t, components.store)
	require.NotNil(t, components.engine)
	require.NotNil(t, components.observer)
	require.NotNil(t, status)
	assert.True(t, status.Enabled)
	assert.True(t, status.Healthy)
}

func TestInitFTS5_BulkIndexesExistingKnowledgeAndLearnings(t *testing.T) {
	entStore := newWiringKnowledgeEntStore(t)
	rawDB := entStore.DB()
	if !search.ProbeFTS5(rawDB) {
		t.Skip("FTS5 not available in current SQLite runtime")
	}

	ctx := context.Background()
	store := knowledge.NewStore(entStore.Client(), testLog())
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", knowledge.KnowledgeEntry{
		Key:      "release-playbook",
		Category: entknowledge.CategoryFact,
		Content:  "legacy release instructions should not be latest",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", knowledge.KnowledgeEntry{
		Key:      "release-playbook",
		Category: entknowledge.CategoryFact,
		Content:  "launch checklist requires smoke tests",
	}))
	require.NoError(t, store.SaveLearning(ctx, "session-1", knowledge.LearningEntry{
		Trigger:      "deploy timeout",
		ErrorPattern: "context deadline exceeded during deployment",
		Fix:          "increase deploy timeout and retry once",
		Category:     entlearning.CategoryTimeout,
	}))

	knowledgeIdx := search.NewFTS5Index(rawDB, "knowledge_fts", []string{"key", "content"})
	require.NoError(t, knowledgeIdx.EnsureTable())
	require.NoError(t, knowledgeIdx.Insert(ctx, "stale-knowledge", []string{"stale", "obsolete"}))
	learningIdx := search.NewFTS5Index(rawDB, "learning_fts", []string{
		"trigger",
		"error_pattern",
		"fix",
	})
	require.NoError(t, learningIdx.EnsureTable())
	require.NoError(t, learningIdx.Insert(ctx, "stale-learning", []string{
		"stale",
		"obsolete",
		"ignore",
	}))

	require.True(t, initFTS5(ctx, rawDB, store))

	assert.Equal(t, []string{"release-playbook"}, fts5SourceIDs(t, rawDB, "knowledge_fts", "launch"))
	assert.Empty(t, fts5SourceIDs(t, rawDB, "knowledge_fts", "legacy"))
	assert.Empty(t, fts5SourceIDs(t, rawDB, "knowledge_fts", "obsolete"))

	learningIDs := fts5SourceIDs(t, rawDB, "learning_fts", "deadline")
	require.Len(t, learningIDs, 1)
	assert.NotEqual(t, "stale-learning", learningIDs[0])
	assert.Empty(t, fts5SourceIDs(t, rawDB, "learning_fts", "obsolete"))
}

func TestInitSkills_LoadsUserSkillsAndSkipsExtPacksWithoutRegistry(t *testing.T) {
	t.Parallel()

	skillsDir := t.TempDir()
	writeWiringKnowledgeSkill(t, skillsDir, "wiringKnowledge5-user", "User-authored knowledge skill")
	writeWiringKnowledgeExtSkill(t, skillsDir, "rogue-pack", "rogue", "Extension skill without registry")

	cfg := &config.Config{
		Skill: config.SkillConfig{
			Enabled:   true,
			SkillsDir: skillsDir,
		},
	}
	baseTool := &agent.Tool{Name: "base_tool"}
	registry, err := initSkills(cfg, []*agent.Tool{baseTool}, eventbus.New(), nil)
	require.NoError(t, err)
	require.NotNil(t, registry)

	_, ok := registry.GetSkillTool("wiringKnowledge5-user")
	assert.True(t, ok)
	_, ok = registry.GetSkillTool("rogue")
	assert.False(t, ok)

	toolNames := make(map[string]bool)
	for _, tool := range registry.AllTools() {
		toolNames[tool.Name] = true
	}
	assert.True(t, toolNames["base_tool"])
	assert.True(t, toolNames["skill_wiringKnowledge5-user"])
	assert.False(t, toolNames["skill_rogue"])

	infos, err := (&skillProviderAdapter{registry: registry}).ListActiveSkillInfos(context.Background())
	require.NoError(t, err)
	assert.Contains(t, skillInfoNames(infos), "wiringKnowledge5-user")
	assert.NotContains(t, skillInfoNames(infos), "rogue")
}

func TestInitSkills_DisabledReturnsNilRegistry(t *testing.T) {
	t.Parallel()

	registry, err := initSkills(&config.Config{}, nil, eventbus.New(), nil)
	require.NoError(t, err)
	assert.Nil(t, registry)
}

func TestInquiryProviderAdapter_ConvertsPendingInquiriesToContextItems(t *testing.T) {
	t.Parallel()

	entStore := newWiringKnowledgeEntStore(t)
	ctx := context.Background()
	store := librarian.NewInquiryStore(entStore.Client(), testLog())
	require.NoError(t, store.SaveInquiry(ctx, librarian.Inquiry{
		SessionKey: "session-knowledge",
		Topic:      "release-readiness",
		Question:   "Which deployment checklist should be used?",
		Context:    "deployment planning",
		Priority:   "high",
	}))
	require.NoError(t, store.SaveInquiry(ctx, librarian.Inquiry{
		SessionKey: "other-session",
		Topic:      "unrelated",
		Question:   "Should not be returned",
		Priority:   "low",
	}))

	items, err := (&inquiryProviderAdapter{store: store}).PendingInquiryItems(
		ctx,
		"session-knowledge",
		10,
	)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, knowledge.LayerPendingInquiries, items[0].Layer)
	assert.Equal(t, "release-readiness", items[0].Key)
	assert.Equal(t, "Which deployment checklist should be used?", items[0].Content)
	assert.Equal(t, "deployment planning", items[0].Source)
}

func TestRunSummaryProviderAdapter_FiltersTerminalRuns(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := runledger.NewMemoryStore()

	seed := func(runID string, status runledger.RunStatus) {
		require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
			RunID:   runID,
			Type:    runledger.EventRunCreated,
			Payload: mustJSON(t, runledger.RunCreatedPayload{SessionKey: "sess-1", Goal: runID}),
		}))
		require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
			RunID: runID,
			Type:  runledger.EventPlanAttached,
			Payload: mustJSON(t, runledger.PlanAttachedPayload{
				Steps: []runledger.Step{{
					StepID:     "s1",
					Goal:       "work",
					OwnerAgent: "operator",
					Status:     runledger.StepStatusPending,
					Validator:  runledger.ValidatorSpec{Type: runledger.ValidatorBuildPass},
					MaxRetries: runledger.DefaultMaxRetries,
				}},
			}),
		}))
		switch status {
		case runledger.RunStatusPaused:
			require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
				RunID:   runID,
				Type:    runledger.EventRunPaused,
				Payload: mustJSON(t, runledger.RunPausedPayload{Reason: "paused"}),
			}))
		case runledger.RunStatusCompleted:
			require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
				RunID:   runID,
				Type:    runledger.EventRunCompleted,
				Payload: mustJSON(t, runledger.RunCompletedPayload{Summary: "done"}),
			}))
		case runledger.RunStatusFailed:
			require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
				RunID:   runID,
				Type:    runledger.EventRunFailed,
				Payload: mustJSON(t, runledger.RunFailedPayload{Reason: "failed"}),
			}))
		}
	}

	seed("run-running", runledger.RunStatusRunning)
	seed("run-paused", runledger.RunStatusPaused)
	seed("run-completed", runledger.RunStatusCompleted)
	seed("run-failed", runledger.RunStatusFailed)

	adapter := &runSummaryProviderAdapter{store: store}
	summaries, err := adapter.ListRunSummaries(ctx, "sess-1", 10)
	require.NoError(t, err)
	require.Len(t, summaries, 2)

	assert.ElementsMatch(t, []string{"run-running", "run-paused"}, []string{
		summaries[0].RunID,
		summaries[1].RunID,
	})
}

func mustJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}

func newWiringKnowledgeEntStore(t *testing.T) *session.EntStore {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "wiringKnowledge5.db")
	store, err := session.NewEntStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})
	return store
}

func fts5SourceIDs(t *testing.T, db *sql.DB, tableName, match string) []string {
	t.Helper()

	rows, err := db.QueryContext(
		context.Background(),
		"SELECT source_id FROM "+tableName+" WHERE "+tableName+" MATCH ? ORDER BY source_id",
		match,
	)
	require.NoError(t, err)
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		require.NoError(t, rows.Scan(&id))
		ids = append(ids, id)
	}
	require.NoError(t, rows.Err())
	return ids
}

func writeWiringKnowledgeSkill(t *testing.T, root, name, description string) {
	t.Helper()

	dir := filepath.Join(root, name)
	require.NoError(t, os.MkdirAll(dir, 0o700))
	content := "---\n" +
		"name: " + name + "\n" +
		"description: " + description + "\n" +
		"type: instruction\n" +
		"status: active\n" +
		"---\n\n" +
		"Use this skill for deterministic wiring tests.\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o600))
}

func writeWiringKnowledgeExtSkill(t *testing.T, root, pack, name, description string) {
	t.Helper()

	dir := filepath.Join(root, "ext-"+pack, name)
	require.NoError(t, os.MkdirAll(dir, 0o700))
	content := "---\n" +
		"name: " + name + "\n" +
		"description: " + description + "\n" +
		"type: instruction\n" +
		"status: active\n" +
		"---\n\n" +
		"Extension-owned skill that should require an allowed pack.\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o600))
}

func skillInfoNames(infos []knowledge.SkillInfo) []string {
	names := make([]string, 0, len(infos))
	for _, info := range infos {
		names = append(names, info.Name)
	}
	return names
}
