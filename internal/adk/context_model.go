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

	"github.com/langoai/lango/internal/embedding"
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
	ragService         *embedding.RAGService
	ragOpts            embedding.RetrieveOptions
	graphRAG           *graph.GraphRAGService
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
	bus                *eventbus.Bus
	logger             *zap.SugaredLogger
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

// WithRAG adds RAG (retrieval-augmented generation) support.
func (m *ContextAwareModelAdapter) WithRAG(svc *embedding.RAGService, opts embedding.RetrieveOptions) *ContextAwareModelAdapter {
	m.ragService = svc
	m.ragOpts = opts
	return m
}

// WithGraphRAG adds graph-enhanced RAG support. When set, graph expansion
// is performed on vector search results to discover structurally connected context.
func (m *ContextAwareModelAdapter) WithGraphRAG(svc *graph.GraphRAGService) *ContextAwareModelAdapter {
	m.graphRAG = svc
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

	userQuery := extractLastUserMessage(req.Contents)

	// ──────────────────────────────────────────────────────────
	// Phase 1: Retrieve all section data in parallel (no truncation).
	//   - Retriever: non-factual layers (RuntimeContext, ToolRegistry, SkillPatterns, PendingInquiries)
	//   - Coordinator: factual layers (UserKnowledge, AgentLearnings, ExternalKnowledge)
	// ──────────────────────────────────────────────────────────
	var knowledgeResult *knowledge.RetrievalResult
	var coordinatorResult *knowledge.RetrievalResult
	var ragResults []embedding.RAGResult
	var graphRAGResult *graph.GraphRAGResult
	var reflections []memory.Reflection
	var observations []memory.Observation
	var runSummaries []RunSummaryContext

	g, gCtx := errgroup.WithContext(ctx)

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
		} else if m.ragService != nil {
			g.Go(func() error {
				ragResults = m.retrieveRAGData(gCtx, userQuery, sessionKey)
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
		RAG:        estimateRAGResultTokens(ragResults, graphRAGResult),
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
		totalMeasured := measured.Knowledge + measured.RAG + measured.Memory + measured.RunSummary
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
	var knowledgeSection, ragSection, memorySection, runSummarySection string

	if knowledgeResult != nil {
		knowledgeResult = knowledge.TruncateResult(knowledgeResult, budgets.Knowledge)
		knowledgeSection = m.retriever.AssemblePrompt("", knowledgeResult)
	}

	if graphRAGResult != nil {
		ragSection = m.formatGraphRAGSection(graphRAGResult, budgets.RAG)
	} else if len(ragResults) > 0 {
		ragSection = formatRAGSection(ragResults, budgets.RAG)
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
	if ragSection != "" {
		prompt = fmt.Sprintf("%s\n\n%s", prompt, ragSection)
	}
	if memorySection != "" {
		prompt = fmt.Sprintf("%s\n\n%s", prompt, memorySection)
	}
	if runSummarySection != "" {
		prompt = fmt.Sprintf("%s\n\n%s", prompt, runSummarySection)
	}

	// Publish context injection event for observability.
	if m.bus != nil {
		knowledgeTokens := types.EstimateTokens(knowledgeSection)
		ragTokens := types.EstimateTokens(ragSection)
		memoryTokens := types.EstimateTokens(memorySection)
		runSummaryTokens := types.EstimateTokens(runSummarySection)
		m.bus.Publish(eventbus.ContextInjectedEvent{
			TurnID:           session.TurnIDFromContext(ctx),
			SessionKey:       sessionKey,
			Query:            userQuery,
			Items:            buildContextInjectedItems(knowledgeResult),
			KnowledgeTokens:  knowledgeTokens,
			RAGTokens:        ragTokens,
			MemoryTokens:     memoryTokens,
			RunSummaryTokens: runSummaryTokens,
			TotalTokens:      knowledgeTokens + ragTokens + memoryTokens + runSummaryTokens,
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

func estimateRAGResultTokens(ragResults []embedding.RAGResult, graphResult *graph.GraphRAGResult) int {
	total := 0
	if len(ragResults) > 0 {
		total += types.EstimateTokens("## Semantic Context (RAG)\n")
		for _, r := range ragResults {
			total += types.EstimateTokens(r.Content) + 30
		}
	}
	if graphResult != nil {
		for _, r := range graphResult.VectorResults {
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
