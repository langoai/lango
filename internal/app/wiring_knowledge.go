package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/learning"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/provider"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/skill"
	"github.com/langoai/lango/internal/supervisor"
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
func initKnowledge(cfg *config.Config, store session.Store, gc *graphComponents) *knowledgeComponents {
	if !cfg.Knowledge.Enabled {
		logger().Info("knowledge system disabled")
		return nil
	}

	entStore, ok := store.(*session.EntStore)
	if !ok {
		logger().Warn("knowledge system requires EntStore, skipping")
		return nil
	}

	client := entStore.Client()
	kLogger := logger()

	kStore := knowledge.NewStore(client, kLogger)

	engine := learning.NewEngine(kStore, kLogger)

	// Select observer: GraphEngine when graph store is available, otherwise base Engine.
	var observer learning.ToolResultObserver = engine
	if gc != nil {
		graphEngine := learning.NewGraphEngine(kStore, gc.store, kLogger)
		graphEngine.SetGraphCallback(func(triples []graph.Triple) {
			gc.buffer.Enqueue(graph.GraphRequest{Triples: triples})
		})
		observer = graphEngine
		logger().Info("graph-enhanced learning engine initialized")
	}

	logger().Info("knowledge system initialized")
	return &knowledgeComponents{
		store:    kStore,
		engine:   engine,
		observer: observer,
	}
}

// initSkills creates the file-based skill registry.
func initSkills(cfg *config.Config, baseTools []*agent.Tool) *skill.Registry {
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

	// Deploy embedded default skills.
	defaultFS, err := skills.DefaultFS()
	if err == nil {
		if err := store.EnsureDefaults(defaultFS); err != nil {
			sLogger.Warnw("deploy default skills error", "error", err)
		}
	}

	registry := skill.NewRegistry(store, baseTools, sLogger)
	ctx := context.Background()
	if err := registry.LoadSkills(ctx); err != nil {
		sLogger.Warnw("load skills error", "error", err)
	}

	sLogger.Infow("skill system initialized", "dir", dir)
	return registry
}

// initConversationAnalysis creates the conversation analysis pipeline if both
// knowledge and observational memory are enabled.
func initConversationAnalysis(cfg *config.Config, sv *supervisor.Supervisor, store session.Store, kc *knowledgeComponents, gc *graphComponents) *learning.AnalysisBuffer {
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
	learner := learning.NewSessionLearner(generator, kc.store, aLogger)

	// Wire graph callbacks if graph store is available.
	if gc != nil && gc.buffer != nil {
		graphCB := func(triples []graph.Triple) {
			gc.buffer.Enqueue(graph.GraphRequest{Triples: triples})
		}
		analyzer.SetGraphCallback(graphCB)
		learner.SetGraphCallback(graphCB)
	}

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

// providerTextGenerator adapts a supervisor.ProviderProxy to the memory.TextGenerator interface.
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
