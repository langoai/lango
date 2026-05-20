package adk

import (
	"context"
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/plugin"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func TestWithPluginsAppendsPluginsInOrder(t *testing.T) {
	t.Parallel()

	first, err := plugin.New(plugin.Config{Name: "first"})
	require.NoError(t, err)
	second, err := plugin.New(plugin.Config{Name: "second"})
	require.NoError(t, err)

	var opts agentOptions
	WithPlugins(first)(&opts)
	WithPlugins(second)(&opts)

	require.Len(t, opts.plugins, 2)
	assert.Same(t, first, opts.plugins[0])
	assert.Same(t, second, opts.plugins[1])
}

func TestSubAgentNamesReturnsImmediateChildrenInOrder(t *testing.T) {
	t.Parallel()

	first := newStaticADKAgent(t, "operator", nil)
	second := newStaticADKAgent(t, "vault", nil)
	root := newStaticADKAgent(t, "root", []adkagent.Agent{first, second})

	assert.Equal(t, []string{"operator", "vault"}, subAgentNames(root))
}

func TestRunStreamingWrapperReturnsDetailedResponse(t *testing.T) {
	t.Parallel()

	root := newStaticADKAgent(t, "stream-wrapper", nil)
	agent, err := NewAgentFromADK(root, newMockStore())
	require.NoError(t, err)

	var chunks []string
	got, err := agent.RunStreaming(
		context.Background(),
		"session-wrapper",
		"say hello",
		func(chunk string) { chunks = append(chunks, chunk) },
	)
	require.NoError(t, err)

	assert.Equal(t, "hello", got)
	assert.Equal(t, []string{"he", "llo"}, chunks)
}

func TestRunStreamingDetailedNonPartialResponseDoesNotInvokeChunkCallback(t *testing.T) {
	t.Parallel()

	root, err := adkagent.New(adkagent.Config{
		Name:        "nonpartial-root",
		Description: "nonpartial root",
		Run: func(adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				yield(&session.Event{
					Author: "nonpartial-root",
					LLMResponse: model.LLMResponse{
						Content: &genai.Content{
							Role:  "model",
							Parts: []*genai.Part{{Text: "complete"}},
						},
					},
				}, nil)
			}
		},
	})
	require.NoError(t, err)
	agent, err := NewAgentFromADK(root, newMockStore())
	require.NoError(t, err)

	var chunks []string
	report, err := agent.RunStreamingDetailed(
		context.Background(),
		"session-nonpartial",
		"say complete",
		func(chunk string) { chunks = append(chunks, chunk) },
	)
	require.NoError(t, err)

	assert.Equal(t, "complete", report.Response)
	assert.Empty(t, chunks)
	assert.Equal(t, 1, report.Diagnostics.VisibleTextCount)
}

func TestContextAwareModelAdapterOptionSettersStoreDependencies(t *testing.T) {
	t.Parallel()

	adapter := &ContextAwareModelAdapter{}
	memoryProvider := &mockMemoryProvider{}
	runtimeAdapter := &RuntimeContextAdapter{}
	runSummaryProvider := &mockRunSummaryProvider{}
	budgetManager := &ContextBudgetManager{}
	compactor := &mockSessionCompactor{}
	recallProvider := &stubRecallProvider{}
	catalog := &stubCatalogSource{}
	modeResolver := &stubModeResolver{}
	waiter := &fakeSyncWaiter{}

	got := adapter.
		WithMemory(memoryProvider).
		WithRuntimeAdapter(runtimeAdapter).
		WithRunSummaryProvider(runSummaryProvider).
		WithMemoryLimits(3, 5).
		WithMemoryTokenBudget(700).
		WithBudgetManager(budgetManager).
		WithSessionCompactor(compactor).
		WithRecallProvider(recallProvider).
		WithCompactionSync(waiter, 0).
		WithCatalog(catalog).
		WithModeResolver(modeResolver)

	assert.Same(t, adapter, got)
	assert.Same(t, memoryProvider, adapter.memoryProvider)
	assert.Same(t, runtimeAdapter, adapter.runtimeAdapter)
	assert.Same(t, runSummaryProvider, adapter.runSummaryProvider)
	assert.Equal(t, 3, adapter.maxReflections)
	assert.Equal(t, 5, adapter.maxObservations)
	assert.Equal(t, 700, adapter.memoryTokenBudget)
	assert.Same(t, budgetManager, adapter.budgetManager)
	assert.Same(t, compactor, adapter.sessionCompactor)
	assert.Same(t, recallProvider, adapter.recallProvider)
	assert.Same(t, waiter, adapter.compactionSync)
	assert.Same(t, catalog, adapter.catalogSource)
	assert.Same(t, modeResolver, adapter.modeResolver)
}

func newStaticADKAgent(t *testing.T, name string, subAgents []adkagent.Agent) adkagent.Agent {
	t.Helper()

	agent, err := adkagent.New(adkagent.Config{
		Name:        name,
		Description: name + " test agent",
		SubAgents:   subAgents,
		Run: func(adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				events := []*session.Event{
					{
						Author: name,
						LLMResponse: model.LLMResponse{
							Content: &genai.Content{
								Role:  "model",
								Parts: []*genai.Part{{Text: "he"}},
							},
							Partial: true,
						},
					},
					{
						Author: name,
						LLMResponse: model.LLMResponse{
							Content: &genai.Content{
								Role:  "model",
								Parts: []*genai.Part{{Text: "llo"}},
							},
							Partial: true,
						},
					},
				}
				for _, event := range events {
					if !yield(event, nil) {
						return
					}
				}
			}
		},
	})
	require.NoError(t, err)
	return agent
}

type stubRecallProvider struct{}

func (stubRecallProvider) RecallRecent(context.Context, string, string) ([]RecallMatch, error) {
	return nil, nil
}

type stubCatalogSource struct{}

func (stubCatalogSource) BuildToolCatalogSection(string) string {
	return "tools"
}

type stubModeResolver struct{}

func (stubModeResolver) LookupModeHint(string) string {
	return "hint"
}
