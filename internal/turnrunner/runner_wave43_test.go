package turnrunner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/adk/model"
	adksession "google.golang.org/adk/session"
	"google.golang.org/genai"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/approval"
	langosession "github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/tools/browser"
	"github.com/langoai/lango/internal/turntrace"
)

type wave43Executor struct {
	run func(
		ctx context.Context,
		sessionID string,
		input string,
		onChunk adk.ChunkCallback,
		hooks adk.RunHooks,
	) (adk.RunReport, error)
}

func (e *wave43Executor) RunStreamingDetailed(
	ctx context.Context,
	sessionID, input string,
	onChunk adk.ChunkCallback,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	hooks := adk.ResolveRunHooks(opts...)
	defer func() {
		if hooks.OnFinish != nil {
			hooks.OnFinish()
		}
	}()
	return e.run(ctx, sessionID, input, onChunk, hooks)
}

func TestWave43RunnerCreatesTraceWithDefaultEntrypointAndContextState(t *testing.T) {
	traceStore := newMemoryTraceStore()
	var captured struct {
		sessionID string
		input     string
		session   string
		turnID    string
		approval  bool
		browser   bool
	}
	executor := &wave43Executor{
		run: func(ctx context.Context, sessionID, input string, _ adk.ChunkCallback, _ adk.RunHooks) (adk.RunReport, error) {
			captured.sessionID = sessionID
			captured.input = input
			captured.session = langosession.SessionKeyFromContext(ctx)
			captured.turnID = langosession.TurnIDFromContext(ctx)
			captured.approval = approval.TurnApprovalStateFromContext(ctx) != nil
			captured.browser = browser.RequestStateFromContext(ctx) != nil
			return adk.RunReport{Response: "visible answer"}, nil
		},
	}

	runner := New(Config{HardCeiling: time.Second, TraceStore: traceStore}, executor, nil, nil)
	result, err := runner.Run(context.Background(), Request{
		SessionKey: "wave43:session",
		Input:      "hello",
	})

	require.NoError(t, err)
	require.NotEmpty(t, result.TraceID)
	assert.Equal(t, turntrace.OutcomeSuccess, result.Outcome)
	assert.Equal(t, "visible answer", result.ResponseText)
	assert.Equal(t, "wave43:session", captured.sessionID)
	assert.Equal(t, "hello", captured.input)
	assert.Equal(t, "wave43:session", captured.session)
	assert.Equal(t, result.TraceID, captured.turnID)
	assert.True(t, captured.approval)
	assert.True(t, captured.browser)

	trace := traceStore.traces[result.TraceID]
	assert.Equal(t, result.TraceID, trace.TraceID)
	assert.Equal(t, "wave43:session", trace.SessionKey)
	assert.Equal(t, "direct", trace.Entrypoint)
	assert.Equal(t, turntrace.OutcomeSuccess, trace.Outcome)
	assert.Equal(t, "visible answer", trace.Summary)
	require.NotNil(t, trace.EndedAt)
	assert.False(t, trace.StartedAt.IsZero())
}

func TestWave43RunnerEmitsDelegationBudgetToolThinkingAndResultCallbacks(t *testing.T) {
	traceStore := newMemoryTraceStore()
	executor := &wave43Executor{
		run: func(_ context.Context, _ string, _ string, _ adk.ChunkCallback, hooks adk.RunHooks) (adk.RunReport, error) {
			for _, target := range []string{"agent-a", "agent-b", "agent-c", "agent-d"} {
				hooks.OnEvent(&adksession.Event{
					Author: "lango-orchestrator",
					Actions: adksession.EventActions{
						TransferToAgent: target,
					},
				})
			}
			hooks.OnEvent(&adksession.Event{
				Author: "agent-d",
				Actions: adksession.EventActions{
					TransferToAgent: "lango-orchestrator",
				},
			})
			hooks.OnEvent(&adksession.Event{
				Author: "agent-d",
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{Parts: []*genai.Part{
						{Thought: true, Text: "inspect tool receipt"},
						{Text: "visible receipt summary"},
						{
							FunctionCall: &genai.FunctionCall{
								ID:   "call-receipt",
								Name: "receipt_lookup",
								Args: map[string]any{"id": "receipt-1"},
							},
						},
						{
							FunctionResponse: &genai.FunctionResponse{
								ID:       "call-receipt",
								Name:     "receipt_lookup",
								Response: map[string]any{"error": "receipt missing"},
							},
						},
					}},
				},
			})
			return adk.RunReport{Response: "done"}, nil
		},
	}
	runner := New(
		Config{HardCeiling: time.Second, TraceStore: traceStore, DelegationBudgetMax: 5},
		executor,
		nil,
		nil,
	)

	var delegations []string
	var budgetWarnings []string
	var toolCalls []string
	var toolResults []struct {
		callID  string
		name    string
		success bool
		preview string
	}
	var thinking []string

	result, err := runner.Run(context.Background(), Request{
		SessionKey: "wave43:callbacks",
		Input:      "lookup receipt",
		OnDelegation: func(from, to, reason string) {
			delegations = append(delegations, from+"->"+to+":"+reason)
		},
		OnBudgetWarning: func(used, max int) {
			budgetWarnings = append(budgetWarnings, fmt.Sprintf("%d/%d", used, max))
		},
		OnToolCall: func(callID, toolName string, params map[string]any) {
			toolCalls = append(toolCalls, callID+":"+toolName+":"+params["id"].(string))
		},
		OnToolResult: func(callID, toolName string, success bool, _ time.Duration, preview string) {
			toolResults = append(toolResults, struct {
				callID  string
				name    string
				success bool
				preview string
			}{callID: callID, name: toolName, success: success, preview: preview})
		},
		OnThinking: func(agentName string, started bool, summary string) {
			thinking = append(thinking, agentName+":"+boolLabel(started)+":"+summary)
		},
	})

	require.NoError(t, err)
	assert.Equal(t, turntrace.OutcomeSuccess, result.Outcome)
	assert.Equal(t, []string{
		"lango-orchestrator->agent-a:",
		"lango-orchestrator->agent-b:",
		"lango-orchestrator->agent-c:",
		"lango-orchestrator->agent-d:",
		"agent-d->lango-orchestrator:",
	}, delegations)
	assert.Equal(t, []string{"4/5"}, budgetWarnings)
	assert.Equal(t, []string{"call-receipt:receipt_lookup:receipt-1"}, toolCalls)
	require.Len(t, toolResults, 1)
	assert.Equal(t, "call-receipt", toolResults[0].callID)
	assert.Equal(t, "receipt_lookup", toolResults[0].name)
	assert.False(t, toolResults[0].success)
	assert.Contains(t, toolResults[0].preview, "receipt missing")
	assert.Equal(t, []string{
		"agent-d:start:inspect tool receipt",
		"agent-d:stop:inspect tool receipt",
	}, thinking)

	events := traceStore.events[result.TraceID]
	require.Len(t, events, 8)
	assertEventTypes(t, events, []string{
		turntrace.EventDelegation,
		turntrace.EventDelegation,
		turntrace.EventDelegation,
		turntrace.EventDelegation,
		turntrace.EventDelegationReturn,
		turntrace.EventText,
		turntrace.EventToolCall,
		turntrace.EventToolResult,
	})
	assert.Equal(t, "agent-d", events[6].AgentName)
	assert.Equal(t, "receipt_lookup", events[6].ToolName)
	assert.Equal(t, `agent-d|receipt_lookup|{"id":"receipt-1"}`, events[6].CallSignature)
	require.JSONEq(t, `{"id":"call-receipt","response":{"error":"receipt missing"}}`, events[7].PayloadJSON)
}

func TestWave43RunnerAnnotatesTimeoutAndRecordsTerminalError(t *testing.T) {
	traceStore := newMemoryTraceStore()
	sessionStore := &stubSessionStore{}
	executor := &fixtureExecutor{
		err: &adk.AgentError{
			Code:            adk.ErrIdleTimeout,
			Message:         "idle timeout",
			CauseClass:      adk.CauseTimeoutIdle,
			CauseDetail:     "no activity observed",
			OperatorSummary: "[E006] timeout_idle",
		},
	}
	runner := New(
		Config{HardCeiling: time.Second, TraceStore: traceStore},
		executor,
		sessionStore,
		nil,
	)

	result, err := runner.Run(context.Background(), Request{
		SessionKey: "wave43:timeout",
		Input:      "wait",
	})

	require.NoError(t, err)
	assert.Equal(t, turntrace.OutcomeTimeout, result.Outcome)
	assert.Equal(t, string(adk.ErrIdleTimeout), result.ErrorCode)
	assert.Equal(t, adk.CauseTimeoutIdle, result.CauseClass)
	assert.Equal(t, []string{"wave43:timeout"}, sessionStore.annotated)
	assert.Equal(t, turntrace.OutcomeTimeout, traceStore.traces[result.TraceID].Outcome)
	events := traceStore.events[result.TraceID]
	require.Len(t, events, 1)
	assert.Equal(t, turntrace.EventTerminalError, events[0].EventType)
	assert.Contains(t, events[0].PayloadJSON, "timeout_idle")
}

func TestWave43RunnerClassifiesPlainExecutorErrorWithTruncatedDetails(t *testing.T) {
	traceStore := newMemoryTraceStore()
	longDetail := strings.Repeat("plain runtime failure ", 40)
	executor := &fixtureExecutor{err: errors.New(longDetail)}
	runner := New(Config{HardCeiling: time.Second, TraceStore: traceStore}, executor, nil, nil)

	result, err := runner.Run(context.Background(), Request{
		SessionKey: "wave43:error",
		Input:      "explode",
	})

	require.NoError(t, err)
	assert.Equal(t, turntrace.OutcomeInternalError, result.Outcome)
	assert.Equal(t, adk.CauseInternalRuntimeError, result.CauseClass)
	assert.Contains(t, result.ResponseText, "An error occurred:")
	assert.LessOrEqual(t, len(result.CauseDetail), 512)
	assert.LessOrEqual(t, len(result.Summary), 240)
	events := traceStore.events[result.TraceID]
	require.Len(t, events, 1)
	assert.Equal(t, turntrace.EventTerminalError, events[0].EventType)
	assert.Contains(t, events[0].PayloadJSON, adk.CauseInternalRuntimeError)
	assert.Equal(t, turntrace.OutcomeInternalError, traceStore.traces[result.TraceID].Outcome)
}

func TestWave43RunnerUsesGenericFallbackForEmptySuccessfulTurn(t *testing.T) {
	traceStore := newMemoryTraceStore()
	executor := &fixtureExecutor{report: adk.RunReport{}}
	runner := New(Config{HardCeiling: time.Second, TraceStore: traceStore}, executor, nil, nil)

	result, err := runner.Run(context.Background(), Request{
		SessionKey: "wave43:empty-success",
		Input:      "quiet",
	})

	require.NoError(t, err)
	assert.Equal(t, turntrace.OutcomeSuccess, result.Outcome)
	assert.Equal(t, EmptyResponseFallback, result.ResponseText)
	assert.Equal(t, "generic empty-response fallback", result.Summary)
	assert.Empty(t, traceStore.events[result.TraceID])
	assert.Equal(t, "generic empty-response fallback", traceStore.traces[result.TraceID].Summary)
}

func TestWave43RunnerParentCancellationDuringRetryBackoffSkipsSleep(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	executor := &wave43Executor{
		run: func(_ context.Context, _ string, _ string, _ adk.ChunkCallback, _ adk.RunHooks) (adk.RunReport, error) {
			calls++
			cancel()
			return adk.RunReport{}, &adk.AgentError{
				Code:       adk.ErrModelError,
				Message:    "transient",
				CauseClass: adk.CauseProviderTransient,
			}
		},
	}
	runner := New(Config{HardCeiling: time.Second}, executor, nil, nil)

	result, err := runner.Run(ctx, Request{
		SessionKey: "wave43:cancel-retry",
		Input:      "retry",
	})

	require.NoError(t, err)
	assert.Equal(t, 1, calls)
	assert.Equal(t, TurnOutcome("cancelled"), result.Outcome)
	assert.Equal(t, "context_cancelled", result.ErrorCode)
}

func TestWave43RunnerCallbacksRegisteredDuringFireWaitForNextTurn(t *testing.T) {
	executor := &fixtureExecutor{report: adk.RunReport{Response: "ok"}}
	runner := New(Config{HardCeiling: time.Second}, executor, nil, nil)
	var calls []string

	runner.OnTurnComplete(func(sessionKey string) {
		calls = append(calls, "first:"+sessionKey)
		runner.OnTurnComplete(func(sessionKey string) {
			calls = append(calls, "second:"+sessionKey)
		})
	})

	_, err := runner.Run(context.Background(), Request{
		SessionKey: "wave43:first",
		Input:      "one",
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"first:wave43:first"}, calls)

	_, err = runner.Run(context.Background(), Request{
		SessionKey: "wave43:second",
		Input:      "two",
	})
	require.NoError(t, err)
	assert.Equal(t, []string{
		"first:wave43:first",
		"first:wave43:second",
		"second:wave43:second",
	}, calls)
}

func boolLabel(v bool) string {
	if v {
		return "start"
	}
	return "stop"
}

func assertEventTypes(t *testing.T, events []turntrace.Event, want []string) {
	t.Helper()
	got := make([]string, 0, len(events))
	for _, event := range events {
		got = append(got, event.EventType)
	}
	assert.Equal(t, want, got)
	for i, event := range events {
		assert.Equal(t, int64(i+1), event.Seq)
	}
}
