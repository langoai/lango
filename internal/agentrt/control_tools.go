package agentrt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolparam"
)

type AgentRunSubmitter interface {
	Submit(context.Context, string, background.Origin) (string, error)
}

// AgentControlPlane provides the dependencies needed by agent lifecycle tools.
// Actual bgManager.Submit integration is deferred to the wiring layer (D4).
type AgentControlPlane struct {
	RunStore          AgentRunStore
	Projection        *AgentRunProjection
	Submitter         AgentRunSubmitter
	CapabilityRuntime *CapabilityRuntime
}

// BuildControlTools creates the agent lifecycle tools: agent_spawn, agent_wait, agent_stop.
func BuildControlTools(cp *AgentControlPlane) []*agent.Tool {
	return []*agent.Tool{
		buildAgentSpawn(cp),
		buildAgentWait(cp),
		buildAgentStop(cp),
	}
}

func buildAgentSpawn(cp *AgentControlPlane) *agent.Tool {
	return &agent.Tool{
		Name:        "agent_spawn",
		Description: "Spawn a child agent to handle a delegated task",
		SafetyLevel: agent.SafetyLevelModerate,
		Parameters: agent.Schema().
			Str("instruction", "The task instruction for the spawned agent (required)").
			Str("agent", "Advisory target specialist name (not guaranteed routing)").
			Str("spawn_reason", "Why the spawned agent is being created").
			Int("timeout", "Timeout in seconds for the spawned agent (default 300)").
			Array("allowed_tools", "string", "Tool names the spawned agent is allowed to use").
			Required("instruction").
			Build(),
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			instruction, err := toolparam.RequireString(params, "instruction")
			if err != nil {
				return nil, err
			}

			requestedAgent := toolparam.OptionalString(params, "agent", "")
			spawnReason := toolparam.OptionalString(params, "spawn_reason", "")
			allowedTools := toolparam.StringSlice(params, "allowed_tools")
			if err := ValidateAllowedToolsForTeammate(requestedAgent, allowedTools); err != nil {
				return nil, err
			}

			// Build the enriched prompt when an agent specialist is requested.
			enrichedInstruction := instruction
			if requestedAgent != "" {
				enrichedInstruction = fmt.Sprintf(
					"[System: This task is best handled by the '%s' specialist.]\n\n%s",
					requestedAgent, instruction,
				)
			}

			agentID, err := generateAgentRunID()
			if err != nil {
				return nil, err
			}

			parentDepth := ctxkeys.SpawnDepthFromContext(ctx)

			run := &AgentRun{
				ID:             agentID,
				RequestedAgent: requestedAgent,
				Instruction:    enrichedInstruction,
				Status:         AgentRunSpawned,
				SpawnDepth:     parentDepth + 1,
				AllowedTools:   allowedTools,
				SpawnReason:    spawnReason,
				CreatedAt:      time.Now(),
			}

			if err := cp.RunStore.Create(run); err != nil {
				return nil, fmt.Errorf("agent spawn: %w", err)
			}

			// Keep legacy pending registration only for non-submitter paths.
			if cp.Projection != nil && cp.Submitter == nil {
				cp.Projection.RegisterPending(agentID)
			}

			if cp.Submitter != nil {
				childCtx := withPendingAgentRunID(ctx, agentID)
				if cp.CapabilityRuntime != nil {
					childCtx = cp.CapabilityRuntime.ContextForRun(childCtx, run)
				}

				parentSession := session.SessionKeyFromContext(ctx)
				submittedID, err := cp.Submitter.Submit(childCtx, enrichedInstruction, background.Origin{
					Channel: "agent_control",
					Session: parentSession,
				})
				if err != nil {
					statusErr := cp.RunStore.UpdateStatus(agentID, AgentRunFailed, "", err.Error())
					if statusErr != nil {
						return nil, fmt.Errorf(
							"agent spawn submit: %w (mark failed: %v)",
							err,
							statusErr,
						)
					}
					return nil, fmt.Errorf("agent spawn submit: %w", err)
				}
				if submittedID != "" && submittedID != agentID {
					err = fmt.Errorf(
						"submitter returned mismatched task ID %q (expected %q)",
						submittedID,
						agentID,
					)
					statusErr := cp.RunStore.UpdateStatus(agentID, AgentRunFailed, "", err.Error())
					if statusErr != nil {
						return nil, fmt.Errorf(
							"agent spawn submit: %w (mark failed: %v)",
							err,
							statusErr,
						)
					}
					return nil, fmt.Errorf("agent spawn submit: %w", err)
				}
			}

			return map[string]interface{}{
				"agent_id":        agentID,
				"status":          string(AgentRunSpawned),
				"requested_agent": requestedAgent,
				"spawn_reason":    spawnReason,
			}, nil
		},
	}
}

func buildAgentWait(cp *AgentControlPlane) *agent.Tool {
	return &agent.Tool{
		Name:        "agent_wait",
		Description: "Wait for a spawned agent to reach a terminal state",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: agent.Schema().
			Str("agent_id", "The agent run ID to wait for (required)").
			Int("timeout", "Timeout in seconds (default 300)").
			Required("agent_id").
			Build(),
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			agentID, err := toolparam.RequireString(params, "agent_id")
			if err != nil {
				return nil, err
			}

			timeoutSec := toolparam.OptionalInt(params, "timeout", 300)
			deadline := time.After(time.Duration(timeoutSec) * time.Second)
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()

			for {
				run, err := cp.RunStore.Get(agentID)
				if err != nil {
					return nil, fmt.Errorf("agent wait: %w", err)
				}

				if run.Status.isTerminal() {
					return agentRunResponse(run), nil
				}

				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("agent wait: %w", ctx.Err())
				case <-deadline:
					resp := agentRunResponse(run)
					resp["timeout"] = true
					return resp, nil
				case <-ticker.C:
					// Poll again.
				}
			}
		},
	}
}

func buildAgentStop(cp *AgentControlPlane) *agent.Tool {
	return &agent.Tool{
		Name:        "agent_stop",
		Description: "Stop a spawned agent by cancelling its run",
		SafetyLevel: agent.SafetyLevelSafe,
		Parameters: agent.Schema().
			Str("agent_id", "The agent run ID to stop (required)").
			Required("agent_id").
			Build(),
		Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
			agentID, err := toolparam.RequireString(params, "agent_id")
			if err != nil {
				return nil, err
			}

			if err := cp.RunStore.Cancel(agentID); err != nil {
				return nil, fmt.Errorf("agent stop: %w", err)
			}

			return map[string]interface{}{
				"agent_id": agentID,
				"status":   string(AgentRunCancelled),
			}, nil
		},
	}
}

func agentRunResponse(run *AgentRun) map[string]interface{} {
	resp := map[string]interface{}{
		"agent_id": run.ID,
		"status":   string(run.Status),
	}
	if run.Result != "" {
		resp["result"] = run.Result
	}
	if run.Error != "" {
		resp["error"] = run.Error
	}
	if run.RuntimeCondition != AgentRunConditionNone {
		resp["condition"] = string(run.RuntimeCondition)
	}
	if run.BlockedReason != "" {
		resp["blocked_reason"] = run.BlockedReason
	}
	if run.GrantRequestID != "" {
		resp["grant_request_id"] = run.GrantRequestID
	}
	if run.WaitingOnRunID != "" {
		resp["waiting_on_run_id"] = run.WaitingOnRunID
	}
	if run.RecoveryState != "" {
		resp["recovery_state"] = run.RecoveryState
	}
	return resp
}

// generateAgentRunID creates a random hex ID for an agent run.
func generateAgentRunID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate agent run ID: %w", err)
	}
	return "arun-" + hex.EncodeToString(b), nil
}
