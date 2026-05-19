package adk

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	adkmodel "google.golang.org/adk/model"
	adktool "google.golang.org/adk/tool"
	"google.golang.org/genai"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/prompt"
	"github.com/langoai/lango/internal/provider"
	"github.com/langoai/lango/internal/retrieval"
	"github.com/langoai/lango/internal/session"
)

func TestWave50ContextAwareModelAdapterCoordinatorPublishesInjectedContext(t *testing.T) {
	t.Parallel()

	p := &mockProvider{
		id: "wave50-provider",
		events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "ok"},
			{Type: provider.StreamEventDone},
		},
	}
	inner := NewModelAdapter(p, "wave50-model")
	retriever := knowledge.NewContextRetriever(nil, 5, zap.NewNop().Sugar())
	coordinatorAgent := &wave50RetrievalAgent{
		findings: []retrieval.Finding{{
			Layer:        knowledge.LayerUserKnowledge,
			Key:          "project-alpha",
			Content:      "Project Alpha uses SQLite for local test fixtures.",
			Score:        0.91,
			SearchSource: "fake-search",
			Category:     "testing",
		}},
	}
	coordinator := retrieval.NewRetrievalCoordinator([]retrieval.RetrievalAgent{coordinatorAgent}, zap.NewNop().Sugar())
	bus := eventbus.New()

	var gotEvent eventbus.ContextInjectedEvent
	var gotEventCount int
	eventbus.SubscribeTyped(bus, func(e eventbus.ContextInjectedEvent) {
		gotEvent = e
		gotEventCount++
	})

	adapter := NewContextAwareModelAdapter(inner, retriever, prompt.DefaultBuilder(), zap.NewNop().Sugar()).
		WithCoordinator(coordinator).
		WithEventBus(bus)
	assert.Equal(t, "wave50-model", adapter.Name())

	ctx := session.WithSessionKey(context.Background(), "wave50:session")
	ctx = session.WithTurnID(ctx, "turn-wave50")
	req := &adkmodel.LLMRequest{
		Model: "wave50-model",
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: "Remember Project Alpha testing details"}}},
		},
	}

	for _, err := range adapter.GenerateContent(ctx, req, false) {
		require.NoError(t, err)
	}

	assert.Equal(t, "Remember Project Alpha testing details", coordinatorAgent.query())
	assert.Equal(t, 10, coordinatorAgent.limit())

	require.NotNil(t, req.Config)
	require.NotNil(t, req.Config.SystemInstruction)
	require.NotEmpty(t, req.Config.SystemInstruction.Parts)
	systemPrompt := req.Config.SystemInstruction.Parts[0].Text
	assert.Contains(t, systemPrompt, "## User Knowledge")
	assert.Contains(t, systemPrompt, "Project Alpha uses SQLite for local test fixtures.")

	require.GreaterOrEqual(t, len(p.lastParams.Messages), 2)
	assert.Equal(t, "system", string(p.lastParams.Messages[0].Role))
	assert.Contains(t, p.lastParams.Messages[0].Content, "project-alpha")

	require.Equal(t, 1, gotEventCount)
	assert.Equal(t, "turn-wave50", gotEvent.TurnID)
	assert.Equal(t, "wave50:session", gotEvent.SessionKey)
	assert.Equal(t, "Remember Project Alpha testing details", gotEvent.Query)
	require.Len(t, gotEvent.Items, 1)
	assert.Equal(t, eventbus.ContextInjectedItem{
		Layer:         knowledge.LayerUserKnowledge.String(),
		Key:           "project-alpha",
		Score:         0.91,
		Source:        "fake-search",
		Category:      "testing",
		TokenEstimate: gotEvent.Items[0].TokenEstimate,
	}, gotEvent.Items[0])
	assert.Positive(t, gotEvent.Items[0].TokenEstimate)
	assert.Positive(t, gotEvent.KnowledgeTokens)
	assert.Equal(t, gotEvent.KnowledgeTokens, gotEvent.TotalTokens)
	assert.False(t, gotEvent.Timestamp.IsZero())
}

func TestWave50AdaptToolWithTimeoutReportsDeadline(t *testing.T) {
	t.Parallel()

	internalTool := &agent.Tool{
		Name:        "deadline_tool",
		Description: "Waits for the adapted context deadline",
		Parameters:  map[string]interface{}{"input": agent.ParameterDef{Type: "string"}},
		Handler: func(ctx context.Context, _ map[string]interface{}) (interface{}, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}

	adapted, err := AdaptToolWithTimeout(internalTool, time.Millisecond)
	require.NoError(t, err)

	runnable := adapted.(interface {
		Run(adktool.Context, any) (map[string]any, error)
	})
	got, err := runnable.Run(&wave31ToolContext{Context: context.Background()}, map[string]any{"input": "x"})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, `tool "deadline_tool" timed out after 1ms`)
}

func TestWave50AdaptToolForAgentInjectsOwnerWithoutTimeout(t *testing.T) {
	t.Parallel()

	internalTool := &agent.Tool{
		Name:        "owner_tool",
		Description: "Returns the injected owner",
		Parameters:  map[string]interface{}{},
		Handler: func(ctx context.Context, _ map[string]interface{}) (interface{}, error) {
			if deadline, ok := ctx.Deadline(); ok {
				return nil, errors.New("unexpected timeout deadline: " + deadline.String())
			}
			return AgentNameFromContext(ctx), nil
		},
	}

	adapted, err := AdaptToolForAgent(internalTool, "planner")
	require.NoError(t, err)

	runnable := adapted.(interface {
		Run(adktool.Context, any) (map[string]any, error)
	})
	got, err := runnable.Run(&wave31ToolContext{Context: context.Background()}, map[string]any{})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"result": "planner"}, got)
}

func TestWave50BuildInputSchemaConvertsInterfaceRequiredBranches(t *testing.T) {
	t.Parallel()

	internalTool := &agent.Tool{
		Name: "schema_tool",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query",
				},
				"filter": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"category": map[string]interface{}{
							"type":        "string",
							"description": "Knowledge category",
						},
						"source": map[string]interface{}{
							"type": "string",
						},
					},
					"required": []interface{}{"category", 42, "source"},
				},
			},
			"required": []interface{}{"query", false, "filter"},
		},
	}

	schema := buildInputSchema(internalTool)
	require.Equal(t, "object", schema.Type)
	assert.Equal(t, []string{"query", "filter"}, schema.Required)
	assert.Equal(t, "string", schema.Properties["query"].Type)
	assert.Equal(t, "Search query", schema.Properties["query"].Description)

	filter := schema.Properties["filter"]
	require.NotNil(t, filter)
	assert.Equal(t, "object", filter.Type)
	require.NotNil(t, filter.Properties)
	assert.Equal(t, "string", filter.Properties["category"].Type)
	assert.Equal(t, "Knowledge category", filter.Properties["category"].Description)
	assert.Equal(t, []string{"category", "source"}, filter.Required)
}

type wave50RetrievalAgent struct {
	mu       sync.Mutex
	findings []retrieval.Finding
	gotQuery string
	gotLimit int
}

func (a *wave50RetrievalAgent) Name() string {
	return "wave50-retrieval"
}

func (a *wave50RetrievalAgent) Layers() []knowledge.ContextLayer {
	return []knowledge.ContextLayer{knowledge.LayerUserKnowledge}
}

func (a *wave50RetrievalAgent) Search(_ context.Context, query string, limit int) ([]retrieval.Finding, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.gotQuery = query
	a.gotLimit = limit
	return append([]retrieval.Finding(nil), a.findings...), nil
}

func (a *wave50RetrievalAgent) query() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.gotQuery
}

func (a *wave50RetrievalAgent) limit() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.gotLimit
}

var (
	_ retrieval.RetrievalAgent = (*wave50RetrievalAgent)(nil)
)
