package adk

import (
	"context"
	"errors"
	"iter"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	internal "github.com/langoai/lango/internal/session"
)

func TestRunAndCollectRetriesMissingAgentWhenSubAgentsExist(t *testing.T) {
	t.Parallel()

	operator := newAgentCollectionErrorBranchesStaticAgent(t, "operator")
	var inputs []string
	root := newAgentCollectionErrorBranchesSequenceAgent(t, "collection-root", []adkagent.Agent{operator}, func(ctx adkagent.InvocationContext) ([]*session.Event, error) {
		inputs = append(inputs, ctx.UserContent().Parts[0].Text)
		if len(inputs) == 1 {
			return nil, errors.New("failed to find agent: ghost_agent")
		}
		return []*session.Event{agentCollectionErrorBranchesTextEvent("collection-root", "rerouted answer", false)}, nil
	})
	agent, err := NewAgentFromADK(root, newMockStore())
	require.NoError(t, err)

	got, err := agent.RunAndCollect(context.Background(), "collection-missing-agent-success", "delegate")

	require.NoError(t, err)
	assert.Equal(t, "rerouted answer", got)
	require.Len(t, inputs, 2)
	assert.Equal(t, "delegate", inputs[0])
	assert.Contains(t, inputs[1], `Agent "ghost_agent" does not exist`)
	assert.Contains(t, inputs[1], "Valid agents: operator")
}

func TestRunAndCollectReturnsLongerRetryPartialWhenMissingAgentRetryFails(t *testing.T) {
	t.Parallel()

	operator := newAgentCollectionErrorBranchesStaticAgent(t, "operator")
	var calls int
	root := newAgentCollectionErrorBranchesSequenceAgent(t, "collection-retry-fails-root", []adkagent.Agent{operator}, func(adkagent.InvocationContext) ([]*session.Event, error) {
		calls++
		if calls == 1 {
			return []*session.Event{agentCollectionErrorBranchesTextEvent("collection-retry-fails-root", "short", false)}, errors.New("failed to find agent: ghost_agent")
		}
		return []*session.Event{agentCollectionErrorBranchesTextEvent("collection-retry-fails-root", "longer retry partial", false)}, errors.New("retry exploded")
	})
	agent, err := NewAgentFromADK(root, newMockStore())
	require.NoError(t, err)

	got, err := agent.RunAndCollect(context.Background(), "collection-missing-agent-retry-fails", "delegate")

	require.Error(t, err)
	assert.Equal(t, "longer retry partial", got)
	assert.Equal(t, 2, calls)
	var agentErr *AgentError
	require.ErrorAs(t, err, &agentErr)
	assert.Contains(t, agentErr.CauseDetail, "retry exploded")
}

func TestRunAndCollectDoesNotRetryMissingCustomAgentWithoutSubAgents(t *testing.T) {
	t.Parallel()

	var calls int
	root := newAgentCollectionErrorBranchesSequenceAgent(t, "collection-no-subagents-root", nil, func(adkagent.InvocationContext) ([]*session.Event, error) {
		calls++
		return nil, errors.New("failed to find agent: custom_agent")
	})
	agent, err := NewAgentFromADK(root, newMockStore())
	require.NoError(t, err)

	got, err := agent.RunAndCollect(context.Background(), "collection-missing-agent-no-subagents", "delegate")

	require.Error(t, err)
	assert.Empty(t, got)
	assert.Equal(t, 1, calls)
}

func TestAgentCleanupFailedTurnForwardsToSessionServiceAndNilIsNoop(t *testing.T) {
	t.Parallel()

	(&Agent{}).cleanupFailedTurn("missing-session", "agent error")

	store := newMockStore()
	sess := &internal.Session{
		Key:       "cleanup-parent",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, store.Create(sess))

	adapter := NewSessionAdapter(sess, store, "lango-orchestrator")
	svc := NewSessionServiceAdapter(store, "lango-orchestrator").
		WithChildLifecycleHook(func(internal.SessionLifecycleEvent) {}).
		WithIsolatedAgents([]string{"operator"})
	require.NoError(t, svc.AppendEvent(context.Background(), adapter, agentCollectionErrorBranchesTextEvent("operator", "child-only result", false)))

	agent := &Agent{sessionService: svc}
	agent.cleanupFailedTurn("cleanup-parent", "agent error")

	dbMsgs := store.messages["cleanup-parent"]
	require.Len(t, dbMsgs, 1)
	assert.Equal(t, "lango-orchestrator", dbMsgs[0].Author)
	assert.Contains(t, dbMsgs[0].Content, "operator discarded: agent error")
	assert.Empty(t, svc.activeChild["cleanup-parent"])
}

func TestCollectionHelpersReturnEmptyWhenNoFunctionCallsExist(t *testing.T) {
	t.Parallel()

	event := &session.Event{
		Author: "operator",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{Parts: []*genai.Part{
				{Text: "visible text"},
				{FunctionResponse: &genai.FunctionResponse{Name: "search", Response: map[string]any{"ok": true}}},
			}},
		},
	}

	assert.Empty(t, extractPrimaryToolName(event))
	assert.Empty(t, extractPrimaryToolSignature(event))
}

func newAgentCollectionErrorBranchesStaticAgent(t *testing.T, name string) adkagent.Agent {
	t.Helper()

	agent, err := adkagent.New(adkagent.Config{
		Name:        name,
		Description: name + " static test agent",
		Run: func(adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(func(*session.Event, error) bool) {}
		},
	})
	require.NoError(t, err)
	return agent
}

func newAgentCollectionErrorBranchesSequenceAgent(
	t *testing.T,
	name string,
	subAgents []adkagent.Agent,
	next func(adkagent.InvocationContext) ([]*session.Event, error),
) adkagent.Agent {
	t.Helper()

	agent, err := adkagent.New(adkagent.Config{
		Name:        name,
		Description: name + " sequence test agent",
		SubAgents:   subAgents,
		Run: func(ctx adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			events, err := next(ctx)
			return func(yield func(*session.Event, error) bool) {
				for _, event := range events {
					if !yield(event, nil) {
						return
					}
				}
				if err != nil {
					yield(nil, err)
				}
			}
		},
	})
	require.NoError(t, err)
	return agent
}

func agentCollectionErrorBranchesTextEvent(author, text string, partial bool) *session.Event {
	return &session.Event{
		Author: author,
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{Role: "model", Parts: []*genai.Part{{Text: text}}},
			Partial: partial,
		},
	}
}
