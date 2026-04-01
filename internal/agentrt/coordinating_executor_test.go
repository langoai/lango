package agentrt

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adksession "google.golang.org/adk/session"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
)

// mockExecutor implements turnrunner.Executor for testing.
type mockExecutor struct {
	report adk.RunReport
	err    error
	calls  int
}

func (m *mockExecutor) RunStreamingDetailed(
	_ context.Context,
	_, _ string,
	onChunk adk.ChunkCallback,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	m.calls++

	// Apply event hooks (simulate ADK behavior).
	var ro struct{ onEvent func(*adksession.Event) }
	for _, opt := range opts {
		// RunOption is func(*runOptions), which we can't inspect directly.
		// For integration tests, we verify the wrapper behavior instead.
		_ = opt
	}
	_ = ro

	if onChunk != nil && m.report.Response != "" {
		onChunk(m.report.Response)
	}
	return m.report, m.err
}

func TestCoordinatingExecutor_PassthroughOnSuccess(t *testing.T) {
	inner := &mockExecutor{
		report: adk.RunReport{Response: "hello world"},
	}

	ce := NewCoordinatingExecutor(
		inner,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil),
		nil,
	)

	report, err := ce.RunStreamingDetailed(context.Background(), "sess-1", "input", nil)
	require.NoError(t, err)
	assert.Equal(t, "hello world", report.Response)
	assert.Equal(t, 1, inner.calls)
}

func TestCoordinatingExecutor_RecoveryRetryOnTransient(t *testing.T) {
	callCount := 0
	inner := &mockExecutor{}

	// First call fails with transient error, second succeeds.
	origRunStreamingDetailed := inner.RunStreamingDetailed
	_ = origRunStreamingDetailed

	// Use a custom mock that changes behavior on second call.
	type multiCallExecutor struct {
		reports []adk.RunReport
		errs    []error
	}
	mce := &multiCallExecutor{
		reports: []adk.RunReport{
			{Response: ""},
			{Response: "recovered"},
		},
		errs: []error{
			&adk.AgentError{Code: adk.ErrModelError, CauseClass: "provider_transient", Message: "503"},
			nil,
		},
	}

	var executor mockCallExecutor = func(ctx context.Context, sid, input string, onChunk adk.ChunkCallback, opts ...adk.RunOption) (adk.RunReport, error) {
		idx := callCount
		callCount++
		if idx >= len(mce.reports) {
			idx = len(mce.reports) - 1
		}
		return mce.reports[idx], mce.errs[idx]
	}

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil),
		nil,
	)

	report, err := ce.RunStreamingDetailed(context.Background(), "sess-1", "input", nil)
	require.NoError(t, err)
	assert.Equal(t, "recovered", report.Response)
	assert.Equal(t, 2, callCount)
}

func TestCoordinatingExecutor_RecoveryEscalateOnTimeout(t *testing.T) {
	timeoutErr := &adk.AgentError{Code: adk.ErrTimeout, Message: "deadline exceeded"}
	executor := mockCallExecutor(func(_ context.Context, _, _ string, _ adk.ChunkCallback, _ ...adk.RunOption) (adk.RunReport, error) {
		return adk.RunReport{}, timeoutErr
	})

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil),
		nil,
	)

	_, err := ce.RunStreamingDetailed(context.Background(), "sess-1", "input", nil)
	require.Error(t, err)
	var agentErr *adk.AgentError
	require.True(t, errors.As(err, &agentErr))
	assert.Equal(t, adk.ErrTimeout, agentErr.Code)
}

func TestCoordinatingExecutor_EventBusPublish(t *testing.T) {
	bus := eventbus.New()
	var recoveryEvents []RecoveryEvent
	eventbus.SubscribeTyped(bus, func(e RecoveryEvent) {
		recoveryEvents = append(recoveryEvents, e)
	})

	toolChurnErr := &adk.AgentError{Code: adk.ErrToolChurn, Message: "stuck"}
	callCount := 0
	executor := mockCallExecutor(func(_ context.Context, _, _ string, _ adk.ChunkCallback, _ ...adk.RunOption) (adk.RunReport, error) {
		callCount++
		if callCount <= 2 {
			return adk.RunReport{}, toolChurnErr
		}
		return adk.RunReport{Response: "ok"}, nil
	})

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 3}, nil),
		bus,
	)

	report, err := ce.RunStreamingDetailed(context.Background(), "sess-1", "input", nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", report.Response)
	assert.Equal(t, 3, callCount)
	assert.Len(t, recoveryEvents, 2) // two recovery events published
}

func TestCoordinatingExecutor_ToolErrorAfterSpecialistUsesRerouteHint(t *testing.T) {
	bus := eventbus.New()
	var recoveryEvents []RecoveryEvent
	eventbus.SubscribeTyped(bus, func(e RecoveryEvent) {
		recoveryEvents = append(recoveryEvents, e)
	})

	var (
		inputs     []string
		recoveries []adk.RecoveryInfo
		callCount  int
	)
	executor := mockCallExecutor(func(_ context.Context, _, input string, _ adk.ChunkCallback, opts ...adk.RunOption) (adk.RunReport, error) {
		inputs = append(inputs, input)
		hooks := adk.ResolveRunHooks(opts...)
		if callCount == 0 {
			callCount++
			require.NotNil(t, hooks.OnEvent)
			hooks.OnEvent(&adksession.Event{
				Author: "lango-orchestrator",
				Actions: adksession.EventActions{
					TransferToAgent: "vault",
				},
			})
			return adk.RunReport{}, &adk.AgentError{Code: adk.ErrToolError, CauseClass: "unknown_tool_error", Message: "tool failed"}
		}
		callCount++
		return adk.RunReport{Response: "rerouted"}, nil
	})

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil),
		bus,
	)

	report, err := ce.RunStreamingDetailed(
		context.Background(),
		"sess-1",
		"check wallet balance",
		nil,
		adk.WithOnRecovery(func(info adk.RecoveryInfo) {
			recoveries = append(recoveries, info)
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, "rerouted", report.Response)
	require.Len(t, inputs, 2)
	assert.NotEqual(t, inputs[0], inputs[1], "second attempt should use reroute hint input")
	assert.Contains(t, inputs[1], "vault")
	require.Len(t, recoveryEvents, 1)
	assert.Equal(t, "vault", recoveryEvents[0].AgentName)
	assert.Equal(t, RecoveryRetryWithHint, recoveryEvents[0].Action)
	require.Len(t, recoveries, 1)
	assert.Equal(t, "vault", recoveries[0].AgentName)
	assert.Equal(t, "retry_with_hint", recoveries[0].Action)
}

func TestCoordinatingExecutor_ToolErrorBeforeSpecialistRetriesSameInput(t *testing.T) {
	var (
		inputs    []string
		callCount int
	)
	executor := mockCallExecutor(func(_ context.Context, _, input string, _ adk.ChunkCallback, _ ...adk.RunOption) (adk.RunReport, error) {
		inputs = append(inputs, input)
		if callCount == 0 {
			callCount++
			return adk.RunReport{}, &adk.AgentError{Code: adk.ErrToolError, CauseClass: "unknown_tool_error", Message: "tool failed"}
		}
		callCount++
		return adk.RunReport{Response: "retried"}, nil
	})

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil),
		nil,
	)

	report, err := ce.RunStreamingDetailed(context.Background(), "sess-1", "check wallet balance", nil)
	require.NoError(t, err)
	assert.Equal(t, "retried", report.Response)
	require.Len(t, inputs, 2)
	assert.Equal(t, inputs[0], inputs[1], "pre-specialist retries should keep same input")
}

func TestCoordinatingExecutor_PreservesExistingOnEventHook(t *testing.T) {
	var existingHookCalls int32

	executor := mockCallExecutor(func(_ context.Context, _, _ string, _ adk.ChunkCallback, opts ...adk.RunOption) (adk.RunReport, error) {
		hooks := adk.ResolveRunHooks(opts...)
		require.NotNil(t, hooks.OnEvent)
		hooks.OnEvent(&adksession.Event{
			Author: "lango-orchestrator",
			Actions: adksession.EventActions{
				TransferToAgent: "operator",
			},
		})
		return adk.RunReport{Response: "ok"}, nil
	})

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil),
		nil,
	)

	_, err := ce.RunStreamingDetailed(
		context.Background(),
		"sess-1",
		"input",
		nil,
		adk.WithOnEvent(func(_ *adksession.Event) {
			atomic.AddInt32(&existingHookCalls, 1)
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&existingHookCalls))
}

func TestCoordinatingExecutor_RecordOutcomeUsesSpecialistNotReturnTarget(t *testing.T) {
	executor := mockCallExecutor(func(_ context.Context, _, _ string, _ adk.ChunkCallback, opts ...adk.RunOption) (adk.RunReport, error) {
		hooks := adk.ResolveRunHooks(opts...)
		require.NotNil(t, hooks.OnEvent)

		hooks.OnEvent(&adksession.Event{
			Author: "lango-orchestrator",
			Actions: adksession.EventActions{
				TransferToAgent: "operator",
			},
		})
		hooks.OnEvent(&adksession.Event{
			Author: "operator",
			Actions: adksession.EventActions{
				TransferToAgent: "lango-orchestrator",
			},
		})

		return adk.RunReport{}, &adk.AgentError{Code: adk.ErrToolChurn, Message: "loop"}
	})

	guard := NewDelegationGuard(config.CircuitBreakerCfg{
		FailureThreshold: 1,
		ResetTimeout:     30 * time.Second,
	}, nil)

	ce := NewCoordinatingExecutor(
		executor,
		guard,
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		nil,
		nil,
	)

	_, err := ce.RunStreamingDetailed(context.Background(), "sess-1", "input", nil)
	require.Error(t, err)
	assert.Equal(t, CircuitOpen, guard.State("operator"))
	assert.Equal(t, CircuitClosed, guard.State("lango-orchestrator"))
}

func TestCoordinatingExecutor_ContextCancelInterruptsBackoffSleep(t *testing.T) {
	// First call returns a transient error (triggers RecoveryRetry with backoff).
	// We cancel the context during the backoff sleep and expect a fast return.
	callCount := 0
	executor := mockCallExecutor(func(_ context.Context, _, _ string, _ adk.ChunkCallback, _ ...adk.RunOption) (adk.RunReport, error) {
		callCount++
		return adk.RunReport{}, &adk.AgentError{
			Code:       adk.ErrModelError,
			CauseClass: adk.CauseProviderTransient,
			Message:    "503 service unavailable",
		}
	})

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 3}, nil),
		nil,
	)

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after a short delay — well before the 1s backoff completes.
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err := ce.RunStreamingDetailed(ctx, "sess-cancel", "input", nil)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	// The first call returns immediately, then backoff (1s) is interrupted by cancel (50ms).
	// Total should be well under the 1s backoff. Allow up to 500ms for CI slack.
	assert.Less(t, elapsed, 500*time.Millisecond, "context cancel should interrupt backoff sleep promptly")
	assert.Equal(t, 1, callCount, "only one call should have been made before backoff was interrupted")
}

func TestCoordinatingExecutor_RecoveryDecisionEventPublished(t *testing.T) {
	bus := eventbus.New()

	var decisionEvents []RecoveryDecisionEvent
	eventbus.SubscribeTyped(bus, func(e RecoveryDecisionEvent) {
		decisionEvents = append(decisionEvents, e)
	})

	callCount := 0
	executor := mockCallExecutor(func(_ context.Context, _, _ string, _ adk.ChunkCallback, _ ...adk.RunOption) (adk.RunReport, error) {
		callCount++
		if callCount <= 1 {
			return adk.RunReport{}, &adk.AgentError{
				Code:       adk.ErrModelError,
				CauseClass: adk.CauseProviderTransient,
				Message:    "503 transient",
			}
		}
		return adk.RunReport{Response: "ok"}, nil
	})

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 3}, nil),
		bus,
	)

	report, err := ce.RunStreamingDetailed(context.Background(), "sess-decision", "input", nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", report.Response)
	assert.Equal(t, 2, callCount)

	// Verify RecoveryDecisionEvent was published with correct fields.
	require.Len(t, decisionEvents, 1, "exactly one RecoveryDecisionEvent should be published")
	evt := decisionEvents[0]
	assert.Equal(t, string(CauseTransient), evt.CauseClass)
	assert.Equal(t, "retry", evt.Action)
	assert.Equal(t, 0, evt.Attempt, "first recovery attempt should be attempt 0")
	assert.Equal(t, ComputeBackoff(0), evt.Backoff)
	assert.Equal(t, "sess-decision", evt.SessionKey)
}

// mockCallExecutor is a function-based turnrunner.Executor for testing.
type mockCallExecutor func(ctx context.Context, sessionID, input string, onChunk adk.ChunkCallback, opts ...adk.RunOption) (adk.RunReport, error)

func (f mockCallExecutor) RunStreamingDetailed(ctx context.Context, sessionID, input string, onChunk adk.ChunkCallback, opts ...adk.RunOption) (adk.RunReport, error) {
	return f(ctx, sessionID, input, onChunk, opts...)
}
