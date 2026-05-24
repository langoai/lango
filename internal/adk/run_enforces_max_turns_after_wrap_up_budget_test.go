package adk

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"testing"

	"github.com/stretchr/testify/require"
	adkagent "google.golang.org/adk/agent"
	adkmodel "google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func TestRunEnforcesMaxTurnsAfterWrapUpBudget(t *testing.T) {
	t.Parallel()

	root := newRunEnforcesMaxTurnsAfterWrapUpBudgetSequenceAgent(t, "runEnforcesMaxTurnsAfterWrapUpBudget7-turn-root", nil, func() []*session.Event {
		return []*session.Event{
			runEnforcesMaxTurnsAfterWrapUpBudgetFunctionCallEvent("runEnforcesMaxTurnsAfterWrapUpBudget7-turn-root", "lookup", map[string]any{"step": 1}),
			runEnforcesMaxTurnsAfterWrapUpBudgetFunctionCallEvent("runEnforcesMaxTurnsAfterWrapUpBudget7-turn-root", "lookup", map[string]any{"step": 2}),
			runEnforcesMaxTurnsAfterWrapUpBudgetFunctionCallEvent("runEnforcesMaxTurnsAfterWrapUpBudget7-turn-root", "lookup", map[string]any{"step": 3}),
		}
	})
	agent, err := NewAgentFromADK(root, newMockStore(), WithAgentMaxTurns(1))
	require.NoError(t, err)

	var events []*session.Event
	for event, err := range agent.Run(context.Background(), "runEnforcesMaxTurnsAfterWrapUpBudget7-max-turns", "run tools") {
		if err != nil {
			require.Nil(t, event)
			require.ErrorContains(t, err, "agent exceeded maximum turn limit (1)")
			require.Len(t, events, 2)
			return
		}
		require.NotNil(t, event)
		events = append(events, event)
	}
	t.Fatal("expected max-turn error after wrap-up budget was exhausted")
}

func TestRunAndCollectClassifiesToolChurnError(t *testing.T) {
	t.Parallel()

	root := newRunEnforcesMaxTurnsAfterWrapUpBudgetSequenceAgent(t, "runEnforcesMaxTurnsAfterWrapUpBudget7-churn-root", nil, func() []*session.Event {
		events := make([]*session.Event, maxConsecutiveSameToolCalls)
		for i := range events {
			events[i] = runEnforcesMaxTurnsAfterWrapUpBudgetFunctionCallEvent(
				"runEnforcesMaxTurnsAfterWrapUpBudget7-churn-root",
				"repeat_lookup",
				map[string]any{"query": "same"},
			)
		}
		return events
	})
	agent, err := NewAgentFromADK(root, newMockStore(), WithAgentMaxTurns(20))
	require.NoError(t, err)

	got, err := agent.RunAndCollect(context.Background(), "runEnforcesMaxTurnsAfterWrapUpBudget7-tool-churn", "repeat")

	require.Error(t, err)
	require.Empty(t, got)
	var agentErr *AgentError
	require.ErrorAs(t, err, &agentErr)
	require.Equal(t, ErrToolChurn, agentErr.Code)
	require.Equal(t, "repeated_call_signature", agentErr.CauseClass)
	require.Contains(t, agentErr.CauseDetail, "repeat_lookup")
}

func TestRunAndCollectCallbacksSeeNilContentAndPartialError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("tool not found: runEnforcesMaxTurnsAfterWrapUpBudget_missing_tool")
	root, err := adkagent.New(adkagent.Config{
		Name:        "runEnforcesMaxTurnsAfterWrapUpBudget7-callback-root",
		Description: "runEnforcesMaxTurnsAfterWrapUpBudget7 callback root",
		Run: func(adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				if !yield(&session.Event{Author: "runEnforcesMaxTurnsAfterWrapUpBudget7-callback-root"}, nil) {
					return
				}
				if !yield(runEnforcesMaxTurnsAfterWrapUpBudgetTextEvent("runEnforcesMaxTurnsAfterWrapUpBudget7-callback-root", "partial answer", true), nil) {
					return
				}
				yield(nil, wantErr)
			}
		},
	})
	require.NoError(t, err)
	agent, err := NewAgentFromADK(root, newMockStore())
	require.NoError(t, err)

	var eventAuthors []string
	var activityCount int
	var finished bool
	got, err := agent.RunAndCollect(
		context.Background(),
		"runEnforcesMaxTurnsAfterWrapUpBudget7-callbacks",
		"collect callbacks",
		WithOnEvent(func(event *session.Event) { eventAuthors = append(eventAuthors, event.Author) }),
		WithOnActivity(func() { activityCount++ }),
		WithOnFinish(func() { finished = true }),
	)

	require.Error(t, err)
	require.Equal(t, "partial answer", got)
	require.Equal(t, []string{"runEnforcesMaxTurnsAfterWrapUpBudget7-callback-root", "runEnforcesMaxTurnsAfterWrapUpBudget7-callback-root"}, eventAuthors)
	require.Equal(t, 1, activityCount)
	require.True(t, finished)
	var agentErr *AgentError
	require.ErrorAs(t, err, &agentErr)
	require.Equal(t, ErrToolError, agentErr.Code)
	require.Equal(t, CauseToolNotFound, agentErr.CauseClass)
	require.Equal(t, "partial answer", agentErr.Partial)
}

func TestRunStreamingDetailedReportsContextCancellationAfterIteratorEnds(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	root, err := adkagent.New(adkagent.Config{
		Name:        "runEnforcesMaxTurnsAfterWrapUpBudget7-cancel-root",
		Description: "runEnforcesMaxTurnsAfterWrapUpBudget7 cancel root",
		Run: func(adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				if !yield(runEnforcesMaxTurnsAfterWrapUpBudgetTextEvent("runEnforcesMaxTurnsAfterWrapUpBudget7-cancel-root", "before cancel", true), nil) {
					return
				}
				cancel()
			}
		},
	})
	require.NoError(t, err)
	agent, err := NewAgentFromADK(root, newMockStore())
	require.NoError(t, err)

	var chunks []string
	report, err := agent.RunStreamingDetailed(
		ctx,
		"runEnforcesMaxTurnsAfterWrapUpBudget7-canceled",
		"cancel after partial",
		func(chunk string) { chunks = append(chunks, chunk) },
	)

	require.Error(t, err)
	require.Equal(t, "before cancel", report.Response)
	require.Equal(t, []string{"before cancel"}, chunks)
	require.Equal(t, 1, report.Diagnostics.VisibleTextCount)
	var agentErr *AgentError
	require.ErrorAs(t, err, &agentErr)
	require.Equal(t, ErrTimeout, agentErr.Code)
	require.Equal(t, context.Canceled.Error(), agentErr.CauseDetail)
}

func TestStreamingDiagnosticsCountMixedToolEvents(t *testing.T) {
	t.Parallel()

	child := newStaticADKAgent(t, "operator", nil)
	root := newRunEnforcesMaxTurnsAfterWrapUpBudgetSequenceAgent(t, "runEnforcesMaxTurnsAfterWrapUpBudget7-diagnostics-root", []adkagent.Agent{child}, func() []*session.Event {
		return []*session.Event{
			{
				Author:  "runEnforcesMaxTurnsAfterWrapUpBudget7-diagnostics-root",
				Actions: session.EventActions{TransferToAgent: "operator"},
			},
			{
				Author: "runEnforcesMaxTurnsAfterWrapUpBudget7-diagnostics-root",
				LLMResponse: adkmodel.LLMResponse{
					Content: &genai.Content{Parts: []*genai.Part{
						{FunctionCall: &genai.FunctionCall{
							Name: "transfer_to_agent",
							Args: map[string]any{"agent_name": "operator"},
						}},
					}},
				},
			},
			{
				Author: "operator",
				LLMResponse: adkmodel.LLMResponse{
					Content: &genai.Content{Parts: []*genai.Part{
						{Text: "tool result text"},
						{FunctionResponse: &genai.FunctionResponse{
							Name:     "search",
							Response: map[string]any{"ok": true},
						}},
					}},
				},
			},
			runEnforcesMaxTurnsAfterWrapUpBudgetTextEvent("runEnforcesMaxTurnsAfterWrapUpBudget7-diagnostics-root", "final answer", false),
		}
	})
	agent, err := NewAgentFromADK(root, newMockStore())
	require.NoError(t, err)

	report, err := agent.RunStreamingDetailed(
		context.Background(),
		"runEnforcesMaxTurnsAfterWrapUpBudget7-diagnostics",
		"delegate safely",
		nil,
	)

	require.NoError(t, err)
	require.Equal(t, "tool result textfinal answer", report.Response)
	require.Equal(t, RunDiagnostics{
		VisibleTextCount:        2,
		ToolCallCount:           1,
		ToolResultCount:         1,
		DelegationCount:         1,
		DirectRootToolCallCount: 1,
	}, report.Diagnostics)
}

func newRunEnforcesMaxTurnsAfterWrapUpBudgetSequenceAgent(
	t *testing.T,
	name string,
	subAgents []adkagent.Agent,
	events func() []*session.Event,
) adkagent.Agent {
	t.Helper()

	agent, err := adkagent.New(adkagent.Config{
		Name:        name,
		Description: fmt.Sprintf("%s runEnforcesMaxTurnsAfterWrapUpBudget7 test agent", name),
		SubAgents:   subAgents,
		Run: func(adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				for _, event := range events() {
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

func runEnforcesMaxTurnsAfterWrapUpBudgetFunctionCallEvent(author, name string, args map[string]any) *session.Event {
	return &session.Event{
		Author: author,
		LLMResponse: adkmodel.LLMResponse{
			Content: &genai.Content{Parts: []*genai.Part{
				{FunctionCall: &genai.FunctionCall{Name: name, Args: args}},
			}},
		},
	}
}

func runEnforcesMaxTurnsAfterWrapUpBudgetTextEvent(author, text string, partial bool) *session.Event {
	return &session.Event{
		Author: author,
		LLMResponse: adkmodel.LLMResponse{
			Content: &genai.Content{Role: "model", Parts: []*genai.Part{{Text: text}}},
			Partial: partial,
		},
	}
}
