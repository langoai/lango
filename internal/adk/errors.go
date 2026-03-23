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

const (
	CauseApprovalDenied           = "approval_denied"
	CauseApprovalTimeout          = "approval_timeout"
	CauseApprovalUnavailable      = "approval_unavailable"
	CauseToolNotFound             = "tool_not_found"
	CauseFunctionCallValidation   = "function_call_validation"
	CauseOrchestratorDirectTool   = "orchestrator_direct_tool_call"
	CauseUnknownToolError         = "unknown_tool_error"
	CauseProviderRateLimit        = "provider_rate_limit"
	CauseProviderTransient        = "provider_transient"
	CauseThoughtSignatureMissing  = "thought_signature_missing"
	CauseTimeoutIdle              = "timeout_idle"
	CauseTimeoutHard              = "timeout_hard"
	CauseRepeatedCallSignature    = "repeated_call_signature"
	CauseTurnLimitExceeded        = "turn_limit_exceeded"
	CauseEmptyAfterToolUse        = "empty_after_tool_use"
	CauseInternalRuntimeError     = "internal_runtime_error"
)

// FailureClassification is the operator-facing classification for a terminal failure.
type FailureClassification struct {
	Code            ErrorCode
	CauseClass      string
	CauseDetail     string
	OperatorSummary string
}

// AgentError is a structured error type that preserves partial results
// accumulated before the failure, along with classification metadata.
type AgentError struct {
	Code            ErrorCode
	Message         string        // internal message
	Cause           error         // underlying error
	Partial         string        // accumulated text before failure
	Elapsed         time.Duration // time spent before failure
	CauseClass      string
	CauseDetail     string
	OperatorSummary string
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

// DiagnosticSummary returns an operator-facing summary with cause class and detail.
func (e *AgentError) DiagnosticSummary() string {
	if strings.TrimSpace(e.OperatorSummary) != "" {
		return e.OperatorSummary
	}
	parts := []string{fmt.Sprintf("[%s]", e.Code)}
	if strings.TrimSpace(e.CauseClass) != "" {
		parts = append(parts, e.CauseClass)
	}
	if strings.TrimSpace(e.CauseDetail) != "" {
		parts = append(parts, e.CauseDetail)
	}
	if len(parts) == 1 && strings.TrimSpace(e.Message) != "" {
		parts = append(parts, e.Message)
	}
	return strings.Join(parts, " ")
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

// classifyError determines the operator-facing failure classification for a given error.
func classifyError(err error) FailureClassification {
	if err == nil {
		return FailureClassification{
			Code:            ErrInternal,
			CauseClass:      CauseInternalRuntimeError,
			OperatorSummary: "[E005] internal_runtime_error",
		}
	}

	// Context-based classification
	if err == context.DeadlineExceeded || err == context.Canceled {
		return FailureClassification{
			Code:            ErrTimeout,
			CauseClass:      CauseTimeoutHard,
			CauseDetail:     err.Error(),
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrTimeout, CauseTimeoutHard),
		}
	}
	// Unwrap to check wrapped context errors
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return FailureClassification{
			Code:            ErrTimeout,
			CauseClass:      CauseTimeoutHard,
			CauseDetail:     err.Error(),
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrTimeout, CauseTimeoutHard),
		}
	}

	if errors.Is(err, approval.ErrDenied) {
		return FailureClassification{
			Code:            ErrToolError,
			CauseClass:      CauseApprovalDenied,
			CauseDetail:     err.Error(),
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrToolError, CauseApprovalDenied),
		}
	}
	if errors.Is(err, approval.ErrTimeout) {
		return FailureClassification{
			Code:            ErrToolError,
			CauseClass:      CauseApprovalTimeout,
			CauseDetail:     err.Error(),
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrToolError, CauseApprovalTimeout),
		}
	}
	if errors.Is(err, approval.ErrUnavailable) {
		return FailureClassification{
			Code:            ErrToolError,
			CauseClass:      CauseApprovalUnavailable,
			CauseDetail:     err.Error(),
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrToolError, CauseApprovalUnavailable),
		}
	}

	msg := err.Error()

	// Turn limit
	if strings.Contains(msg, "maximum turn limit") || strings.Contains(msg, "max turns exceeded") {
		return FailureClassification{
			Code:            ErrTurnLimit,
			CauseClass:      CauseTurnLimitExceeded,
			CauseDetail:     msg,
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrTurnLimit, CauseTurnLimitExceeded),
		}
	}

	// Tool churn (consecutive same-tool loop detected by Run())
	if strings.Contains(msg, "consecutively, forcing stop") {
		return FailureClassification{
			Code:            ErrToolChurn,
			CauseClass:      CauseRepeatedCallSignature,
			CauseDetail:     msg,
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrToolChurn, CauseRepeatedCallSignature),
		}
	}

	// thought_signature errors — Gemini API rejects replayed thought data.
	if strings.Contains(msg, "thought_signature") || strings.Contains(msg, "thoughtSignature") {
		return FailureClassification{
			Code:            ErrModelError,
			CauseClass:      CauseThoughtSignatureMissing,
			CauseDetail:     msg,
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrModelError, CauseThoughtSignatureMissing),
		}
	}

	if strings.Contains(msg, "tool not found") || strings.Contains(msg, "failed to find tool") {
		return FailureClassification{
			Code:            ErrToolError,
			CauseClass:      CauseToolNotFound,
			CauseDetail:     msg,
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrToolError, CauseToolNotFound),
		}
	}

	if strings.Contains(msg, "orchestrator emitted direct tool call") {
		return FailureClassification{
			Code:            ErrToolError,
			CauseClass:      CauseOrchestratorDirectTool,
			CauseDetail:     msg,
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrToolError, CauseOrchestratorDirectTool),
		}
	}

	if strings.Contains(msg, "function call") || strings.Contains(msg, "failed to parse tool output") {
		return FailureClassification{
			Code:            ErrToolError,
			CauseClass:      CauseFunctionCallValidation,
			CauseDetail:     msg,
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrToolError, CauseFunctionCallValidation),
		}
	}

	if strings.Contains(msg, "429") || strings.Contains(msg, "rate limit") {
		return FailureClassification{
			Code:            ErrModelError,
			CauseClass:      CauseProviderRateLimit,
			CauseDetail:     msg,
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrModelError, CauseProviderRateLimit),
		}
	}
	if strings.Contains(msg, "500") || strings.Contains(msg, "503") || strings.Contains(msg, "transient") {
		return FailureClassification{
			Code:            ErrModelError,
			CauseClass:      CauseProviderTransient,
			CauseDetail:     msg,
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrModelError, CauseProviderTransient),
		}
	}
	if strings.Contains(msg, "tool") {
		return FailureClassification{
			Code:            ErrToolError,
			CauseClass:      CauseUnknownToolError,
			CauseDetail:     msg,
			OperatorSummary: fmt.Sprintf("[%s] %s", ErrToolError, CauseUnknownToolError),
		}
	}

	return FailureClassification{
		Code:            ErrInternal,
		CauseClass:      CauseInternalRuntimeError,
		CauseDetail:     msg,
		OperatorSummary: fmt.Sprintf("[%s] %s", ErrInternal, CauseInternalRuntimeError),
	}
}
