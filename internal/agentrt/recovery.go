package agentrt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/config"
)

// CauseClass categorizes errors for per-class retry limit decisions.
type CauseClass string

const (
	CauseRateLimit         CauseClass = "rate_limit"
	CauseTransient         CauseClass = "transient"
	CauseMalformedToolCall CauseClass = "malformed_tool_call"
	CauseTimeout           CauseClass = "timeout"
	CauseUnknown           CauseClass = "unknown"
)


const (
	backoffBaseDelay = 1 * time.Second
	backoffMaxDelay  = 30 * time.Second
)

// ComputeBackoff returns an exponential backoff duration for the given attempt.
// The formula is min(baseDelay * 2^attempt, maxBackoff).
func ComputeBackoff(attempt int) time.Duration {
	backoff := time.Duration(1<<uint(attempt)) * backoffBaseDelay
	if backoff > backoffMaxDelay {
		backoff = backoffMaxDelay
	}
	return backoff
}

// classifyForRetry maps an AgentError's CauseClass to a recovery CauseClass
// for per-class retry limit lookups.
func classifyForRetry(agentErr *adk.AgentError) CauseClass {
	if agentErr == nil {
		return CauseUnknown
	}
	switch {
	case agentErr.CauseClass == adk.CauseProviderRateLimit:
		return CauseRateLimit
	case agentErr.CauseClass == adk.CauseProviderTransient:
		return CauseTransient
	case agentErr.CauseClass == adk.CauseFunctionCallValidation:
		return CauseMalformedToolCall
	case agentErr.Code == adk.ErrTimeout || agentErr.Code == adk.ErrIdleTimeout:
		return CauseTimeout
	default:
		return CauseUnknown
	}
}

// retryLimitForClass returns the effective retry limit for a cause class.
// If the class has a specific limit, that limit is used. Otherwise the
// global maxRetries is returned.
func retryLimitForClass(class CauseClass, globalMax int) int {
	switch class {
	case CauseRateLimit:
		return 5
	case CauseTransient:
		return 3
	case CauseMalformedToolCall:
		return 1
	case CauseTimeout:
		return 3
	default:
		return globalMax
	}
}

// RecoveryAction describes the decision made by the recovery policy.
type RecoveryAction int

const (
	RecoveryNone          RecoveryAction = iota + 1
	RecoveryRetry                        // same agent, same input
	RecoveryRetryWithHint                // root orchestrator with "try different specialist" hint
	RecoveryDirectAnswer                 // use partial result to compose response
	RecoveryEscalate                     // return error to caller
)

// RecoveryContext provides information about a failed execution.
type RecoveryContext struct {
	Error            error
	AgentName        string
	PartialResult    string
	RetryCount       int
	SessionID        string
	LearningFix      string // populated by tryLearningFix if ErrorFixProvider returns a fix
	ClassRetryCounts map[CauseClass]int // per-class retry counts across attempts
}

// RecoveryPolicy decides how to handle agent execution failures.
// RecoveryRetryWithHint is NOT per-agent direct execution — it adds a prompt
// hint requesting the root orchestrator to try a different specialist.
type RecoveryPolicy struct {
	maxRetries       int
	errorFixProvider adk.ErrorFixProvider
}

// NewRecoveryPolicy creates a recovery policy from config.
func NewRecoveryPolicy(cfg config.RecoveryCfg, provider adk.ErrorFixProvider) *RecoveryPolicy {
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 2
	}
	return &RecoveryPolicy{
		maxRetries:       maxRetries,
		errorFixProvider: provider,
	}
}

// Decide evaluates a failure and returns the recommended recovery action.
// It checks per-error-class retry limits in addition to the global maxRetries.
func (p *RecoveryPolicy) Decide(ctx context.Context, failure *RecoveryContext) RecoveryAction {
	// Initialize class retry counts if needed.
	if failure.ClassRetryCounts == nil {
		failure.ClassRetryCounts = make(map[CauseClass]int)
	}

	// Compute effective retry limit: class-specific limit may exceed global max.
	effectiveMax := p.maxRetries
	var agentErr *adk.AgentError
	if errors.As(failure.Error, &agentErr) {
		class := classifyForRetry(agentErr)
		if classLimit := retryLimitForClass(class, p.maxRetries); classLimit > effectiveMax {
			effectiveMax = classLimit
		}
	}

	// Check global/effective retry limit.
	if failure.RetryCount >= effectiveMax {
		if failure.PartialResult != "" {
			return RecoveryDirectAnswer
		}
		return RecoveryEscalate
	}

	if !errors.As(failure.Error, &agentErr) {
		// Non-agent error: try learning-based fix if available.
		if p.tryLearningFix(ctx, "", failure.Error, failure) {
			return RecoveryRetry
		}
		return RecoveryEscalate
	}

	// Check per-class retry limit.
	class := classifyForRetry(agentErr)
	classCount := failure.ClassRetryCounts[class]
	classLimit := retryLimitForClass(class, p.maxRetries)
	if classCount >= classLimit {
		if failure.PartialResult != "" {
			return RecoveryDirectAnswer
		}
		return RecoveryEscalate
	}

	// Increment per-class counter.
	failure.ClassRetryCounts[class]++

	switch agentErr.Code {
	case adk.ErrToolChurn:
		return RecoveryRetryWithHint

	case adk.ErrModelError:
		if agentErr.CauseClass == adk.CauseProviderRateLimit || agentErr.CauseClass == adk.CauseProviderTransient {
			return RecoveryRetry
		}
		return RecoveryEscalate

	case adk.ErrTimeout, adk.ErrIdleTimeout:
		return RecoveryEscalate

	case adk.ErrToolError:
		if agentErr.CauseClass == adk.CauseApprovalDenied {
			return RecoveryEscalate
		}
		if agentErr.CauseClass == adk.CauseOrchestratorDirectTool {
			return RecoveryEscalate // same-input retry cannot fix a guard violation
		}
		if agentErr.CauseClass == adk.CauseFunctionCallValidation {
			return RecoveryRetry // malformed tool call — class limit will stop after 1 retry
		}
		if failure.AgentName != "" {
			_ = p.tryLearningFix(ctx, "", failure.Error, failure)
			return RecoveryRetryWithHint
		}
		// Try learning-based error correction before generic retry.
		if p.tryLearningFix(ctx, "", failure.Error, failure) {
			return RecoveryRetry
		}
		return RecoveryRetry

	case adk.ErrEmptyAfterToolUse:
		return RecoveryRetryWithHint

	case adk.ErrTurnLimit:
		if failure.PartialResult != "" {
			return RecoveryDirectAnswer
		}
		return RecoveryEscalate

	default:
		return RecoveryEscalate
	}
}

// tryLearningFix attempts to get a fix from the error fix provider.
// If successful, it stores the fix hint in failure.LearningFix.
func (p *RecoveryPolicy) tryLearningFix(ctx context.Context, toolName string, err error, failure *RecoveryContext) bool {
	if p.errorFixProvider == nil {
		return false
	}
	fix, ok := p.errorFixProvider.GetFixForError(ctx, toolName, err)
	if !ok {
		return false
	}
	failure.LearningFix = fix
	return true
}

// AddRerouteHint wraps the input with a hint for the root orchestrator
// to try a different specialist agent. Incorporates learning-based fix if available.
func AddRerouteHint(input string, failure RecoveryContext) string {
	fixClause := ""
	if failure.LearningFix != "" {
		fixClause = fmt.Sprintf(" Suggested fix: %s.", failure.LearningFix)
	}
	subject := "previous attempt"
	if failure.AgentName != "" {
		subject = fmt.Sprintf("previous sub-agent (%s)", failure.AgentName)
	}
	return fmt.Sprintf(
		"[System: The %s failed: %v.%s "+
			"Do NOT delegate to the same agent again for this request. "+
			"Re-evaluate and route to a different agent or answer directly. "+
			"Original request: %s]",
		subject, failure.Error, fixClause, input,
	)
}

func (a RecoveryAction) String() string {
	switch a {
	case RecoveryNone:
		return "none"
	case RecoveryRetry:
		return "retry"
	case RecoveryRetryWithHint:
		return "retry_with_hint"
	case RecoveryDirectAnswer:
		return "direct_answer"
	case RecoveryEscalate:
		return "escalate"
	default:
		return "unknown"
	}
}
