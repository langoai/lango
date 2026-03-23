package adk

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrorCode identifies the category of an agent error.
type ErrorCode string

const (
	ErrTimeout     ErrorCode = "E001"
	ErrModelError  ErrorCode = "E002"
	ErrToolError   ErrorCode = "E003"
	ErrTurnLimit   ErrorCode = "E004"
	ErrInternal    ErrorCode = "E005"
	ErrIdleTimeout ErrorCode = "E006"
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
		return fmt.Sprintf("[%s] A tool execution failed. Please try again or rephrase your request.", e.Code)
	case ErrTurnLimit:
		return fmt.Sprintf("[%s] The agent reached its maximum turn limit before producing a final answer. Try a simpler request or increase `agent.maxTurns`.", e.Code)
	case ErrIdleTimeout:
		return fmt.Sprintf("[%s] The request was cancelled due to %s of inactivity. The agent may be stuck — try rephrasing your question.", e.Code, e.Elapsed.Truncate(time.Second))
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

	// Tool errors
	if strings.Contains(msg, "tool") || strings.Contains(msg, "function call") {
		return ErrToolError
	}

	// thought_signature errors — Gemini API rejects replayed thought data.
	// Classify as model error to skip learning-based retry (not a fixable tool error).
	if strings.Contains(msg, "thought_signature") || strings.Contains(msg, "thoughtSignature") {
		return ErrModelError
	}

	// Model errors
	if strings.Contains(msg, "model") || strings.Contains(msg, "429") || strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "500") || strings.Contains(msg, "503") {
		return ErrModelError
	}

	return ErrInternal
}
