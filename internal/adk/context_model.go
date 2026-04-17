package adk

import (
	"context"
	"fmt"
	"iter"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/adk/model"
	"google.golang.org/genai"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/prompt"
	"github.com/langoai/lango/internal/retrieval"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
)

// MemoryProvider retrieves observations and reflections for a session.
type MemoryProvider interface {
	ListObservations(ctx context.Context, sessionKey string) ([]memory.Observation, error)
	ListReflections(ctx context.Context, sessionKey string) ([]memory.Reflection, error)
	ListRecentReflections(ctx context.Context, sessionKey string, limit int) ([]memory.Reflection, error)
	ListRecentObservations(ctx context.Context, sessionKey string, limit int) ([]memory.Observation, error)
}

// SessionCompactor abstracts session message compaction for the context model.
// When the measured context exceeds the model window threshold, the adapter
// calls CompactMessages to shrink the message history.
type SessionCompactor interface {
	CompactMessages(key string, upToIndex int, summary string) error
}

// RecallMatch is a single prior-session recall hit.
type RecallMatch struct {
	SessionKey string
	Summary    string
	Rank       float64
}

// RecallProvider surfaces prior-session summaries from the FTS recall index
// at turn start. Implementations should exclude the current session key and
// respect their own rank floor and topN limits.
type RecallProvider interface {
	// RecallRecent queries for matches against the user's input. The current
	// session key is supplied so implementations can exclude it.
	RecallRecent(ctx context.Context, currentSessionKey, query string) ([]RecallMatch, error)
}

// CompactionSyncWaiter waits up to timeout for any in-flight background
// hygiene compaction for the session to complete. Returns true if no
// compaction was in flight or it finished within the bound; false on timeout.
// The returned duration is the actual time spent waiting (useful for slow-path
// telemetry). Implementations should be goroutine-safe.
type CompactionSyncWaiter interface {
	WaitForSession(ctx context.Context, key string, timeout time.Duration) (bool, time.Duration)
}

// RunSummaryProvider retrieves active RunLedger summaries for a session.
type RunSummaryProvider interface {
	ListRunSummaries(ctx context.Context, sessionKey string, limit int) ([]RunSummaryContext, error)
	MaxJournalSeqForSession(ctx context.Context, sessionKey string) (int64, error)
}

// RunSummaryContext is the compact command-context view injected from RunLedger.
type RunSummaryContext struct {
	RunID          string
	Goal           string
	Status         string
	CurrentStep    string
	CurrentBlocker string
}

// ContextAwareModelAdapter wraps a ModelAdapter with context retrieval.
// Before each LLM call, it retrieves relevant knowledge and injects it
// into the system instruction.
type ContextAwareModelAdapter struct {
	inner              *ModelAdapter
	retriever          *knowledge.ContextRetriever
	memoryProvider     MemoryProvider
	graphRAG           *graph.GraphRAGService
	retrievedLimit     int
	coordinator        *retrieval.RetrievalCoordinator
	runtimeAdapter     *RuntimeContextAdapter
	runSummaryProvider RunSummaryProvider
	runSummaryCache    *runSummaryCache
	basePrompt         string
	maxReflections     int
	maxObservations    int
	memoryTokenBudget  int // max tokens for the memory section; 0 = default (4000)
	budgetManager      *ContextBudgetManager
	sessionCompactor   SessionCompactor
	compactionSync     CompactionSyncWaiter
	compactionSyncWait time.Duration
	recallProvider     RecallProvider
	catalogSource      CatalogSource
	modeResolver       ModeResolver
	bus                *eventbus.Bus
	logger             *zap.SugaredLogger
}

// CatalogSource generates the dynamic tool catalog prompt section for a turn.
// An implementation typically holds a *toolcatalog.Catalog and produces a
// prompt-ready string listing visible tools, optionally filtered by the
// session's active mode.
type CatalogSource interface {
	// BuildToolCatalogSection returns the tool catalog prompt section for
	// the given mode name. An empty modeName means "no mode filter".
	BuildToolCatalogSection(modeName string) string
}

// ModeResolver looks up a SessionMode definition by name. It is used to
// resolve the system hint and skill allowlist for the active session.
// Consumers inject a thin adapter around config.LookupMode to avoid an
// import cycle between adk and config.
type ModeResolver interface {
	// LookupModeHint returns the SystemHint for the given mode name, or ""
	// when the mode is unknown or has no hint.
	LookupModeHint(modeName string) string
}

// NewContextAwareModelAdapter creates a context-aware model adapter.
// The builder is used to produce the base system prompt; dynamic context
// (knowledge, memory, RAG) is still appended at call time.
func NewContextAwareModelAdapter(
	inner *ModelAdapter,
	retriever *knowledge.ContextRetriever,
	builder *prompt.Builder,
	logger *zap.SugaredLogger,
) *ContextAwareModelAdapter {
	return &ContextAwareModelAdapter{
		inner:      inner,
		retriever:  retriever,
		basePrompt: builder.Build(),
		logger:     logger,
		runSummaryCache: &runSummaryCache{
			entries: make(map[string]summaryCacheEntry),
		},
	}
}

// WithMemory adds observational memory support to the adapter.
// The session key is resolved at call time from the request context
// via session.SessionKeyFromContext.
func (m *ContextAwareModelAdapter) WithMemory(provider MemoryProvider) *ContextAwareModelAdapter {
	m.memoryProvider = provider
	return m
}

// WithRuntimeAdapter adds runtime context support to the adapter.
func (m *ContextAwareModelAdapter) WithRuntimeAdapter(adapter *RuntimeContextAdapter) *ContextAwareModelAdapter {
	m.runtimeAdapter = adapter
	return m
}

// WithRunSummaryProvider adds RunLedger command-context injection support.
func (m *ContextAwareModelAdapter) WithRunSummaryProvider(provider RunSummaryProvider) *ContextAwareModelAdapter {
	m.runSummaryProvider = provider
	return m
}

// WithGraphRAG adds graph-enhanced retrieved context support. When set, graph
// expansion is performed on phase-1 retrieved results to discover structurally
// connected context.
func (m *ContextAwareModelAdapter) WithGraphRAG(svc *graph.GraphRAGService, limit int) *ContextAwareModelAdapter {
	m.graphRAG = svc
	if limit <= 0 {
		limit = 5
	}
	m.retrievedLimit = limit
	return m
}

// WithCoordinator adds the agentic retrieval coordinator for factual layer retrieval.
// When set, the coordinator runs in Phase 1 alongside the retriever (non-factual layers)
// and their results are merged before context assembly.
func (m *ContextAwareModelAdapter) WithCoordinator(c *retrieval.RetrievalCoordinator) *ContextAwareModelAdapter {
	m.coordinator = c
	return m
}

// WithMemoryLimits sets the maximum number of reflections and observations
// to include in the LLM context. Zero means unlimited (existing behavior).
func (m *ContextAwareModelAdapter) WithMemoryLimits(maxReflections, maxObservations int) *ContextAwareModelAdapter {
	m.maxReflections = maxReflections
	m.maxObservations = maxObservations
	return m
}

// WithMemoryTokenBudget sets the maximum token budget for the memory section
// injected into the system prompt. Reflections are prioritized first (higher
// information density), then observations fill the remaining budget.
// Zero means use default (4000 tokens).
func (m *ContextAwareModelAdapter) WithMemoryTokenBudget(budget int) *ContextAwareModelAdapter {
	m.memoryTokenBudget = budget
	return m
}

// WithBudgetManager sets the context budget manager for per-section token allocation.
// When set, GenerateContent computes per-section budgets and truncates content to fit.
func (m *ContextAwareModelAdapter) WithBudgetManager(bm *ContextBudgetManager) *ContextAwareModelAdapter {
	m.budgetManager = bm
	return m
}

// WithSessionCompactor sets the session message compactor for emergency context
// compaction. When the measured token total exceeds 90% of the model window,
// GenerateContent invokes the compactor to shrink the message history before
// proceeding with the LLM call.
func (m *ContextAwareModelAdapter) WithSessionCompactor(sc SessionCompactor) *ContextAwareModelAdapter {
	m.sessionCompactor = sc
	return m
}

// WithRecallProvider sets the session recall provider. When set,
// GenerateContent queries it at turn start and prepends matching prior-session
// summaries to the RAG section (bounded by the RAG section budget).
func (m *ContextAwareModelAdapter) WithRecallProvider(p RecallProvider) *ContextAwareModelAdapter {
	m.recallProvider = p
	return m
}

// WithCompactionSync sets the background-compaction sync-point waiter and the
// per-turn timeout. When set, GenerateContent waits up to timeout at turn
// start for any in-flight compaction for the session to finish. On timeout
// the turn proceeds with the current state and emits a CompactionSlowEvent
// on the event bus. A nil waiter or zero timeout disables the sync point.
func (m *ContextAwareModelAdapter) WithCompactionSync(w CompactionSyncWaiter, timeout time.Duration) *ContextAwareModelAdapter {
	m.compactionSync = w
	m.compactionSyncWait = timeout
	return m
}

// WithCatalog sets the tool catalog source used to generate the per-turn tool
// catalog prompt section. When set, the tool catalog section is generated
// dynamically in GenerateContent() from the current session mode, rather than
// being baked into basePrompt at boot time.
func (m *ContextAwareModelAdapter) WithCatalog(cs CatalogSource) *ContextAwareModelAdapter {
	m.catalogSource = cs
	return m
}

// WithModeResolver sets the mode resolver used to fetch the SystemHint for
// the active session mode. A nil resolver disables mode hint injection.
func (m *ContextAwareModelAdapter) WithModeResolver(r ModeResolver) *ContextAwareModelAdapter {
	m.modeResolver = r
	return m
}

// WithEventBus sets the event bus for context injection observability.
// When set, GenerateContent publishes a ContextInjectedEvent after context assembly.
func (m *ContextAwareModelAdapter) WithEventBus(bus *eventbus.Bus) *ContextAwareModelAdapter {
	m.bus = bus
	return m
}

// Name delegates to the inner adapter.
func (m *ContextAwareModelAdapter) Name() string {
	return m.inner.Name()
}

// GenerateContent retrieves context and injects an augmented system prompt before delegating to the inner adapter.
func (m *ContextAwareModelAdapter) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	prompt := m.basePrompt

	// Resolve session key from request context (set by gateway/channels).
	sessionKey := session.SessionKeyFromContext(ctx)

	// Update runtime session state before retrieval
	if m.runtimeAdapter != nil && sessionKey != "" {
		m.runtimeAdapter.SetSession(sessionKey)
	}

	// Sync point: wait for any in-flight background hygiene compaction to
	// settle before assembling the new turn. Bounded by compactionSyncWait;
	// on timeout we proceed with the current state and emit a slow event.
	if m.compactionSync != nil && m.compactionSyncWait > 0 && sessionKey != "" {
		done, waited := m.compactionSync.WaitForSession(ctx, sessionKey, m.compactionSyncWait)
		if !done {
			m.logger.Warnw("compaction sync-point timeout; proceeding with current context",
				"sessionKey", sessionKey,
				"waited", waited,
			)
			if m.bus != nil {
				m.bus.Publish(eventbus.CompactionSlowEvent{
					SessionKey: sessionKey,
					WaitedFor:  waited,
					Timestamp:  time.Now(),
				})
			}
		}
	}

	userQuery := extractLastUserMessage(req.Contents)

	// ──────────────────────────────────────────────────────────
	// Phase 1: Retrieve all section data in parallel (no truncation).
	//   - Retriever: non-factual layers (RuntimeContext, ToolRegistry, SkillPatterns, PendingInquiries)
	//   - Coordinator: factual layers (UserKnowledge, AgentLearnings, ExternalKnowledge)
	// ──────────────────────────────────────────────────────────
	var knowledgeResult *knowledge.RetrievalResult
	var coordinatorResult *knowledge.RetrievalResult
	var graphRAGResult *graph.GraphRAGResult
	var reflections []memory.Reflection
	var observations []memory.Observation
	var runSummaries []RunSummaryContext
	var recallMatches []RecallMatch

	g, gCtx := errgroup.WithContext(ctx)

	if userQuery != "" && m.recallProvider != nil {
		g.Go(func() error {
			matches, err := m.recallProvider.RecallRecent(gCtx, sessionKey, userQuery)
			if err != nil {
				m.logger.Warnw("session recall error", "error", err)
				return nil
			}
			recallMatches = matches
			return nil
		})
	}

	if userQuery != "" && m.retriever != nil {
		g.Go(func() error {
			layers := []knowledge.ContextLayer{
				knowledge.LayerRuntimeContext,
				knowledge.LayerToolRegistry,
				knowledge.LayerSkillPatterns,
				knowledge.LayerPendingInquiries,
			}
			retrieved, err := m.retriever.Retrieve(gCtx, knowledge.RetrievalRequest{
				Query:  userQuery,
				Layers: layers,
			})
			if err != nil {
				m.logger.Warnw("context retrieval error", "error", err)
			} else if retrieved != nil && retrieved.TotalItems > 0 {
				knowledgeResult = retrieved
			}
			return nil
		})
	}

	if userQuery != "" && m.coordinator != nil {
		g.Go(func() error {
			findings, err := m.coordinator.Retrieve(gCtx, userQuery, 0)
			if err != nil {
				m.logger.Warnw("coordinator retrieval error", "error", err)
				return nil
			}
			coordinatorResult = retrieval.ToRetrievalResult(findings)
			return nil
		})
	}

	if userQuery != "" {
		if m.graphRAG != nil {
			g.Go(func() error {
				graphRAGResult = m.retrieveGraphRAGData(gCtx, userQuery, sessionKey)
				return nil
			})
		}
	}

	if m.memoryProvider != nil && sessionKey != "" {
		g.Go(func() error {
			reflections, observations = m.retrieveMemoryData(gCtx, sessionKey)
			return nil
		})
	}

	if m.runSummaryProvider != nil && sessionKey != "" {
		g.Go(func() error {
			runSummaries = m.retrieveRunSummaryData(gCtx, sessionKey)
			return nil
		})
	}

	_ = g.Wait()

	// Merge retriever (non-factual) and coordinator (factual) results.
	knowledgeResult = mergeRetrievalResults(knowledgeResult, coordinatorResult)

	// ──────────────────────────────────────────────────────────
	// Phase 2: Measure actual content → Reallocate budgets.
	// ──────────────────────────────────────────────────────────
	measured := SectionTokens{
		Knowledge:  estimateKnowledgeTokens(knowledgeResult),
		RAG:        estimateRetrievedResultTokens(graphRAGResult),
		Memory:     estimateMemoryTokens(reflections, observations),
		RunSummary: estimateRunSummaryTokens(runSummaries),
	}

	var budgets SectionBudgets
	if m.budgetManager != nil {
		budgets = m.budgetManager.ReallocateBudgets(measured)
		if budgets.Degraded {
			m.logger.Warnw("context budget degraded to unlimited — model window too small for base prompt",
				"modelWindow", m.budgetManager.ModelWindow())
		}
	}

	// ──────────────────────────────────────────────────────────
	// Phase 2.5: Emergency compaction if context nears model window.
	// Trigger: measured total > modelWindow × 0.9, compactor available.
	// NOT triggered by budgets.Degraded (that's a config issue).
	// ──────────────────────────────────────────────────────────
	if m.sessionCompactor != nil && m.budgetManager != nil && !budgets.Degraded && sessionKey != "" {
		// Include conversation history and base prompt in the measurement
		// so long chats also trigger compaction (not just injected context).
		var historyTokens int
		for _, c := range req.Contents {
			for _, p := range c.Parts {
				if p.Text != "" {
					historyTokens += types.EstimateTokens(p.Text)
				}
			}
		}
		baseTokens := types.EstimateTokens(m.basePrompt)
		totalMeasured := measured.Knowledge + measured.RAG + measured.Memory + measured.RunSummary + historyTokens + baseTokens
		threshold := int(float64(m.budgetManager.ModelWindow()) * 0.9)
		if totalMeasured > threshold {
			m.logger.Warnw("emergency context compaction triggered",
				"measured", totalMeasured,
				"threshold", threshold,
				"sessionKey", sessionKey,
			)
			if err := m.sessionCompactor.CompactMessages(sessionKey, -1, ""); err != nil {
				m.logger.Errorw("emergency compaction failed", "error", err)
			}
			// Note: compaction ran at most once per GenerateContent call.
			// We do NOT restart Phase 1 here because the LLM request's
			// Contents already carry the session history from ADK.
			// Compaction shortens future turns, not the current one.
		}
	}

	// ──────────────────────────────────────────────────────────
	// Phase 3: Truncate + Format each section with reallocated budgets.
	// ──────────────────────────────────────────────────────────
	var knowledgeSection, retrievedSection, memorySection, runSummarySection string

	if knowledgeResult != nil {
		knowledgeResult = knowledge.TruncateResult(knowledgeResult, budgets.Knowledge)
		knowledgeSection = m.retriever.AssemblePrompt("", knowledgeResult)
	}

	// Split retrieved-context budget between graph-expanded results and session
	// recall when both are present. Recall gets 1/3, retrieved context gets 2/3. When only one source
	// exists it gets the full budget.
	retrievedBudget := budgets.RAG
	recallBudget := budgets.RAG
	hasRetrieved := graphRAGResult != nil
	if len(recallMatches) > 0 && hasRetrieved {
		recallBudget = budgets.RAG / 3
		retrievedBudget = budgets.RAG - recallBudget
	}

	if graphRAGResult != nil {
		retrievedSection = m.formatGraphRAGSection(graphRAGResult, retrievedBudget)
	}

	// Prepend session recall matches to the retrieved context section. Each source is
	// independently budget-capped so their combined size stays within the
	// total retrieved context allocation.
	if len(recallMatches) > 0 {
		recallSection := formatRecallSection(recallMatches, recallBudget)
		if recallSection != "" {
			if retrievedSection == "" {
				retrievedSection = recallSection
			} else {
				retrievedSection = recallSection + "\n\n" + retrievedSection
			}
		}
	}

	if len(reflections) > 0 || len(observations) > 0 {
		memorySection = m.formatMemorySection(reflections, observations, budgets.Memory)
	}

	if len(runSummaries) > 0 {
		runSummarySection = formatRunSummarySection(runSummaries, budgets.RunSummary)
	}

	// Combine sections
	if knowledgeSection != "" {
		prompt = fmt.Sprintf("%s\n\n%s", prompt, knowledgeSection)
	}
	if retrievedSection != "" {
		prompt = fmt.Sprintf("%s\n\n%s", prompt, retrievedSection)
	}
	if memorySection != "" {
		prompt = fmt.Sprintf("%s\n\n%s", prompt, memorySection)
	}
	if runSummarySection != "" {
		prompt = fmt.Sprintf("%s\n\n%s", prompt, runSummarySection)
	}

	// Dynamic tool catalog section — generated per turn, mode-aware.
	modeName := session.ModeNameFromContext(ctx)
	if m.catalogSource != nil {
		if toolCatalogSection := m.catalogSource.BuildToolCatalogSection(modeName); toolCatalogSection != "" {
			prompt = fmt.Sprintf("%s\n\n%s", prompt, toolCatalogSection)
		}
	}

	// Mode system hint injection.
	if modeName != "" && m.modeResolver != nil {
		if hint := m.modeResolver.LookupModeHint(modeName); hint != "" {
			prompt = fmt.Sprintf("%s\n\n## Session Mode: %s\n\n%s", prompt, modeName, hint)
		}
	}

	// Publish context injection event for observability.
	if m.bus != nil {
		knowledgeTokens := types.EstimateTokens(knowledgeSection)
		retrievedTokens := types.EstimateTokens(retrievedSection)
		memoryTokens := types.EstimateTokens(memorySection)
		runSummaryTokens := types.EstimateTokens(runSummarySection)
		m.bus.Publish(eventbus.ContextInjectedEvent{
			TurnID:           session.TurnIDFromContext(ctx),
			SessionKey:       sessionKey,
			Query:            userQuery,
			Items:            buildContextInjectedItems(knowledgeResult),
			KnowledgeTokens:  knowledgeTokens,
			RetrievedTokens:  retrievedTokens,
			MemoryTokens:     memoryTokens,
			RunSummaryTokens: runSummaryTokens,
			TotalTokens:      knowledgeTokens + retrievedTokens + memoryTokens + runSummaryTokens,
			Timestamp:        time.Now(),
		})
	}

	// Set the augmented system instruction
	if prompt != m.basePrompt {
		if req.Config == nil {
			req.Config = &genai.GenerateContentConfig{}
		}
		req.Config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{{Text: prompt}},
		}
	}

	return m.inner.GenerateContent(ctx, req, stream)
}

// buildContextInjectedItems converts knowledge RetrievalResult items into event items.
// Only knowledge items are included — RAG/memory/runSummary are opaque sections.
func buildContextInjectedItems(retrieved *knowledge.RetrievalResult) []eventbus.ContextInjectedItem {
	if retrieved == nil {
		return nil
	}
	var items []eventbus.ContextInjectedItem
	for layer, ctxItems := range retrieved.Items {
		for _, item := range ctxItems {
			items = append(items, eventbus.ContextInjectedItem{
				Layer:         layer.String(),
				Key:           item.Key,
				Score:         item.Score,
				Source:        item.Source,
				Category:      item.Category,
				TokenEstimate: types.EstimateTokens(item.Content),
			})
		}
	}
	return items
}

// ──────────────────────────────────────────────────────────
// Token estimation helpers for pre-assembly measurement.
// ──────────────────────────────────────────────────────────

func estimateKnowledgeTokens(result *knowledge.RetrievalResult) int {
	if result == nil {
		return 0
	}
	total := 0
	for _, items := range result.Items {
		for _, item := range items {
			total += types.EstimateTokens(item.Content) + 20 // header overhead per item
		}
	}
	return total
}

func estimateRetrievedResultTokens(graphResult *graph.GraphRAGResult) int {
	total := 0
	if graphResult != nil {
		total += types.EstimateTokens("## Retrieved Context\n")
		for _, r := range graphResult.ContentResults {
			total += types.EstimateTokens(r.Content) + 30
		}
		for range graphResult.GraphResults {
			total += 40 // approximate per graph node
		}
	}
	return total
}

func estimateMemoryTokens(reflections []memory.Reflection, observations []memory.Observation) int {
	total := 0
	for _, r := range reflections {
		total += types.EstimateTokens(r.Content) + 5
	}
	for _, o := range observations {
		total += types.EstimateTokens(o.Content) + 5
	}
	return total
}

func estimateRunSummaryTokens(summaries []RunSummaryContext) int {
	total := 0
	for _, s := range summaries {
		total += types.EstimateTokens(s.Goal) + types.EstimateTokens(s.RunID) + 20
	}
	return total
}

// mergeRetrievalResults combines two RetrievalResults by merging their Items maps.
// The two results should cover disjoint layer sets (no key conflict expected).
func mergeRetrievalResults(a, b *knowledge.RetrievalResult) *knowledge.RetrievalResult {
	if a == nil && b == nil {
		return nil
	}
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	merged := &knowledge.RetrievalResult{
		Items: make(map[knowledge.ContextLayer][]knowledge.ContextItem, len(a.Items)+len(b.Items)),
	}
	for layer, items := range a.Items {
		merged.Items[layer] = items
		merged.TotalItems += len(items)
	}
	for layer, items := range b.Items {
		merged.Items[layer] = items
		merged.TotalItems += len(items)
	}
	return merged
}
