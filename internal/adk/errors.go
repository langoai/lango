package adk

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/langoai/lango/internal/approval"
)

// ErrorCode identifies the category of an agent error.
type ErrorCode string

const (
	ErrTimeout           ErrorCode = "E001"
	ErrModelError        ErrorCode = "E002"
	ErrToolError         ErrorCode = "E003"
	ErrTurnLimit         ErrorCode = "E004"
	ErrInternal          ErrorCode = "E005"
	ErrIdleTimeout       ErrorCode = "E006"
	ErrToolChurn         ErrorCode = "E007"
	ErrEmptyAfterToolUse ErrorCode = "E008"
)

// AgentError is a structured error type that preserves partial results
// accumulated before the failure, along with classification metadata.
type AgentError struct {
	Code    ErrorCode
	Message string        // internal message
	Cause   error         // underlying error
	Partial string        // accumulated text before failure
	Elapsed time.Duration // time spent before failure
}

func (e *AgentError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AgentError) Unwrap() error {
	return e.Cause
}

// UserMessage returns a user-facing formatted message with error code and hint.
func (e *AgentError) UserMessage() string {
	switch e.Code {
	case ErrTimeout:
		return fmt.Sprintf("[%s] The request timed out after %s. Try breaking your question into smaller parts or increasing the timeout.", e.Code, e.Elapsed.Truncate(time.Second))
	case ErrModelError:
		return fmt.Sprintf("[%s] The AI model returned an error. Please try again.", e.Code)
	case ErrToolError:
		switch {
		case errors.Is(e.Cause, approval.ErrDenied):
			return fmt.Sprintf("[%s] The action was denied by approval. Approve it again if you want to continue.", e.Code)
		case errors.Is(e.Cause, approval.ErrTimeout):
			return fmt.Sprintf("[%s] The approval request expired before confirmation. Try again and approve the action promptly.", e.Code)
		case errors.Is(e.Cause, approval.ErrUnavailable):
			return fmt.Sprintf("[%s] No approval channel was available for this action. Check your active channel or companion connection.", e.Code)
		}
		return fmt.Sprintf("[%s] A tool execution failed. Please try again or rephrase your request.", e.Code)
	case ErrTurnLimit:
		return fmt.Sprintf("[%s] The agent reached its maximum turn limit before producing a final answer. Try a simpler request or increase `agent.maxTurns`.", e.Code)
	case ErrIdleTimeout:
		return fmt.Sprintf("[%s] The request was cancelled due to %s of inactivity. The agent may be stuck — try rephrasing your question.", e.Code, e.Elapsed.Truncate(time.Second))
	case ErrToolChurn:
		return fmt.Sprintf("[%s] The agent got stuck calling the same tool repeatedly and was stopped. Please try rephrasing your request.", e.Code)
	case ErrEmptyAfterToolUse:
		return fmt.Sprintf("[%s] The agent completed tool actions but failed to produce a visible final response. Please try again or ask a narrower follow-up question.", e.Code)
	default:
		return fmt.Sprintf("[%s] An internal error occurred. Please try again.", e.Code)
	}
}

// classifyError determines the ErrorCode for a given error.
func classifyError(err error) ErrorCode {
	if err == nil {
		return ErrInternal
	}

	// Context-based classification
	if err == context.DeadlineExceeded || err == context.Canceled {
		return ErrTimeout
	}
	// Unwrap to check wrapped context errors
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return ErrTimeout
	}

	msg := err.Error()

	// Turn limit
	if strings.Contains(msg, "maximum turn limit") || strings.Contains(msg, "max turns exceeded") {
		return ErrTurnLimit
	}

	// Tool churn (consecutive same-tool loop detected by Run())
	if strings.Contains(msg, "consecutively, forcing stop") {
		return ErrToolChurn
	}

	if errors.Is(err, approval.ErrDenied) || errors.Is(err, approval.ErrTimeout) || errors.Is(err, approval.ErrUnavailable) {
		return ErrToolError
	}

	// thought_signature errors — Gemini API rejects replayed thought data.
	// Must be checked BEFORE "tool"/"function call" keywords because the
	// Gemini error message contains both (e.g., "Function call is missing
	// a thought_signature in functionCall parts").
	if strings.Contains(msg, "thought_signature") || strings.Contains(msg, "thoughtSignature") {
		return ErrModelError
	}

	// Tool errors
	if strings.Contains(msg, "tool") || strings.Contains(msg, "function call") {
		return ErrToolError
	}

	// Model errors
	if strings.Contains(msg, "model") || strings.Contains(msg, "429") || strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "500") || strings.Contains(msg, "503") {
		return ErrModelError
	}

	return ErrInternal
}
