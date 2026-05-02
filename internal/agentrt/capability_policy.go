package agentrt

import (
	"fmt"

	"github.com/langoai/lango/internal/agent"
)

type CapabilityDecisionKind string

const (
	CapabilityDecisionAllow         CapabilityDecisionKind = "allow"
	CapabilityDecisionNeedsApproval CapabilityDecisionKind = "needs_approval"
	CapabilityDecisionDeny          CapabilityDecisionKind = "deny"
)

type CapabilityRequest struct {
	RunID          string
	TeammateType   string
	ToolName       string
	CurrentAllowed []string
	ToolSafety     agent.SafetyLevel
}

type CapabilityDecision struct {
	Kind           CapabilityDecisionKind
	Reason         string
	GrantRequestID string
}

type CapabilityPolicy struct {
	ActiveGrants map[string]map[string]bool
}

func grantRequestID(runID, toolName string) string {
	return fmt.Sprintf("grant-%s-%s", runID, toolName)
}

func (p *CapabilityPolicy) Evaluate(req CapabilityRequest) CapabilityDecision {
	if !teammateAllowsTool(req.TeammateType, req.ToolName, req.CurrentAllowed) {
		return CapabilityDecision{
			Kind:   CapabilityDecisionDeny,
			Reason: fmt.Sprintf("tool %q outside role maximum scope for teammate type %q", req.ToolName, req.TeammateType),
		}
	}

	if p.hasGrant(req.RunID, req.ToolName) {
		return CapabilityDecision{
			Kind:   CapabilityDecisionAllow,
			Reason: "existing grant",
		}
	}

	if containsTool(req.CurrentAllowed, req.ToolName) {
		return CapabilityDecision{
			Kind:   CapabilityDecisionAllow,
			Reason: "already allowed by current projection",
		}
	}

	if req.ToolSafety.IsDangerous() {
		return CapabilityDecision{
			Kind:           CapabilityDecisionNeedsApproval,
			Reason:         "dangerous tool requires approval",
			GrantRequestID: grantRequestID(req.RunID, req.ToolName),
		}
	}

	return CapabilityDecision{
		Kind:   CapabilityDecisionAllow,
		Reason: "safe tool inside role maximum scope",
	}
}

func (p *CapabilityPolicy) hasGrant(runID, toolName string) bool {
	if p.ActiveGrants == nil {
		return false
	}

	return p.ActiveGrants[runID][toolName]
}

func (p *CapabilityPolicy) Grant(runID, toolName string) {
	if p.ActiveGrants == nil {
		p.ActiveGrants = make(map[string]map[string]bool)
	}
	if p.ActiveGrants[runID] == nil {
		p.ActiveGrants[runID] = make(map[string]bool)
	}

	p.ActiveGrants[runID][toolName] = true
}

func teammateAllowsTool(teammateType, toolName string, currentAllowed []string) bool {
	teammate, ok := BuiltinTeammateTypes()[teammateType]
	if !ok {
		// Custom or remote teammates do not have a built-in role ceiling.
		// For those paths, treat the current allowlist as the effective ceiling.
		return containsTool(currentAllowed, toolName)
	}

	return teammate.AllowsTool(toolName)
}

func containsTool(toolNames []string, toolName string) bool {
	for _, candidate := range toolNames {
		if candidate == toolName {
			return true
		}
	}

	return false
}
