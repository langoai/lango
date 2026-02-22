package adk

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/adk/model"
	"google.golang.org/genai"

	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/prompt"
	"github.com/langoai/lango/internal/session"
)

// MemoryProvider retrieves observations and reflections for a session.
type MemoryProvider interface {
	ListObservations(ctx context.Context, sessionKey string) ([]memory.Observation, error)
	ListReflections(ctx context.Context, sessionKey string) ([]memory.Reflection, error)
	ListRecentReflections(ctx context.Context, sessionKey string, limit int) ([]memory.Reflection, error)
	ListRecentObservations(ctx context.Context, sessionKey string, limit int) ([]memory.Observation, error)
}

// ContextAwareModelAdapter wraps a ModelAdapter with context retrieval.
// Before each LLM call, it retrieves relevant knowledge and injects it
// into the system instruction.
type ContextAwareModelAdapter struct {
	inner           *ModelAdapter
	retriever       *knowledge.ContextRetriever
	memoryProvider  MemoryProvider
	ragService      *embedding.RAGService
	ragOpts         embedding.RetrieveOptions
	graphRAG        *graph.GraphRAGService
	runtimeAdapter  *RuntimeContextAdapter
	basePrompt      string
	maxReflections  int
	maxObservations int
	logger          *zap.SugaredLogger
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

// WithMemoryLimits sets the maximum number of reflections and observations
// to include in the LLM context. Zero means unlimited (existing behavior).
func (m *ContextAwareModelAdapter) WithMemoryLimits(maxReflections, maxObservations int) *ContextAwareModelAdapter {
	m.maxReflections = maxReflections
	m.maxObservations = maxObservations
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

	var knowledgeSection, ragSection, memorySection string

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
				knowledgeSection = m.retriever.AssemblePrompt("", retrieved)
			}
			return nil
		})
	}

	// RAG/GraphRAG retrieval
	if userQuery != "" {
		if m.graphRAG != nil {
			g.Go(func() error {
				ragSection = m.assembleGraphRAGSection(gCtx, userQuery, sessionKey)
				return nil
			})
		} else if m.ragService != nil {
			g.Go(func() error {
				ragSection = m.assembleRAGSection(gCtx, userQuery, sessionKey)
				return nil
			})
		}
	}

	// Memory retrieval
	if m.memoryProvider != nil && sessionKey != "" {
		g.Go(func() error {
			memorySection = m.assembleMemorySection(gCtx, sessionKey)
			return nil
		})
	}

	_ = g.Wait()

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

// assembleMemorySection builds the "Conversation Memory" section from observations and reflections.
func (m *ContextAwareModelAdapter) assembleMemorySection(ctx context.Context, sessionKey string) string {
	var reflections []memory.Reflection
	var observations []memory.Observation
	var err error

	if m.maxReflections > 0 {
		reflections, err = m.memoryProvider.ListRecentReflections(ctx, sessionKey, m.maxReflections)
	} else {
		reflections, err = m.memoryProvider.ListReflections(ctx, sessionKey)
	}
	if err != nil {
		m.logger.Warnw("memory reflection retrieval error", "error", err)
	}

	if m.maxObservations > 0 {
		observations, err = m.memoryProvider.ListRecentObservations(ctx, sessionKey, m.maxObservations)
	} else {
		observations, err = m.memoryProvider.ListObservations(ctx, sessionKey)
	}
	if err != nil {
		m.logger.Warnw("memory observation retrieval error", "error", err)
	}

	if len(reflections) == 0 && len(observations) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("## Conversation Memory\n")

	if len(reflections) > 0 {
		b.WriteString("\n### Summary\n")
		for _, ref := range reflections {
			b.WriteString(ref.Content)
			b.WriteString("\n")
		}
	}

	if len(observations) > 0 {
		b.WriteString("\n### Recent Observations\n")
		for _, obs := range observations {
			b.WriteString("- ")
			b.WriteString(obs.Content)
			b.WriteString("\n")
		}
	}

	return b.String()
}

// assembleGraphRAGSection builds a combined section from vector search + graph expansion.
func (m *ContextAwareModelAdapter) assembleGraphRAGSection(ctx context.Context, query, sessionKey string) string {
	opts := graph.VectorRetrieveOptions{
		Collections: m.ragOpts.Collections,
		Limit:       m.ragOpts.Limit,
		SessionKey:  m.ragOpts.SessionKey,
		MaxDistance: m.ragOpts.MaxDistance,
	}
	if sessionKey != "" {
		opts.SessionKey = sessionKey
	}
	result, err := m.graphRAG.Retrieve(ctx, query, opts)
	if err != nil {
		m.logger.Warnw("graph rag retrieval error", "error", err)
		return ""
	}
	return m.graphRAG.AssembleSection(result)
}

// assembleRAGSection builds a "Semantic Context" section from RAG retrieval results.
func (m *ContextAwareModelAdapter) assembleRAGSection(ctx context.Context, query, sessionKey string) string {
	opts := m.ragOpts
	if sessionKey != "" {
		opts.SessionKey = sessionKey
	}
	results, err := m.ragService.Retrieve(ctx, query, opts)
	if err != nil {
		m.logger.Warnw("rag retrieval error", "error", err)
		return ""
	}
	if len(results) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("## Semantic Context (RAG)\n")
	for _, r := range results {
		if r.Content == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("\n### [%s] %s\n", r.Collection, r.SourceID))
		b.WriteString(r.Content)
		b.WriteString("\n")
	}
	return b.String()
}

// extractLastUserMessage finds the last user message from the content history.
func extractLastUserMessage(contents []*genai.Content) string {
	for i := len(contents) - 1; i >= 0; i-- {
		c := contents[i]
		if c.Role == "user" {
			for _, p := range c.Parts {
				if p.Text != "" {
					return p.Text
				}
			}
		}
	}
	return ""
}
