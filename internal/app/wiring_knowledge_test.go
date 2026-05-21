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
	"github.com/langoai/lango/internal/extension"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/search"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/supervisor"
	"github.com/langoai/lango/internal/testutil"
	_ "github.com/mattn/go-sqlite3"
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

func TestInitFTS5AndBulkIndexErrorBranches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assert.False(t, initFTS5(ctx, nil, nil))

	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "fts-errors.db"))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	if !search.ProbeFTS5(db) {
		t.Skip("FTS5 not available in current SQLite runtime")
	}

	knowledgeIdx := search.NewFTS5Index(db, "knowledge_fts", []string{"key", "content"})
	err = bulkIndexKnowledge(ctx, db, knowledgeIdx)
	require.Error(t, err)
	assert.ErrorContains(t, err, "clear knowledge FTS5")

	require.NoError(t, knowledgeIdx.EnsureTable())
	err = bulkIndexKnowledge(ctx, db, knowledgeIdx)
	require.Error(t, err)
	assert.ErrorContains(t, err, "query knowledge for FTS5 index")

	learningIdx := search.NewFTS5Index(db, "learning_fts", []string{"trigger", "error_pattern", "fix"})
	err = bulkIndexLearnings(ctx, db, learningIdx)
	require.Error(t, err)
	assert.ErrorContains(t, err, "clear learning FTS5")

	require.NoError(t, learningIdx.EnsureTable())
	err = bulkIndexLearnings(ctx, db, learningIdx)
	require.Error(t, err)
	assert.ErrorContains(t, err, "query learnings for FTS5 index")
}

func TestInitFTS5_ReturnsFalseWhenKnowledgeTableCreationFails(t *testing.T) {
	t.Parallel()

	entStore := newWiringKnowledgeEntStore(t)
	rawDB := entStore.DB()
	if !search.ProbeFTS5(rawDB) {
		t.Skip("FTS5 not available in current SQLite runtime")
	}

	require.NoError(t, createFTS5ShadowTableConflict(rawDB, "knowledge_fts"))

	store := knowledge.NewStore(entStore.Client(), testLog())
	assert.False(t, initFTS5(context.Background(), rawDB, store))
}

func TestInitFTS5_ReturnsFalseWhenLearningTableCreationFails(t *testing.T) {
	t.Parallel()

	entStore := newWiringKnowledgeEntStore(t)
	rawDB := entStore.DB()
	if !search.ProbeFTS5(rawDB) {
		t.Skip("FTS5 not available in current SQLite runtime")
	}

	require.NoError(t, createFTS5ShadowTableConflict(rawDB, "learning_fts"))

	store := knowledge.NewStore(entStore.Client(), testLog())
	assert.False(t, initFTS5(context.Background(), rawDB, store))
}

func TestBulkIndexKnowledge_ReturnsBulkInsertErrorForIncompatibleIndexShape(t *testing.T) {
	t.Parallel()

	entStore := newWiringKnowledgeEntStore(t)
	rawDB := entStore.DB()
	if !search.ProbeFTS5(rawDB) {
		t.Skip("FTS5 not available in current SQLite runtime")
	}

	ctx := context.Background()
	store := knowledge.NewStore(entStore.Client(), testLog())
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", knowledge.KnowledgeEntry{
		Key:      "incompatible-knowledge-index",
		Category: entknowledge.CategoryFact,
		Content:  "record forces bulk insert after source query succeeds",
	}))
	require.NoError(t, createIncompatibleKnowledgeFTS5Table(rawDB))

	idx := search.NewFTS5Index(rawDB, "knowledge_fts", []string{"key", "content"})
	err := bulkIndexKnowledge(ctx, rawDB, idx)
	require.Error(t, err)
	assert.ErrorContains(t, err, "bulk insert knowledge FTS5")
}

func TestBulkIndexLearnings_ReturnsBulkInsertErrorForIncompatibleIndexShape(t *testing.T) {
	t.Parallel()

	entStore := newWiringKnowledgeEntStore(t)
	rawDB := entStore.DB()
	if !search.ProbeFTS5(rawDB) {
		t.Skip("FTS5 not available in current SQLite runtime")
	}

	ctx := context.Background()
	store := knowledge.NewStore(entStore.Client(), testLog())
	require.NoError(t, store.SaveLearning(ctx, "session-1", knowledge.LearningEntry{
		Trigger:      "incompatible learning index",
		ErrorPattern: "record forces bulk insert after source query succeeds",
		Fix:          "use matching FTS5 column shape",
		Category:     entlearning.CategoryTimeout,
	}))
	require.NoError(t, createIncompatibleLearningFTS5Table(rawDB))

	idx := search.NewFTS5Index(rawDB, "learning_fts", []string{
		"trigger",
		"error_pattern",
		"fix",
	})
	err := bulkIndexLearnings(ctx, rawDB, idx)
	require.Error(t, err)
	assert.ErrorContains(t, err, "bulk insert learning FTS5")
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

func TestInitSkills_AllowsOnlyHealthyExtensionPacksFromRegistry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	extensionsDir := filepath.Join(root, "extensions")

	inst := &extension.Installer{
		ExtensionsDir: extensionsDir,
		SkillsDir:     skillsDir,
	}
	src := extension.NewLocalSource(writeWiringKnowledgeExtensionPack(t, root))
	_, wc, err := inst.Inspect(context.Background(), src)
	require.NoError(t, err)
	t.Cleanup(func() { _ = wc.Cleanup() })
	require.NoError(t, inst.Install(context.Background(), src, wc, extension.InstallOptions{}))

	writeWiringKnowledgeExtSkill(t, skillsDir, "rogue-pack", "rogue-skill", "Untrusted extension skill")
	extReg, err := extension.LoadRegistry(extensionsDir, false)
	require.NoError(t, err)
	require.Len(t, extReg.OKPacks(), 1)

	cfg := &config.Config{
		Skill: config.SkillConfig{
			Enabled:   true,
			SkillsDir: skillsDir,
		},
	}
	baseTool := &agent.Tool{Name: "base_tool"}
	registry, err := initSkills(cfg, []*agent.Tool{baseTool}, eventbus.New(), extReg)
	require.NoError(t, err)
	require.NotNil(t, registry)

	_, ok := registry.GetSkillTool("trusted-skill")
	assert.True(t, ok)
	_, ok = registry.GetSkillTool("rogue-skill")
	assert.False(t, ok)

	toolNames := make(map[string]bool)
	for _, tool := range registry.AllTools() {
		toolNames[tool.Name] = true
	}
	assert.True(t, toolNames["base_tool"])
	assert.True(t, toolNames["skill_trusted-skill"])
	assert.False(t, toolNames["skill_rogue-skill"])
}

func TestInitSkills_DisabledReturnsNilRegistry(t *testing.T) {
	t.Parallel()

	registry, err := initSkills(&config.Config{}, nil, eventbus.New(), nil)
	require.NoError(t, err)
	assert.Nil(t, registry)
}

func TestInitConversationAnalysisAndRetrievalGuards(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	assert.Nil(t, initConversationAnalysis(cfg, nil, testutil.NewMockSessionStore(), nil, nil, eventbus.New()))

	entStore := newWiringKnowledgeEntStore(t)
	kc := &knowledgeComponents{store: knowledge.NewStore(entStore.Client(), testLog())}
	assert.Nil(t, initConversationAnalysis(cfg, nil, testutil.NewMockSessionStore(), kc, nil, eventbus.New()))

	cfg.ObservationalMemory.Enabled = true
	cfg.Agent.Provider = "test-provider"
	cfg.Agent.Model = "test-model"
	cfg.Knowledge.AnalysisTurnThreshold = 2
	cfg.Knowledge.AnalysisTokenThreshold = 20
	require.NotNil(t, initConversationAnalysis(cfg, nil, testutil.NewMockSessionStore(), kc, nil, eventbus.New()))

	initFeedbackProcessor(&config.Config{}, eventbus.New())
	initFeedbackProcessor(&config.Config{Retrieval: config.RetrievalConfig{Feedback: true}}, nil)
	feedbackBus := eventbus.New()
	initFeedbackProcessor(&config.Config{Retrieval: config.RetrievalConfig{Feedback: true}}, feedbackBus)
	feedbackBus.Publish(wiringKnowledgeContextInjectedEvent("wiringKnowledge5-feedback"))

	ctx := context.Background()
	require.NoError(t, kc.store.SaveKnowledge(ctx, "session-1", knowledge.KnowledgeEntry{
		Key:      "wiringKnowledge5-relevance",
		Category: entknowledge.CategoryFact,
		Content:  "retrieval guard branch coverage",
	}))
	initRelevanceAdjuster(&config.Config{}, kc.store, eventbus.New())
	initRelevanceAdjuster(&config.Config{Retrieval: config.RetrievalConfig{AutoAdjust: config.AutoAdjustConfig{Enabled: true}}}, nil, eventbus.New())
	initRelevanceAdjuster(&config.Config{Retrieval: config.RetrievalConfig{AutoAdjust: config.AutoAdjustConfig{Enabled: true}}}, kc.store, nil)
	disabledBus := eventbus.New()
	initRelevanceAdjuster(&config.Config{}, kc.store, disabledBus)
	disabledBus.Publish(wiringKnowledgeContextInjectedEvent("wiringKnowledge5-relevance"))
	assert.Equal(t, 1.0, wiringKnowledgeRelevanceScore(t, entStore, "wiringKnowledge5-relevance"))

	activeBus := eventbus.New()
	initRelevanceAdjuster(&config.Config{
		Retrieval: config.RetrievalConfig{
			AutoAdjust: config.AutoAdjustConfig{
				Enabled:       true,
				Mode:          "active",
				BoostDelta:    0.2,
				DecayDelta:    0.1,
				DecayInterval: 3,
				MinScore:      0.1,
				MaxScore:      4,
				WarmupTurns:   0,
			},
		},
	}, kc.store, activeBus)
	activeBus.Publish(wiringKnowledgeContextInjectedEvent("wiringKnowledge5-relevance"))
	assert.Equal(t, 1.2, wiringKnowledgeRelevanceScore(t, entStore, "wiringKnowledge5-relevance"))
}

func TestProviderTextGeneratorGenerateTextWrapsProxyError(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = ""
	cfg.Agent.Model = ""
	cfg.Providers = nil

	sv, err := supervisor.New(cfg)
	require.NoError(t, err)
	generator := &providerTextGenerator{
		proxy: supervisor.NewProviderProxy(sv, "missing-provider", "test-model"),
	}

	result, err := generator.GenerateText(context.Background(), "system prompt", "user prompt")
	require.Error(t, err)
	assert.Empty(t, result)
	assert.ErrorContains(t, err, "generate text")
	assert.ErrorContains(t, err, "provider not found: missing-provider")
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

func createFTS5ShadowTableConflict(db *sql.DB, tableName string) error {
	_, err := db.Exec(`CREATE TABLE ` + tableName + `_data(id INTEGER PRIMARY KEY, block BLOB)`)
	return err
}

func createIncompatibleKnowledgeFTS5Table(db *sql.DB) error {
	_, err := db.Exec(
		`CREATE VIRTUAL TABLE knowledge_fts USING fts5(title, source_id UNINDEXED, tokenize='unicode61')`,
	)
	return err
}

func createIncompatibleLearningFTS5Table(db *sql.DB) error {
	_, err := db.Exec(
		`CREATE VIRTUAL TABLE learning_fts USING fts5(title, source_id UNINDEXED, tokenize='unicode61')`,
	)
	return err
}

func wiringKnowledgeContextInjectedEvent(key string) eventbus.ContextInjectedEvent {
	return eventbus.ContextInjectedEvent{
		SessionKey: "session-1",
		Query:      "retrieval guard branch coverage",
		Items: []eventbus.ContextInjectedItem{{
			Layer:  knowledge.LayerUserKnowledge.String(),
			Key:    key,
			Source: "like",
		}},
	}
}

func wiringKnowledgeRelevanceScore(t *testing.T, store *session.EntStore, key string) float64 {
	t.Helper()

	row, err := store.Client().Knowledge.Query().
		Where(entknowledge.Key(key), entknowledge.IsLatest(true)).
		Only(context.Background())
	require.NoError(t, err)
	return row.RelevanceScore
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

func writeWiringKnowledgeExtensionPack(t *testing.T, root string) string {
	t.Helper()

	dir := filepath.Join(root, "trusted-pack-source")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "skills", "trusted-skill"), 0o700))

	manifest := `schema: lango.extension/v1
name: trusted-pack
version: 0.1.0
description: Trusted pack for wiring tests
contents:
  skills:
    - name: trusted-skill
      path: skills/trusted-skill/SKILL.md
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "extension.yaml"), []byte(manifest), 0o600))

	content := "---\n" +
		"name: trusted-skill\n" +
		"description: Trusted extension skill\n" +
		"type: instruction\n" +
		"status: active\n" +
		"---\n\n" +
		"Trusted extension-owned skill for wiring tests.\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skills", "trusted-skill", "SKILL.md"), []byte(content), 0o600))
	return dir
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
