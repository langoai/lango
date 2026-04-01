package agentrt

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/config"
)

func TestRecoveryPolicy_Decide(t *testing.T) {
	tests := []struct {
		give       string
		failure    RecoveryContext
		wantAction RecoveryAction
	}{
		{
			give: "tool churn → retry with hint",
			failure: RecoveryContext{
				Error: &adk.AgentError{Code: adk.ErrToolChurn, Message: "loop"},
			},
			wantAction: RecoveryRetryWithHint,
		},
		{
			give: "rate limit → retry",
			failure: RecoveryContext{
				Error: &adk.AgentError{Code: adk.ErrModelError, CauseClass: "provider_rate_limit"},
			},
			wantAction: RecoveryRetry,
		},
		{
			give: "transient error → retry",
			failure: RecoveryContext{
				Error: &adk.AgentError{Code: adk.ErrModelError, CauseClass: "provider_transient"},
			},
			wantAction: RecoveryRetry,
		},
		{
			give: "timeout → escalate",
			failure: RecoveryContext{
				Error: &adk.AgentError{Code: adk.ErrTimeout},
			},
			wantAction: RecoveryEscalate,
		},
		{
			give: "approval denied → escalate",
			failure: RecoveryContext{
				Error: &adk.AgentError{Code: adk.ErrToolError, CauseClass: "approval_denied"},
			},
			wantAction: RecoveryEscalate,
		},
		{
			give: "orchestrator direct tool call → escalate",
			failure: RecoveryContext{
				Error: &adk.AgentError{Code: adk.ErrToolError, CauseClass: adk.CauseOrchestratorDirectTool},
			},
			wantAction: RecoveryEscalate,
		},
		{
			give: "tool error → retry",
			failure: RecoveryContext{
				Error: &adk.AgentError{Code: adk.ErrToolError, CauseClass: "unknown_tool_error"},
			},
			wantAction: RecoveryRetry,
		},
		{
			give: "specialist tool error → retry with hint",
			failure: RecoveryContext{
				AgentName: "vault",
				Error:     &adk.AgentError{Code: adk.ErrToolError, CauseClass: "unknown_tool_error"},
			},
			wantAction: RecoveryRetryWithHint,
		},
		{
			give: "empty after tool use → retry with hint",
			failure: RecoveryContext{
				Error: &adk.AgentError{Code: adk.ErrEmptyAfterToolUse},
			},
			wantAction: RecoveryRetryWithHint,
		},
		{
			give: "turn limit with partial → direct answer",
			failure: RecoveryContext{
				Error:         &adk.AgentError{Code: adk.ErrTurnLimit},
				PartialResult: "partial text",
			},
			wantAction: RecoveryDirectAnswer,
		},
		{
			give: "turn limit no partial → escalate",
			failure: RecoveryContext{
				Error: &adk.AgentError{Code: adk.ErrTurnLimit},
			},
			wantAction: RecoveryEscalate,
		},
		{
			give: "non-agent error → escalate",
			failure: RecoveryContext{
				Error: errors.New("random error"),
			},
			wantAction: RecoveryEscalate,
		},
		{
			give: "retries exhausted with partial → direct answer",
			failure: RecoveryContext{
				Error:         &adk.AgentError{Code: adk.ErrToolChurn},
				RetryCount:    5,
				PartialResult: "partial",
			},
			wantAction: RecoveryDirectAnswer,
		},
		{
			give: "retries exhausted no partial → escalate",
			failure: RecoveryContext{
				Error:      &adk.AgentError{Code: adk.ErrToolChurn},
				RetryCount: 5,
			},
			wantAction: RecoveryEscalate,
		},
	}

	policy := NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil)

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			failure := tt.failure
			action := policy.Decide(context.Background(), &failure)
			assert.Equal(t, tt.wantAction, action)
		})
	}
}

func TestAddRerouteHint(t *testing.T) {
	result := AddRerouteHint("hello", RecoveryContext{
		AgentName: "operator",
		Error:     errors.New("tool churn"),
	})
	assert.Contains(t, result, "operator")
	assert.Contains(t, result, "Do NOT delegate to the same agent")
	assert.Contains(t, result, "hello")
}

func TestComputeBackoff(t *testing.T) {
	tests := []struct {
		give        string
		attempt     int
		wantBackoff time.Duration
	}{
		{
			give:        "attempt 0 → 1s",
			attempt:     0,
			wantBackoff: 1 * time.Second,
		},
		{
			give:        "attempt 1 → 2s",
			attempt:     1,
			wantBackoff: 2 * time.Second,
		},
		{
			give:        "attempt 2 → 4s",
			attempt:     2,
			wantBackoff: 4 * time.Second,
		},
		{
			give:        "attempt 3 → 8s",
			attempt:     3,
			wantBackoff: 8 * time.Second,
		},
		{
			give:        "attempt 4 → 16s",
			attempt:     4,
			wantBackoff: 16 * time.Second,
		},
		{
			give:        "attempt 5 → 30s (capped)",
			attempt:     5,
			wantBackoff: 30 * time.Second,
		},
		{
			give:        "attempt 10 → 30s (capped)",
			attempt:     10,
			wantBackoff: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := ComputeBackoff(tt.attempt)
			assert.Equal(t, tt.wantBackoff, got)
		})
	}
}

func TestClassifyForRetry(t *testing.T) {
	tests := []struct {
		give      string
		agentErr  *adk.AgentError
		wantClass CauseClass
	}{
		{
			give:      "nil error → unknown",
			agentErr:  nil,
			wantClass: CauseUnknown,
		},
		{
			give:      "rate limit → CauseRateLimit",
			agentErr:  &adk.AgentError{Code: adk.ErrModelError, CauseClass: adk.CauseProviderRateLimit},
			wantClass: CauseRateLimit,
		},
		{
			give:      "transient → CauseTransient",
			agentErr:  &adk.AgentError{Code: adk.ErrModelError, CauseClass: adk.CauseProviderTransient},
			wantClass: CauseTransient,
		},
		{
			give:      "function call validation → CauseMalformedToolCall",
			agentErr:  &adk.AgentError{Code: adk.ErrToolError, CauseClass: adk.CauseFunctionCallValidation},
			wantClass: CauseMalformedToolCall,
		},
		{
			give:      "timeout → CauseTimeout",
			agentErr:  &adk.AgentError{Code: adk.ErrTimeout},
			wantClass: CauseTimeout,
		},
		{
			give:      "idle timeout → CauseTimeout",
			agentErr:  &adk.AgentError{Code: adk.ErrIdleTimeout},
			wantClass: CauseTimeout,
		},
		{
			give:      "unknown cause → CauseUnknown",
			agentErr:  &adk.AgentError{Code: adk.ErrToolChurn, CauseClass: "something_else"},
			wantClass: CauseUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := classifyForRetry(tt.agentErr)
			assert.Equal(t, tt.wantClass, got)
		})
	}
}

func TestRetryLimitForClass(t *testing.T) {
	tests := []struct {
		give      string
		class     CauseClass
		globalMax int
		wantLimit int
	}{
		{
			give:      "rate limit has class-specific limit",
			class:     CauseRateLimit,
			globalMax: 2,
			wantLimit: 5,
		},
		{
			give:      "transient has class-specific limit",
			class:     CauseTransient,
			globalMax: 2,
			wantLimit: 3,
		},
		{
			give:      "malformed tool call has class-specific limit",
			class:     CauseMalformedToolCall,
			globalMax: 2,
			wantLimit: 1,
		},
		{
			give:      "timeout has class-specific limit",
			class:     CauseTimeout,
			globalMax: 2,
			wantLimit: 3,
		},
		{
			give:      "unknown class uses global max",
			class:     CauseUnknown,
			globalMax: 2,
			wantLimit: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := retryLimitForClass(tt.class, tt.globalMax)
			assert.Equal(t, tt.wantLimit, got)
		})
	}
}

func TestRecoveryPolicy_PerClassRetryLimits(t *testing.T) {
	t.Run("malformed tool call escalates after 1 retry", func(t *testing.T) {
		policy := NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 5}, nil)
		classRetries := make(map[CauseClass]int)

		// First attempt: should retry (class count 0 < limit 1).
		failure := &RecoveryContext{
			Error:            &adk.AgentError{Code: adk.ErrToolError, CauseClass: adk.CauseFunctionCallValidation},
			RetryCount:       0,
			ClassRetryCounts: classRetries,
		}
		action := policy.Decide(context.Background(), failure)
		assert.Equal(t, RecoveryRetry, action)

		// Second attempt: should escalate (class count 1 >= limit 1).
		failure2 := &RecoveryContext{
			Error:            &adk.AgentError{Code: adk.ErrToolError, CauseClass: adk.CauseFunctionCallValidation},
			RetryCount:       1,
			ClassRetryCounts: classRetries, // same map, count incremented by first Decide
		}
		action2 := policy.Decide(context.Background(), failure2)
		assert.Equal(t, RecoveryEscalate, action2)
	})

	t.Run("rate limit allows more retries than global max", func(t *testing.T) {
		policy := NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil)
		classRetries := make(map[CauseClass]int)

		// Retry count 2 (at global max), but rate limit allows 5.
		failure := &RecoveryContext{
			Error:            &adk.AgentError{Code: adk.ErrModelError, CauseClass: adk.CauseProviderRateLimit},
			RetryCount:       2,
			ClassRetryCounts: classRetries,
		}
		action := policy.Decide(context.Background(), failure)
		assert.Equal(t, RecoveryRetry, action)
	})

	t.Run("rate limit eventually exhausts class limit", func(t *testing.T) {
		policy := NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil)
		classRetries := make(map[CauseClass]int)

		// Simulate 5 retries on rate limit.
		for i := 0; i < 5; i++ {
			failure := &RecoveryContext{
				Error:            &adk.AgentError{Code: adk.ErrModelError, CauseClass: adk.CauseProviderRateLimit},
				RetryCount:       i,
				ClassRetryCounts: classRetries,
			}
			action := policy.Decide(context.Background(), failure)
			require.Equal(t, RecoveryRetry, action, "attempt %d should retry", i)
		}

		// Attempt 6: class limit (5) exhausted.
		failure := &RecoveryContext{
			Error:            &adk.AgentError{Code: adk.ErrModelError, CauseClass: adk.CauseProviderRateLimit},
			RetryCount:       5,
			ClassRetryCounts: classRetries,
		}
		action := policy.Decide(context.Background(), failure)
		assert.Equal(t, RecoveryEscalate, action)
	})

	t.Run("class retry counts initialized on nil map", func(t *testing.T) {
		policy := NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil)

		failure := &RecoveryContext{
			Error:      &adk.AgentError{Code: adk.ErrModelError, CauseClass: adk.CauseProviderTransient},
			RetryCount: 0,
			// ClassRetryCounts is nil — should be initialized by Decide.
		}
		action := policy.Decide(context.Background(), failure)
		assert.Equal(t, RecoveryRetry, action)
		assert.NotNil(t, failure.ClassRetryCounts)
	})
}
