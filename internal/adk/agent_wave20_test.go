package adk

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func TestWave20RunOptionHelpersChainAndResolve(t *testing.T) {
	t.Parallel()

	var calls []string
	hooks := ResolveRunHooks(
		WithOnActivity(func() { calls = append(calls, "activity") }),
		WithOnRecovery(func(info RecoveryInfo) {
			calls = append(calls, "recovery:"+info.Action)
		}),
		WithOnEvent(func(*session.Event) { calls = append(calls, "first-event") }),
		ChainOnEvent(func(*session.Event) { calls = append(calls, "second-event") }),
		WithOnFinish(func() { calls = append(calls, "finish") }),
	)

	require.NotNil(t, hooks.OnActivity)
	require.NotNil(t, hooks.OnRecovery)
	require.NotNil(t, hooks.OnEvent)
	require.NotNil(t, hooks.OnFinish)

	hooks.OnActivity()
	hooks.OnRecovery(RecoveryInfo{Action: "retry"})
	hooks.OnEvent(&session.Event{})
	hooks.OnFinish()

	assert.Equal(t, []string{
		"activity",
		"recovery:retry",
		"first-event",
		"second-event",
		"finish",
	}, calls)
}

func TestWave20CollectionHelpersCoverNilEmptyAndMultipleParts(t *testing.T) {
	t.Parallel()

	assert.False(t, hasText(&session.Event{}))
	assert.Zero(t, countTextParts(&session.Event{}))
	assert.Zero(t, countFunctionCalls(&session.Event{}))
	assert.Zero(t, countFunctionResponses(&session.Event{}))
	assert.Empty(t, extractPrimaryToolName(&session.Event{}))
	assert.Empty(t, extractPrimaryToolSignature(&session.Event{}))

	event := &session.Event{
		Author: "operator",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{
				Parts: []*genai.Part{
					{Text: "visible"},
					{Text: ""},
					{FunctionCall: &genai.FunctionCall{
						Name: "search",
						Args: map[string]any{"query": "lango"},
					}},
					{FunctionCall: &genai.FunctionCall{Name: "read_file"}},
					{FunctionResponse: &genai.FunctionResponse{
						ID:       "call-search",
						Name:     "search",
						Response: map[string]any{"result": "ok"},
					}},
				},
			},
		},
	}

	assert.True(t, hasText(event))
	assert.Equal(t, 1, countTextParts(event))
	assert.Equal(t, 2, countFunctionCalls(event))
	assert.Equal(t, 1, countFunctionResponses(event))
	assert.Equal(t, "search", extractPrimaryToolName(event))
	assert.Equal(t, `operator|search|{"query":"lango"}`, extractPrimaryToolSignature(event))
}

func TestWave20RecordDiagnosticsCountsDelegationsRootToolsAndResults(t *testing.T) {
	t.Parallel()

	var diagnostics RunDiagnostics
	recordDiagnostics(nil, "root", true, &session.Event{})
	recordDiagnostics(&diagnostics, "root", true, nil)

	recordDiagnostics(&diagnostics, "root", true, &session.Event{
		Author: "root",
		Actions: session.EventActions{
			TransferToAgent: "operator",
		},
	})
	recordDiagnostics(&diagnostics, "root", true, &session.Event{
		Author: "root",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{Parts: []*genai.Part{
				{Text: "thinking"},
				{FunctionCall: &genai.FunctionCall{Name: "exec"}},
				{FunctionCall: &genai.FunctionCall{Name: "read_file"}},
			}},
		},
	})
	recordDiagnostics(&diagnostics, "root", true, &session.Event{
		Author: "operator",
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{Parts: []*genai.Part{
				{FunctionResponse: &genai.FunctionResponse{Name: "exec"}},
			}},
		},
	})

	assert.Equal(t, RunDiagnostics{
		VisibleTextCount:        1,
		ToolCallCount:           2,
		ToolResultCount:         1,
		DelegationCount:         1,
		DirectRootToolCallCount: 2,
	}, diagnostics)
}

func TestWave20IsolatedAgentSetAndCollectionPolicy(t *testing.T) {
	t.Parallel()

	assert.Nil(t, makeIsolatedAgentSet(nil))
	assert.Nil(t, makeIsolatedAgentSet([]string{}))

	got := makeIsolatedAgentSet([]string{" operator ", "", "vault", "operator"})
	assert.Contains(t, got, "operator")
	assert.Contains(t, got, "vault")
	assert.Len(t, got, 2)

	agent := &Agent{isolatedAgents: got}
	assert.False(t, agent.shouldCollectUserText("operator"))
	assert.False(t, agent.shouldCollectUserText("vault"))
	assert.True(t, agent.shouldCollectUserText("planner"))
	assert.True(t, agent.shouldCollectUserText(""))
	assert.True(t, (&Agent{}).shouldCollectUserText("operator"))
}

func TestWave20DiscardReasonClassifiesRecoverableCleanupReasons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "tool churn",
			err:  errors.New(`call signature "operator|search|{}" repeated 5 times consecutively, forcing stop`),
			want: "loop_detected",
		},
		{
			name: "generic",
			err:  errors.New("provider failed"),
			want: "agent error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, discardReasonForError(tt.err))
		})
	}
}

func TestWave20TruncatePreservesShortTextAndRuneBoundaries(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "short", truncate("short", 10))
	assert.Equal(t, "🙂a...", truncate("🙂abc", 2))
}

func TestWave20RunStreamingDetailedErrorReturnsPartialDiagnosticsAndFinish(t *testing.T) {
	t.Parallel()

	root, err := adkagent.New(adkagent.Config{
		Name:        "wave20-error-root",
		Description: "wave20 error root",
		Run: func(adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				if !yield(&session.Event{
					Author: "wave20-error-root",
					LLMResponse: model.LLMResponse{
						Content: &genai.Content{Parts: []*genai.Part{{Text: "partial"}}},
						Partial: true,
					},
				}, nil) {
					return
				}
				yield(nil, fmt.Errorf("tool not found: missing_tool"))
			}
		},
	})
	require.NoError(t, err)

	agent, err := NewAgentFromADK(root, newMockStore())
	require.NoError(t, err)

	var chunks []string
	var finished bool
	report, err := agent.RunStreamingDetailed(
		context.Background(),
		"wave20-stream-error",
		"fail after partial",
		func(chunk string) { chunks = append(chunks, chunk) },
		WithOnFinish(func() { finished = true }),
	)

	require.Error(t, err)
	var agentErr *AgentError
	require.ErrorAs(t, err, &agentErr)
	assert.Equal(t, ErrToolError, agentErr.Code)
	assert.Equal(t, CauseToolNotFound, agentErr.CauseClass)
	assert.Equal(t, "partial", agentErr.Partial)
	assert.Equal(t, "partial", report.Response)
	assert.Equal(t, []string{"partial"}, chunks)
	assert.Equal(t, 1, report.Diagnostics.VisibleTextCount)
	assert.True(t, finished)
}
