package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/extension"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/learning"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/provider"
	"github.com/langoai/lango/internal/retrieval"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/search"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/skill"
	"github.com/langoai/lango/internal/storagebroker"
	"github.com/langoai/lango/internal/supervisor"
	"github.com/langoai/lango/internal/types"
	"github.com/langoai/lango/skills"
)

// knowledgeComponents holds optional self-learning components.
type knowledgeComponents struct {
	store    *knowledge.Store
	engine   *learning.Engine
	observer learning.ToolResultObserver
}

// initKnowledge creates the self-learning components if enabled.
// When gc is provided, a GraphEngine is used as the observer instead of the base Engine.
func initKnowledge(cfg *config.Config, store session.Store, gc *graphComponents, bus *eventbus.Bus, broker storagebroker.API) (*knowledgeComponents, *types.FeatureStatus) {
	const featureName = "Knowledge"

	if !cfg.Knowledge.Enabled {
		logger().Info("knowledge system disabled")
		return nil, &types.FeatureStatus{Name: featureName, Enabled: false, Healthy: true}
	}

	entStore, ok := store.(*session.EntStore)
	if !ok {
		logger().Warn("knowledge system requires EntStore, skipping")
		return nil, &types.FeatureStatus{
			Name: featureName, Enabled: false, Healthy: false,
			Reason:     "requires EntStore backend",
			Suggestion: "configure session.databasePath with an Ent-backed store",
		}
	}

	client := entStore.Client()
	kLogger := logger()

	kStore := knowledge.NewStore(client, kLogger)
	kStore.SetEventBus(bus)
	if broker != nil {
		kStore.SetPayloadProtector(storagebroker.NewPayloadProtector(broker))
	}

	engine := learning.NewEngine(kStore, kLogger)

	// Select observer: GraphEngine when graph store is available, otherwise base Engine.
	var observer learning.ToolResultObserver = engine
	if gc != nil {
		graphEngine := learning.NewGraphEngine(kStore, gc.store, kLogger)
		graphEngine.SetEventBus(bus)
		observer = graphEngine
		logger().Info("graph-enhanced learning engine initialized")
	}

	logger().Info("knowledge system initialized")
	return &knowledgeComponents{
		store:    kStore,
		engine:   engine,
		observer: observer,
	}, &types.FeatureStatus{Name: featureName, Enabled: true, Healthy: true}
}

// initFTS5 probes for FTS5 support, creates FTS5 indexes for knowledge and learning,
// bulk-indexes existing entries, and injects indexes into the knowledge store.
// Returns true if FTS5 is available and initialized.
func initFTS5(ctx context.Context, rawDB *sql.DB, kStore *knowledge.Store) bool {
	if rawDB == nil {
		return false
	}
	if !search.ProbeFTS5(rawDB) {
		logger().Info("FTS5 unavailable, using LIKE search fallback")
		return false
	}

	// Create knowledge FTS5 index (columns: key, content).
	knowledgeIdx := search.NewFTS5Index(rawDB, "knowledge_fts", []string{"key", "content"})
	if err := knowledgeIdx.EnsureTable(); err != nil {
		logger().Warnw("FTS5 knowledge table creation failed", "error", err)
		return false
	}

	// Create learning FTS5 index (columns: trigger, error_pattern, fix).
	learningIdx := search.NewFTS5Index(rawDB, "learning_fts", []string{"trigger", "error_pattern", "fix"})
	if err := learningIdx.EnsureTable(); err != nil {
		logger().Warnw("FTS5 learning table creation failed", "error", err)
		return false
	}

	// Bulk-index existing entries.
	if err := bulkIndexKnowledge(ctx, rawDB, knowledgeIdx); err != nil {
		logger().Warnw("FTS5 knowledge bulk index failed", "error", err)
	}
	if err := bulkIndexLearnings(ctx, rawDB, learningIdx); err != nil {
		logger().Warnw("FTS5 learning bulk index failed", "error", err)
	}

	// Inject into knowledge store.
	kStore.SetFTS5Index(knowledgeIdx)
	kStore.SetLearningFTS5Index(learningIdx)

	logger().Info("FTS5 search initialized")
	return true
}

// bulkIndexKnowledge clears and re-indexes all knowledge entries into FTS5.
func bulkIndexKnowledge(ctx context.Context, db *sql.DB, idx *search.FTS5Index) error {
	// Clear existing index data for idempotent re-index.
	if _, err := db.ExecContext(ctx, `DELETE FROM knowledge_fts`); err != nil {
		return fmt.Errorf("clear knowledge FTS5: %w", err)
	}

	rows, err := db.QueryContext(ctx, `SELECT "key", content FROM knowledges WHERE is_latest = 1`)
	if err != nil {
		return fmt.Errorf("query knowledge for FTS5 index: %w", err)
	}
	defer rows.Close()

	var records []search.Record
	for rows.Next() {
		var key, content string
		if err := rows.Scan(&key, &content); err != nil {
			return fmt.Errorf("scan knowledge row: %w", err)
		}
		records = append(records, search.Record{RowID: key, Values: []string{key, content}})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate knowledge rows: %w", err)
	}

	if len(records) > 0 {
		if err := idx.BulkInsert(ctx, records); err != nil {
			return fmt.Errorf("bulk insert knowledge FTS5: %w", err)
		}
		logger().Infow("FTS5 knowledge index populated", "count", len(records))
	}
	return nil
}

// bulkIndexLearnings clears and re-indexes all learning entries into FTS5.
func bulkIndexLearnings(ctx context.Context, db *sql.DB, idx *search.FTS5Index) error {
	if _, err := db.ExecContext(ctx, `DELETE FROM learning_fts`); err != nil {
		return fmt.Errorf("clear learning FTS5: %w", err)
	}

	rows, err := db.QueryContext(ctx, `SELECT id, trigger, COALESCE(error_pattern, ''), COALESCE(fix, '') FROM learnings`)
	if err != nil {
		return fmt.Errorf("query learnings for FTS5 index: %w", err)
	}
	defer rows.Close()

	var records []search.Record
	for rows.Next() {
		var id, trigger, errorPattern, fix string
		if err := rows.Scan(&id, &trigger, &errorPattern, &fix); err != nil {
			return fmt.Errorf("scan learning row: %w", err)
		}
		records = append(records, search.Record{RowID: id, Values: []string{trigger, errorPattern, fix}})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate learning rows: %w", err)
	}

	if len(records) > 0 {
		if err := idx.BulkInsert(ctx, records); err != nil {
			return fmt.Errorf("bulk insert learning FTS5: %w", err)
		}
		logger().Infow("FTS5 learning index populated", "count", len(records))
	}
	return nil
}

// initSkills creates the file-based skill registry.
func initSkills(cfg *config.Config, baseTools []*agent.Tool, bus *eventbus.Bus, extReg *extension.Registry) *skill.Registry {
	if !cfg.Skill.Enabled {
		logger().Info("skill system disabled")
		return nil
	}

	dir := cfg.Skill.SkillsDir
	if dir == "" {
		dir = "~/.lango/skills"
	}
	// Expand ~ to home directory.
	if len(dir) > 1 && dir[:2] == "~/" {
		if home, err := os.UserHomeDir(); err == nil {
			dir = filepath.Join(home, dir[2:])
		}
	}

	sLogger := logger()
	store := skill.NewFileSkillStore(dir, sLogger)

	// Restrict ext-<pack>/ skill loading to healthy packs from the extension
	// registry. When extensions are disabled or no registry exists,
	// AllowedExtPacks stays nil → all ext-packs are skipped (safe default).
	if extReg != nil {
		allowed := make(map[string]bool)
		for _, p := range extReg.OKPacks() {
			if p.Manifest != nil {
				allowed[p.Manifest.Name] = true
			}
		}
		store.AllowedExtPacks = allowed
	}

	// Deploy embedded default skills.
	defaultFS, err := skills.DefaultFS()
	if err == nil {
		if err := store.EnsureDefaults(defaultFS); err != nil {
			sLogger.Warnw("deploy default skills error", "error", err)
		}
	}

	registry := skill.NewRegistry(store, baseTools, sLogger)

	// Inject OS-level sandbox if enabled.
	if iso := initOSSandbox(cfg); iso != nil {
		if iso.Available() {
			workDir := cfg.Sandbox.WorkspacePath
			if workDir == "" {
				workDir, _ = os.Getwd()
			}
			registry.SetOSIsolator(iso, workDir, cfg.DataRoot)
			registry.SetProtectedPaths(resolvedProtectedPaths(cfg, nil))
		}
		registry.SetFailClosed(cfg.Sandbox.FailClosed)
	}
	if bus != nil {
		registry.SetEventBus(bus)
	}

	ctx := context.Background()
	if err := registry.LoadSkills(ctx); err != nil {
		sLogger.Warnw("load skills error", "error", err)
	}

	sLogger.Infow("skill system initialized", "dir", dir)
	return registry
}

// initConversationAnalysis creates the conversation analysis pipeline if both
// knowledge and observational memory are enabled.
func initConversationAnalysis(cfg *config.Config, sv *supervisor.Supervisor, store session.Store, kc *knowledgeComponents, gc *graphComponents, bus *eventbus.Bus) *learning.AnalysisBuffer {
	if kc == nil {
		return nil
	}
	if !cfg.ObservationalMemory.Enabled {
		return nil
	}

	// Create LLM proxy reusing the observational memory provider/model.
	omProvider := cfg.ObservationalMemory.Provider
	if omProvider == "" {
		omProvider = cfg.Agent.Provider
	}
	omModel := cfg.ObservationalMemory.Model
	if omModel == "" {
		omModel = cfg.Agent.Model
	}

	proxy := supervisor.NewProviderProxy(sv, omProvider, omModel)
	generator := &providerTextGenerator{proxy: proxy}

	aLogger := logger()

	analyzer := learning.NewConversationAnalyzer(generator, kc.store, aLogger)
	analyzer.SetEventBus(bus)
	learner := learning.NewSessionLearner(generator, kc.store, aLogger)
	learner.SetEventBus(bus)

	// Message provider.
	getMessages := func(sessionKey string) ([]session.Message, error) {
		sess, err := store.Get(sessionKey)
		if err != nil {
			return nil, err
		}
		if sess == nil {
			return nil, nil
		}
		return sess.History, nil
	}

	turnThreshold := cfg.Knowledge.AnalysisTurnThreshold
	tokenThreshold := cfg.Knowledge.AnalysisTokenThreshold

	buf := learning.NewAnalysisBuffer(analyzer, learner, getMessages, turnThreshold, tokenThreshold, aLogger)

	logger().Infow("conversation analysis initialized",
		"turnThreshold", turnThreshold,
		"tokenThreshold", tokenThreshold,
	)

	return buf
}

// providerTextGenerator adapts a supervisor.ProviderProxy to the llm.TextGenerator interface.
type providerTextGenerator struct {
	proxy *supervisor.ProviderProxy
}

func (g *providerTextGenerator) GenerateText(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	params := provider.GenerateParams{
		Messages: []provider.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	stream, err := g.proxy.Generate(ctx, params)
	if err != nil {
		return "", fmt.Errorf("generate text: %w", err)
	}

	var result strings.Builder
	for evt, err := range stream {
		if err != nil {
			return "", fmt.Errorf("stream text: %w", err)
		}
		if evt.Type == provider.StreamEventPlainText {
			result.WriteString(evt.Text)
		}
		if evt.Type == provider.StreamEventError && evt.Error != nil {
			return "", evt.Error
		}
	}
	return result.String(), nil
}

// inquiryProviderAdapter bridges librarian.InquiryStore → knowledge.InquiryProvider.
type inquiryProviderAdapter struct {
	store *librarian.InquiryStore
}

func (a *inquiryProviderAdapter) PendingInquiryItems(ctx context.Context, sessionKey string, limit int) ([]knowledge.ContextItem, error) {
	inquiries, err := a.store.ListPendingInquiries(ctx, sessionKey, limit)
	if err != nil {
		return nil, err
	}

	items := make([]knowledge.ContextItem, 0, len(inquiries))
	for _, inq := range inquiries {
		items = append(items, knowledge.ContextItem{
			Layer:   knowledge.LayerPendingInquiries,
			Key:     inq.Topic,
			Content: inq.Question,
			Source:  inq.Context,
		})
	}
	return items, nil
}

// skillProviderAdapter adapts *skill.Registry to knowledge.SkillProvider.
type skillProviderAdapter struct {
	registry *skill.Registry
}

func (a *skillProviderAdapter) ListActiveSkillInfos(ctx context.Context) ([]knowledge.SkillInfo, error) {
	entries, err := a.registry.ListActiveSkills(ctx)
	if err != nil {
		return nil, err
	}
	infos := make([]knowledge.SkillInfo, len(entries))
	for i, e := range entries {
		infos[i] = knowledge.SkillInfo{
			Name:        e.Name,
			Description: e.Description,
			Type:        string(e.Type),
		}
	}
	return infos, nil
}

// runSummaryProviderAdapter adapts RunLedger summaries for ADK command-context injection.
type runSummaryProviderAdapter struct {
	store runledger.RunLedgerStore
}

func (a *runSummaryProviderAdapter) ListRunSummaries(
	ctx context.Context,
	sessionKey string,
	limit int,
) ([]adk.RunSummaryContext, error) {
	summaries, err := a.store.ListRunSummariesBySession(ctx, sessionKey, limit)
	if err != nil {
		return nil, err
	}
	result := make([]adk.RunSummaryContext, 0, len(summaries))
	for _, summary := range summaries {
		if summary.Status != runledger.RunStatusRunning && summary.Status != runledger.RunStatusPaused {
			continue
		}
		result = append(result, adk.RunSummaryContext{
			RunID:          summary.RunID,
			Goal:           summary.Goal,
			Status:         string(summary.Status),
			CurrentStep:    summary.CurrentStepGoal,
			CurrentBlocker: summary.CurrentBlocker,
		})
	}
	return result, nil
}

func (a *runSummaryProviderAdapter) MaxJournalSeqForSession(
	ctx context.Context,
	sessionKey string,
) (int64, error) {
	return a.store.MaxJournalSeqForSession(ctx, sessionKey)
}

// initRetrievalCoordinator creates the agentic retrieval coordinator if enabled.
func initRetrievalCoordinator(cfg *config.Config, kStore *knowledge.Store) *retrieval.RetrievalCoordinator {
	if !cfg.Retrieval.Enabled {
		return nil
	}

	agents := []retrieval.RetrievalAgent{
		retrieval.NewFactSearchAgent(kStore),
		retrieval.NewTemporalSearchAgent(kStore),
	}

	coordinator := retrieval.NewRetrievalCoordinator(agents, logger())

	logger().Infow("retrieval coordinator initialized", "agents", len(agents))
	return coordinator
}

// initFeedbackProcessor creates and subscribes the context injection feedback
// processor if enabled. This operates independently of the knowledge system
// and retrieval coordinator — it observes all GenerateContent context injection.
func initFeedbackProcessor(cfg *config.Config, bus *eventbus.Bus) {
	if !cfg.Retrieval.Feedback {
		return
	}
	if bus == nil {
		return
	}

	fp := retrieval.NewFeedbackProcessor(logger())
	fp.Subscribe(bus)

	logger().Info("retrieval feedback processor initialized")
}

// initRelevanceAdjuster creates and subscribes the relevance score adjuster if enabled.
func initRelevanceAdjuster(cfg *config.Config, kStore *knowledge.Store, bus *eventbus.Bus) {
	if !cfg.Retrieval.AutoAdjust.Enabled {
		return
	}
	if bus == nil || kStore == nil {
		return
	}

	adjCfg := retrieval.RelevanceAdjusterConfig{
		Mode:          cfg.Retrieval.AutoAdjust.Mode,
		BoostDelta:    cfg.Retrieval.AutoAdjust.BoostDelta,
		DecayDelta:    cfg.Retrieval.AutoAdjust.DecayDelta,
		DecayInterval: cfg.Retrieval.AutoAdjust.DecayInterval,
		MinScore:      cfg.Retrieval.AutoAdjust.MinScore,
		MaxScore:      cfg.Retrieval.AutoAdjust.MaxScore,
		WarmupTurns:   cfg.Retrieval.AutoAdjust.WarmupTurns,
	}

	adj := retrieval.NewRelevanceAdjuster(kStore, adjCfg, logger())
	adj.Subscribe(bus)

	logger().Infow("relevance adjuster initialized", "mode", adjCfg.Mode, "warmupTurns", adjCfg.WarmupTurns)
}
