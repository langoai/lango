package agentrt

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

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
