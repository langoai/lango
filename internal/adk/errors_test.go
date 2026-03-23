package adk

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		err  *AgentError
		want string
	}{
		{
			give: "with cause",
			err: &AgentError{
				Code:    ErrTimeout,
				Message: "agent error",
				Cause:   context.DeadlineExceeded,
			},
			want: "[E001] agent error: context deadline exceeded",
		},
		{
			give: "without cause",
			err: &AgentError{
				Code:    ErrModelError,
				Message: "model failed",
			},
			want: "[E002] model failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestAgentError_Unwrap(t *testing.T) {
	t.Parallel()

	cause := fmt.Errorf("root cause")
	err := &AgentError{Code: ErrInternal, Message: "wrapped", Cause: cause}

	assert.True(t, errors.Is(err, cause))
}

func TestAgentError_UserMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    string
		err     *AgentError
		wantSub string
	}{
		{
			give:    "timeout with partial",
			err:     &AgentError{Code: ErrTimeout, Partial: "some text", Elapsed: 30 * time.Second},
			wantSub: "timed out",
		},
		{
			give:    "timeout without partial",
			err:     &AgentError{Code: ErrTimeout, Elapsed: 5 * time.Minute},
			wantSub: "breaking your question",
		},
		{
			give:    "model error",
			err:     &AgentError{Code: ErrModelError},
			wantSub: "AI model",
		},
		{
			give:    "tool error",
			err:     &AgentError{Code: ErrToolError},
			wantSub: "tool execution",
		},
		{
			give:    "turn limit with partial",
			err:     &AgentError{Code: ErrTurnLimit, Partial: "partial"},
			wantSub: "turn limit",
		},
		{
			give:    "internal error",
			err:     &AgentError{Code: ErrInternal},
			wantSub: "internal error",
		},
		{
			give:    "tool churn",
			err:     &AgentError{Code: ErrToolChurn},
			wantSub: "same tool repeatedly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			msg := tt.err.UserMessage()
			assert.Contains(t, msg, tt.wantSub)
			assert.Contains(t, msg, string(tt.err.Code))
		})
	}
}

func TestClassifyError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		err  error
		want ErrorCode
		cause string
	}{
		{
			give: "nil error",
			err:  nil,
			want: ErrInternal,
			cause: CauseInternalRuntimeError,
		},
		{
			give: "deadline exceeded",
			err:  context.DeadlineExceeded,
			want: ErrTimeout,
			cause: CauseTimeoutHard,
		},
		{
			give: "wrapped deadline",
			err:  fmt.Errorf("agent: %w", context.DeadlineExceeded),
			want: ErrTimeout,
			cause: CauseTimeoutHard,
		},
		{
			give: "context canceled",
			err:  context.Canceled,
			want: ErrTimeout,
			cause: CauseTimeoutHard,
		},
		{
			give: "turn limit",
			err:  fmt.Errorf("agent exceeded maximum turn limit (25)"),
			want: ErrTurnLimit,
			cause: CauseTurnLimitExceeded,
		},
		{
			give: "tool error",
			err:  fmt.Errorf("tool execution failed"),
			want: ErrToolError,
			cause: CauseUnknownToolError,
		},
		{
			give: "model error 429",
			err:  fmt.Errorf("429 rate limit exceeded"),
			want: ErrModelError,
			cause: CauseProviderRateLimit,
		},
		{
			give: "thought_signature error",
			err:  fmt.Errorf("invalid thought_signature in request"),
			want: ErrModelError,
			cause: CauseThoughtSignatureMissing,
		},
		{
			give: "thoughtSignature camelCase error",
			err:  fmt.Errorf("field thoughtSignature is not valid"),
			want: ErrModelError,
			cause: CauseThoughtSignatureMissing,
		},
		{
			give: "thought_signature in functionCall parts (Gemini API error)",
			err:  fmt.Errorf("Error 400, Message: Function call is missing a thought_signature in functionCall parts"),
			want: ErrModelError,
			cause: CauseThoughtSignatureMissing,
		},
		{
			give: "tool churn",
			err:  fmt.Errorf(`tool "browser_search" called 5 times consecutively, forcing stop`),
			want: ErrToolChurn,
			cause: CauseRepeatedCallSignature,
		},
		{
			give: "generic error",
			err:  fmt.Errorf("something unknown"),
			want: ErrInternal,
			cause: CauseInternalRuntimeError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := classifyError(tt.err)
			assert.Equal(t, tt.want, got.Code)
			assert.Equal(t, tt.cause, got.CauseClass)
		})
	}
}

func TestAgentError_ErrorsAs(t *testing.T) {
	t.Parallel()

	original := &AgentError{
		Code:    ErrTimeout,
		Message: "timed out",
		Partial: "partial result",
		Cause:   context.DeadlineExceeded,
	}
	wrapped := fmt.Errorf("outer: %w", original)

	var agentErr *AgentError
	require.True(t, errors.As(wrapped, &agentErr))
	assert.Equal(t, ErrTimeout, agentErr.Code)
	assert.Equal(t, "partial result", agentErr.Partial)
}
