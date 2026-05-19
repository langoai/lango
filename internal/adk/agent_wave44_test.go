package adk

import (
	"context"
	"errors"
	"iter"
	"testing"

	"github.com/stretchr/testify/require"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	internal "github.com/langoai/lango/internal/session"
)

type wave44FixProvider struct {
	calls int
	errs  []string
}

func (p *wave44FixProvider) GetFixForError(_ context.Context, toolName string, err error) (string, bool) {
	p.calls++
	p.errs = append(p.errs, toolName+":"+err.Error())
	return "retry with deterministic fix", true
}

func TestWave44RunAndCollectUsesLearnedFixRetryResponse(t *testing.T) {
	t.Parallel()

	var inputs []string
	root, err := adkagent.New(adkagent.Config{
		Name:        "wave44-root",
		Description: "wave44 root",
		Run: func(ctx adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			inputs = append(inputs, ctx.UserContent().Parts[0].Text)
			attempt := len(inputs)
			return func(yield func(*session.Event, error) bool) {
				if attempt == 1 {
					yield(nil, errors.New("tool failed before fix"))
					return
				}
				yield(&session.Event{
					Author: "wave44-root",
					LLMResponse: model.LLMResponse{
						Content: &genai.Content{Role: "model", Parts: []*genai.Part{{Text: "fixed response"}}},
					},
				}, nil)
			}
		},
	})
	require.NoError(t, err)

	provider := &wave44FixProvider{}
	agent, err := NewAgentFromADK(root, newMockStore(), WithAgentErrorFixProvider(provider))
	require.NoError(t, err)

	got, err := agent.RunAndCollect(context.Background(), "wave44-learned-fix", "original input")
	require.NoError(t, err)
	require.Equal(t, "fixed response", got)
	require.Equal(t, 1, provider.calls)
	require.Len(t, inputs, 2)
	require.Equal(t, "original input", inputs[0])
	require.Contains(t, inputs[1], "Previous action failed with:")
	require.Contains(t, inputs[1], "retry with deterministic fix")
}

func TestWave44NewAgentFromADKAppliesOptionsAndFiltersIsolatedNames(t *testing.T) {
	t.Parallel()

	var rootObserved string
	root := newStaticADKAgent(t, "wave44-options-root", nil)
	agent, err := NewAgentFromADK(
		root,
		newMockStore(),
		WithAgentTokenBudget(321),
		WithAgentMaxTurns(9),
		WithAgentRootSessionObserver(func(sessionID string) { rootObserved = sessionID }),
		WithAgentChildLifecycleHook(func(internal.SessionLifecycleEvent) {}),
		WithAgentIsolatedAgents([]string{"operator", "", "vault"}),
	)
	require.NoError(t, err)

	require.Equal(t, 9, agent.maxTurns)
	require.Equal(t, 321, agent.sessionService.tokenBudget)
	require.NotNil(t, agent.sessionService.rootSessionObserver)
	require.NotNil(t, agent.sessionService.childStore)
	require.Contains(t, agent.sessionService.isolatedAgents, "operator")
	require.Contains(t, agent.sessionService.isolatedAgents, "vault")
	require.NotContains(t, agent.sessionService.isolatedAgents, "")
	require.Contains(t, agent.isolatedAgents, "operator")
	require.Contains(t, agent.isolatedAgents, "vault")
	require.NotContains(t, agent.isolatedAgents, "")

	agent.sessionService.rootSessionObserver("root-session")
	require.Equal(t, "root-session", rootObserved)
}
