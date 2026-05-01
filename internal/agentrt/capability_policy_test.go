package agentrt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

func TestCapabilityPolicy_DeniesOutsideRoleScope(t *testing.T) {
	policy := CapabilityPolicy{}

	decision := policy.Evaluate(CapabilityRequest{
		RunID:        "run-1",
		TeammateType: "planner",
		ToolName:     "exec",
		ToolSafety:   agent.SafetyLevelDangerous,
	})

	assert.Equal(t, CapabilityDecisionDeny, decision.Kind)
	assert.Contains(t, decision.Reason, `outside role maximum scope`)
	assert.Empty(t, decision.GrantRequestID)
}

func TestCapabilityPolicy_AllowsAlreadyGrantedTool(t *testing.T) {
	policy := CapabilityPolicy{
		ActiveGrants: map[string]map[string]bool{
			"run-1": {
				"exec": true,
			},
		},
	}

	decision := policy.Evaluate(CapabilityRequest{
		RunID:        "run-1",
		TeammateType: "operator",
		ToolName:     "exec",
		ToolSafety:   agent.SafetyLevelDangerous,
	})

	assert.Equal(t, CapabilityDecisionAllow, decision.Kind)
	assert.Equal(t, "existing grant", decision.Reason)
	assert.Empty(t, decision.GrantRequestID)
}

func TestCapabilityPolicy_AllowsAlreadyAllowedByCurrentProjection(t *testing.T) {
	policy := CapabilityPolicy{}

	decision := policy.Evaluate(CapabilityRequest{
		RunID:          "run-1",
		TeammateType:   "operator",
		ToolName:       "exec",
		CurrentAllowed: []string{"exec"},
		ToolSafety:     agent.SafetyLevelDangerous,
	})

	assert.Equal(t, CapabilityDecisionAllow, decision.Kind)
	assert.Equal(t, "already allowed by current projection", decision.Reason)
	assert.Empty(t, decision.GrantRequestID)
}

func TestCapabilityPolicy_RequiresApprovalForDangerousInScopeTool(t *testing.T) {
	policy := CapabilityPolicy{}

	decision := policy.Evaluate(CapabilityRequest{
		RunID:        "run-1",
		TeammateType: "operator",
		ToolName:     "exec",
		ToolSafety:   agent.SafetyLevelDangerous,
	})

	assert.Equal(t, CapabilityDecisionNeedsApproval, decision.Kind)
	assert.Equal(t, "dangerous tool requires approval", decision.Reason)
	assert.Equal(t, "grant-run-1-exec", decision.GrantRequestID)
}

func TestCapabilityPolicy_AllowsSafeInScopeTool(t *testing.T) {
	policy := CapabilityPolicy{}

	decision := policy.Evaluate(CapabilityRequest{
		RunID:        "run-1",
		TeammateType: "operator",
		ToolName:     "fs_read",
		ToolSafety:   agent.SafetyLevelSafe,
	})

	assert.Equal(t, CapabilityDecisionAllow, decision.Kind)
	assert.Equal(t, "safe tool inside role maximum scope", decision.Reason)
	assert.Empty(t, decision.GrantRequestID)
}

func TestCapabilityPolicy_UnknownTeammateTypeIsDenied(t *testing.T) {
	policy := CapabilityPolicy{}

	decision := policy.Evaluate(CapabilityRequest{
		RunID:        "run-1",
		TeammateType: "unknown-agent",
		ToolName:     "exec",
		ToolSafety:   agent.SafetyLevelSafe,
	})

	assert.Equal(t, CapabilityDecisionDeny, decision.Kind)
	assert.Contains(t, decision.Reason, `outside role maximum scope`)
	assert.Empty(t, decision.GrantRequestID)
}

func TestCapabilityPolicy_UnknownTeammateTypeFallsBackToCurrentAllowed(t *testing.T) {
	policy := CapabilityPolicy{}

	decision := policy.Evaluate(CapabilityRequest{
		RunID:          "run-1",
		TeammateType:   "unknown-agent",
		ToolName:       "exec",
		CurrentAllowed: []string{"exec"},
		ToolSafety:     agent.SafetyLevelDangerous,
	})

	assert.Equal(t, CapabilityDecisionAllow, decision.Kind)
	assert.Equal(t, "already allowed by current projection", decision.Reason)
	assert.Empty(t, decision.GrantRequestID)
}

func TestCapabilityPolicy_GrantInitializesActiveGrantsMap(t *testing.T) {
	var policy CapabilityPolicy

	policy.Grant("run-1", "exec")

	require.NotNil(t, policy.ActiveGrants)
	assert.True(t, policy.ActiveGrants["run-1"]["exec"])
}
