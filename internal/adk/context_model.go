package adk

import (
	"context"
	"fmt"
	"iter"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/adk/model"
	"google.golang.org/genai"

	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/prompt"
	"github.com/langoai/lango/internal/retrieval"
	"github.com/langoai/lango/internal/session"
)

// MemoryProvider retrieves observations and reflections for a session.
type MemoryProvider interface {
	ListObservations(ctx context.Context, sessionKey string) ([]memory.Observation, error)
	ListReflections(ctx context.Context, sessionKey string) ([]memory.Reflection, error)
	ListRecentReflections(ctx context.Context, sessionKey string, limit int) ([]memory.Reflection, error)
	ListRecentObservations(ctx context.Context, sessionKey string, limit int) ([]memory.Observation, error)
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

// WithCoordinator adds the agentic retrieval coordinator. When set in shadow mode,
// the coordinator runs in parallel with the existing retrieval path and logs comparison metrics.
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

	// Compute per-section budgets (0 = unlimited when no budget manager).
	var budgets SectionBudgets
	if m.budgetManager != nil {
		budgets = m.budgetManager.SectionBudgets()
		if budgets.Degraded {
			m.logger.Warnw("context budget degraded to unlimited — model window too small for base prompt",
				"modelWindow", m.budgetManager.ModelWindow())
		}
	}

	var knowledgeSection, ragSection, memorySection, runSummarySection string
	var oldRetrieved *knowledge.RetrievalResult // captured for shadow comparison

	g, gCtx := errgroup.WithContext(ctx)

	// Knowledge retrieval
	if userQuery != "" && m.retriever != nil {
		g.Go(func() error {
			layers := []knowledge.ContextLayer{
				knowledge.LayerRuntimeContext,
				knowledge.LayerToolRegistry,
				knowledge.LayerUserKnowledge,
				knowledge.LayerSkillPatterns,
				knowledge.LayerExternalKnowledge,
				knowledge.LayerAgentLearnings,
			}
			retrieved, err := m.retriever.Retrieve(gCtx, knowledge.RetrievalRequest{
				Query:  userQuery,
				Layers: layers,
			})
			if err != nil {
				m.logger.Warnw("context retrieval error", "error", err)
			} else if retrieved != nil && retrieved.TotalItems > 0 {
				retrieved = knowledge.TruncateResult(retrieved, budgets.Knowledge)
				knowledgeSection = m.retriever.AssemblePrompt("", retrieved)
				oldRetrieved = retrieved
			}
			return nil
		})
	}

	// RAG/GraphRAG retrieval
	if userQuery != "" {
		if m.graphRAG != nil {
			g.Go(func() error {
				ragSection = m.assembleGraphRAGSection(gCtx, userQuery, sessionKey, budgets.RAG)
				return nil
			})
		} else if m.ragService != nil {
			g.Go(func() error {
				ragSection = m.assembleRAGSection(gCtx, userQuery, sessionKey, budgets.RAG)
				return nil
			})
		}
	}

	// Memory retrieval
	if m.memoryProvider != nil && sessionKey != "" {
		g.Go(func() error {
			memorySection = m.assembleMemorySection(gCtx, sessionKey, budgets.Memory)
			return nil
		})
	}

	// Run summary retrieval
	if m.runSummaryProvider != nil && sessionKey != "" {
		g.Go(func() error {
			runSummarySection = m.assembleRunSummarySection(gCtx, sessionKey, budgets.RunSummary)
			return nil
		})
	}

	_ = g.Wait()

	// Shadow: coordinator comparison (fire-and-forget, does not block LLM).
	if m.coordinator != nil && m.coordinator.Shadow() && userQuery != "" {
		shadowRetrieved := oldRetrieved
		go func() {
			shadowCtx := context.Background()
			shadowFindings, err := m.coordinator.Retrieve(shadowCtx, userQuery, 0)
			if err != nil {
				m.logger.Warnw("shadow retrieval error", "error", err)
				return
			}
			if shadowRetrieved != nil {
				retrieval.CompareShadowResults(shadowRetrieved, shadowFindings, m.logger)
			}
		}()
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
