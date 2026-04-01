package exec

import (
	"context"
	"strings"

	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/session"
)

var policyLogger = logging.SubsystemSugar("tool.exec.policy")

// Verdict represents the policy evaluator's determination.
type Verdict int

const (
	// VerdictAllow indicates the command is clearly safe to proceed.
	VerdictAllow Verdict = iota
	// VerdictObserve indicates the command has opaque elements; allow but flag for monitoring.
	VerdictObserve
	// VerdictBlock indicates the command must be blocked.
	VerdictBlock
)

// String returns the verdict as a lowercase string.
func (v Verdict) String() string {
	switch v {
	case VerdictAllow:
		return "allow"
	case VerdictObserve:
		return "observe"
	case VerdictBlock:
		return "block"
	default:
		return "unknown"
	}
}

// ReasonCode provides machine-readable classification of the policy decision.
type ReasonCode string

const (
	ReasonNone            ReasonCode = ""
	ReasonKillVerb        ReasonCode = "kill_verb"
	ReasonProtectedPath   ReasonCode = "protected_path"
	ReasonLangoCLI        ReasonCode = "lango_cli"
	ReasonSkillImport     ReasonCode = "skill_import"
	ReasonCmdSubstitution ReasonCode = "cmd_substitution"
	ReasonUnsafeVarExpand ReasonCode = "unsafe_var_expansion"
	ReasonEvalVerb        ReasonCode = "eval_verb"
	ReasonEncodedPipe     ReasonCode = "encoded_pipe"
)

// PolicyDecision is the structured result of evaluating a command.
type PolicyDecision struct {
	Verdict   Verdict
	Reason    ReasonCode
	Message   string
	Command   string
	Unwrapped string // inner command after shell wrapper unwrap (empty if no wrapper)
}

// PolicyEvent is the minimal event interface expected by the publisher.
// Matches eventbus.Event without importing the eventbus package.
type PolicyEvent interface {
	EventName() string
}

// EventPublisher is satisfied by *eventbus.Bus. Defined here to avoid
// importing eventbus from the exec package.
type EventPublisher interface {
	Publish(event PolicyEvent)
}

// PolicyEvaluator performs structured command policy evaluation with
// shell wrapper unwrapping and opaque pattern detection.
type PolicyEvaluator struct {
	guard           *CommandGuard
	langoClassifier func(cmd string) (message string, reason ReasonCode)
	bus             EventPublisher // nil = no event publishing
	safeVars        map[string]struct{}
}

// NewPolicyEvaluator creates a PolicyEvaluator.
// safeVars is initialized from internal defaults. The bus may be nil to
// disable event publishing (respecting cfg.Hooks.EventPublishing gate).
func NewPolicyEvaluator(
	guard *CommandGuard,
	classifier func(cmd string) (string, ReasonCode),
	bus EventPublisher,
) *PolicyEvaluator {
	return &PolicyEvaluator{
		guard:           guard,
		langoClassifier: classifier,
		bus:             bus,
		safeVars: map[string]struct{}{
			"HOME": {}, "PATH": {}, "USER": {}, "PWD": {},
			"SHELL": {}, "TERM": {}, "LANG": {}, "LC_ALL": {},
			"LC_CTYPE": {}, "TMPDIR": {},
		},
	}
}

// Evaluate performs the full policy check on a command string.
// It unwraps one level of shell wrapper, applies all existing checks
// to the inner command, and detects opaque patterns.
func (pe *PolicyEvaluator) Evaluate(cmd string) PolicyDecision {
	original := cmd

	// Step 1: Shell wrapper unwrap.
	inner, didUnwrap := unwrapShellWrapper(cmd)
	effectiveCmd := cmd
	unwrapped := ""
	if didUnwrap {
		effectiveCmd = inner
		unwrapped = inner
	}

	// Step 2: Lango CLI / skill-import classification.
	if msg, reason := pe.langoClassifier(effectiveCmd); msg != "" {
		return PolicyDecision{
			Verdict:   VerdictBlock,
			Reason:    reason,
			Message:   msg,
			Command:   original,
			Unwrapped: unwrapped,
		}
	}

	// Step 3: CommandGuard checks (kill verbs, protected paths).
	if blocked, reason := pe.guard.CheckCommand(effectiveCmd); blocked {
		rc := ReasonProtectedPath
		if strings.Contains(reason, "process management") {
			rc = ReasonKillVerb
		}
		return PolicyDecision{
			Verdict:   VerdictBlock,
			Reason:    rc,
			Message:   reason,
			Command:   original,
			Unwrapped: unwrapped,
		}
	}

	// Step 4: Opaque pattern detection.
	if opaqueReason := detectOpaquePattern(effectiveCmd, pe.safeVars); opaqueReason != ReasonNone {
		return PolicyDecision{
			Verdict:   VerdictObserve,
			Reason:    opaqueReason,
			Message:   "command contains opaque pattern (" + string(opaqueReason) + ") — proceeding with monitoring",
			Command:   original,
			Unwrapped: unwrapped,
		}
	}

	// Step 5: Allow.
	return PolicyDecision{
		Verdict:   VerdictAllow,
		Reason:    ReasonNone,
		Command:   original,
		Unwrapped: unwrapped,
	}
}

// publishAndLog logs the policy decision and publishes an event if the bus is non-nil.
// Only called for Observe and Block verdicts.
func (pe *PolicyEvaluator) publishAndLog(d PolicyDecision, ctx context.Context) {
	policyLogger.Warnw("exec policy decision",
		"verdict", d.Verdict.String(),
		"reason", string(d.Reason),
		"command", d.Command,
		"unwrapped", d.Unwrapped,
	)

	if pe.bus == nil {
		return
	}

	pe.bus.Publish(PolicyDecisionData{
		Command:    d.Command,
		Unwrapped:  d.Unwrapped,
		Verdict:    d.Verdict.String(),
		Reason:     string(d.Reason),
		Message:    d.Message,
		SessionKey: session.SessionKeyFromContext(ctx),
		AgentName:  ctxkeys.AgentNameFromContext(ctx),
	})
}

// PolicyDecisionData holds the fields for a policy decision event.
// Exported so that the policyBusAdapter in app can convert it to
// eventbus.PolicyDecisionEvent for SubscribeTyped compatibility.
type PolicyDecisionData struct {
	Command    string
	Unwrapped  string
	Verdict    string
	Reason     string
	Message    string
	SessionKey string
	AgentName  string
}

// EventName implements PolicyEvent.
func (e PolicyDecisionData) EventName() string { return "policy.decision" }
