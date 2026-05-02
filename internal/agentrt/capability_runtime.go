package agentrt

import (
	"context"
	"sync"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/toolchain"
)

const dynamicAllowedToolsBlockReason = "tool restricted by DynamicAllowedTools"

type ToolSafetyLookup func(toolName string) agent.SafetyLevel

type CapabilityRuntime struct {
	Store      AgentRunStore
	Policy     *CapabilityPolicy
	ToolSafety ToolSafetyLookup
	mu         sync.RWMutex
}

func NewCapabilityRuntime(
	store AgentRunStore,
	policy *CapabilityPolicy,
	lookup ToolSafetyLookup,
) *CapabilityRuntime {
	if policy == nil {
		policy = &CapabilityPolicy{}
	}

	return &CapabilityRuntime{
		Store:      store,
		Policy:     policy,
		ToolSafety: lookup,
	}
}

func (r *CapabilityRuntime) BlockedToolSinkForRun(runID string) toolchain.BlockedToolCallSink {
	return func(call toolchain.BlockedToolCall) {
		if err := r.HandleBlockedToolCall(runID, call); err != nil {
			logger().Errorw("blocked tool call handling failed",
				"run_id", runID,
				"tool", call.ToolName,
				"error", err,
			)
		}
	}
}

func (r *CapabilityRuntime) ContextForRun(ctx context.Context, run *AgentRun) context.Context {
	if run == nil {
		return ctx
	}

	ctx = ctxkeys.WithAgentName(ctx, run.RequestedAgent)
	ctx = ctxkeys.WithDynamicAllowedTools(ctx, run.AllowedTools)
	ctx = toolchain.WithBlockedToolCallSink(ctx, r.BlockedToolSinkForRun(run.ID))
	return toolchain.WithToolGrantChecker(ctx, func(toolName string) bool {
		return r.hasGrant(run.ID, toolName)
	})
}

func (r *CapabilityRuntime) HandleBlockedToolCall(runID string, call toolchain.BlockedToolCall) error {
	if r == nil || r.Store == nil || r.Policy == nil {
		return nil
	}
	if call.BlockReason != dynamicAllowedToolsBlockReason {
		return nil
	}

	run, err := r.Store.Get(runID)
	if err != nil {
		return err
	}

	safety := agent.SafetyLevelDangerous
	if r.ToolSafety != nil {
		safety = r.ToolSafety(call.ToolName)
	}

	decision := r.evaluate(CapabilityRequest{
		RunID:          runID,
		TeammateType:   run.RequestedAgent,
		ToolName:       call.ToolName,
		CurrentAllowed: run.AllowedTools,
		ToolSafety:     safety,
	})

	switch decision.Kind {
	case CapabilityDecisionNeedsApproval:
		latest, err := r.Store.Get(runID)
		if err != nil {
			return err
		}
		if r.hasGrant(runID, call.ToolName) || containsTool(latest.AllowedTools, call.ToolName) {
			return nil
		}

		attempt := 1
		if latest.GrantRequestID == decision.GrantRequestID && latest.GrantAttempt > 0 {
			attempt = latest.GrantAttempt + 1
		}

		if err := r.Store.UpdateProjection(runID, RunProjectionPatch{
			ApplyRuntimeCondition: true,
			ApplyBlockedReason:    true,
			ApplyGrantRequestID:   true,
			ApplyGrantAttempt:     true,
			ApplyGrantState:       true,
			RuntimeCondition:      AgentRunConditionBlockedWaitingApproval,
			BlockedReason:         decision.Reason,
			GrantRequestID:        decision.GrantRequestID,
			GrantAttempt:          attempt,
			GrantState:            "pending",
		}); err != nil {
			return err
		}

		latest, err = r.Store.Get(runID)
		if err != nil {
			return err
		}
		if r.hasGrant(runID, call.ToolName) || containsTool(latest.AllowedTools, call.ToolName) {
			if latest.RuntimeCondition == AgentRunConditionBlockedWaitingApproval &&
				latest.GrantRequestID == decision.GrantRequestID &&
				latest.GrantAttempt == attempt &&
				latest.GrantState == "pending" {
				return r.Store.UpdateProjection(runID, RunProjectionPatch{
					ApplyRuntimeCondition: true,
					ApplyBlockedReason:    true,
					ApplyGrantRequestID:   true,
					ApplyGrantAttempt:     true,
					ApplyGrantState:       true,
					RuntimeCondition:      AgentRunConditionNone,
					BlockedReason:         "",
					GrantRequestID:        "",
					GrantAttempt:          0,
					GrantState:            "",
				})
			}

			return nil
		}

		return nil
	case CapabilityDecisionDeny:
		return r.Store.UpdateProjection(runID, RunProjectionPatch{
			ApplyRuntimeCondition: true,
			ApplyBlockedReason:    true,
			ApplyGrantRequestID:   true,
			ApplyGrantAttempt:     true,
			ApplyGrantState:       true,
			RuntimeCondition:      AgentRunConditionNone,
			BlockedReason:         decision.Reason,
			GrantRequestID:        "",
			GrantAttempt:          0,
			GrantState:            "denied",
		})
	default:
		return nil
	}
}

func (r *CapabilityRuntime) ApplyGrant(runID, toolName string) error {
	if r == nil || r.Store == nil || r.Policy == nil {
		return nil
	}

	r.mu.Lock()
	r.Policy.Grant(runID, toolName)
	r.mu.Unlock()

	return r.Store.UpdateProjection(runID, RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		ApplyGrantAttempt:     true,
		ApplyGrantState:       true,
		RuntimeCondition:      AgentRunConditionNone,
		BlockedReason:         "",
		GrantRequestID:        "",
		GrantAttempt:          0,
		GrantState:            "granted",
		AddAllowedTool:        toolName,
	})
}

func (r *CapabilityRuntime) evaluate(req CapabilityRequest) CapabilityDecision {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Policy.Evaluate(req)
}

func (r *CapabilityRuntime) hasGrant(runID, toolName string) bool {
	if r == nil || r.Policy == nil {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Policy.hasGrant(runID, toolName)
}
