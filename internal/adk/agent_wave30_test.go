package adk

import (
	"context"
	"errors"
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/plugin"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	internal "github.com/langoai/lango/internal/session"
)

func TestWave30AgentOptionHelpersPopulateOptions(t *testing.T) {
	t.Parallel()

	var rootSessions []string
	var lifecycleEvents []internal.SessionLifecycleEvent
	fixProvider := &wave30FixProvider{}
	plugins := []*plugin.Plugin{{}}

	var opts agentOptions
	WithAgentTokenBudget(321)(&opts)
	WithAgentMaxTurns(12)(&opts)
	WithAgentErrorFixProvider(fixProvider)(&opts)
	WithAgentRootSessionObserver(func(sessionID string) {
		rootSessions = append(rootSessions, sessionID)
	})(&opts)
	WithAgentChildLifecycleHook(func(event internal.SessionLifecycleEvent) {
		lifecycleEvents = append(lifecycleEvents, event)
	})(&opts)
	names := []string{"operator", "vault"}
	WithAgentIsolatedAgents(names)(&opts)
	names[0] = "mutated"
	WithPlugins(plugins...)(&opts)

	require.NotNil(t, opts.errorFixProvider)
	require.NotNil(t, opts.rootSessionObserver)
	require.NotNil(t, opts.childLifecycleHook)
	assert.Equal(t, 321, opts.tokenBudget)
	assert.Equal(t, 12, opts.maxTurns)
	assert.Equal(t, []string{"operator", "vault"}, opts.isolatedAgents)
	assert.Equal(t, plugins, opts.plugins)

	opts.rootSessionObserver("root-session")
	opts.childLifecycleHook(internal.SessionLifecycleEvent{ParentKey: "parent"})
	assert.Equal(t, []string{"root-session"}, rootSessions)
	assert.Len(t, lifecycleEvents, 1)
}

func TestWave30AgentMethodsAndSubAgentNamesExposeConfiguredState(t *testing.T) {
	t.Parallel()

	operator := newWave30StaticAgent(t, "operator", nil, nil)
	vault := newWave30StaticAgent(t, "vault", nil, nil)
	root := newWave30StaticAgent(t, "root", []adkagent.Agent{operator, vault}, nil)

	agent := (&Agent{adkAgent: root}).WithMaxTurns(9).WithErrorFixProvider(&wave30FixProvider{})

	assert.Equal(t, 9, agent.maxTurns)
	assert.NotNil(t, agent.errorFixProvider)
	assert.Same(t, root, agent.ADKAgent())
	assert.Equal(t, []string{"operator", "vault"}, subAgentNames(root))
}

func TestWave30RunAndCollectRetriesRejectTextFromSubAgent(t *testing.T) {
	operator := newWave30StaticAgent(t, "operator", nil, nil)
	var calls int
	root := newWave30StaticAgent(t, "wave30-root", []adkagent.Agent{operator}, func() (*session.Event, error) {
		calls++
		if calls == 1 {
			return wave30TextEvent("operator", "[REJECT] operator cannot complete this"), nil
		}
		return wave30TextEvent("wave30-root", "routed answer"), nil
	})
	agent, err := NewAgentFromADK(root, newMockStore())
	require.NoError(t, err)

	got, err := agent.RunAndCollect(context.Background(), "wave30-reject", "delegate this")

	require.NoError(t, err)
	assert.Equal(t, "routed answer", got)
	assert.Equal(t, 2, calls)
}

func TestWave30RunAndCollectAppliesLearnedFixAfterRecoverableError(t *testing.T) {
	provider := &wave30FixProvider{fix: "retry with a smaller query"}
	var calls int
	root := newWave30StaticAgent(t, "wave30-fix-root", nil, func() (*session.Event, error) {
		calls++
		if calls == 1 {
			return nil, errors.New("tool not found: missing_wave30_tool")
		}
		return wave30TextEvent("wave30-fix-root", "fixed answer"), nil
	})
	agent, err := NewAgentFromADK(root, newMockStore(), WithAgentErrorFixProvider(provider))
	require.NoError(t, err)

	got, err := agent.RunAndCollect(context.Background(), "wave30-fix", "use missing tool")

	require.NoError(t, err)
	assert.Equal(t, "fixed answer", got)
	assert.Equal(t, 2, calls)
	assert.ErrorContains(t, provider.observed, "tool not found: missing_wave30_tool")
}

func newWave30StaticAgent(
	t *testing.T,
	name string,
	subAgents []adkagent.Agent,
	next func() (*session.Event, error),
) adkagent.Agent {
	t.Helper()

	agent, err := adkagent.New(adkagent.Config{
		Name:        name,
		Description: name + " wave30 test agent",
		SubAgents:   subAgents,
		Run: func(adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				if next == nil {
					return
				}
				event, err := next()
				yield(event, err)
			}
		},
	})
	require.NoError(t, err)
	return agent
}

func wave30TextEvent(author, text string) *session.Event {
	return &session.Event{
		Author: author,
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{Role: "model", Parts: []*genai.Part{{Text: text}}},
		},
	}
}

type wave30FixProvider struct {
	fix      string
	observed error
}

func (p *wave30FixProvider) GetFixForError(_ context.Context, _ string, err error) (string, bool) {
	p.observed = err
	if p.fix == "" {
		return "retry", true
	}
	return p.fix, true
}
