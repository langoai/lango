package agentrt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/toolparam"
)

// AgentControlPlane provides the dependencies needed by agent lifecycle tools.
// Actual bgManager.Submit integration is deferred to the wiring layer (D4).
type AgentControlPlane struct {
	RunStore   AgentRunStore
	Projection *AgentRunProjection
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
			allowedTools := toolparam.StringSlice(params, "allowed_tools")

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
				CreatedAt:      time.Now(),
			}

			if err := cp.RunStore.Create(run); err != nil {
				return nil, fmt.Errorf("agent spawn: %w", err)
			}

			// Register the ID with the projection so that bgManager.Submit (D4)
			// will reuse it instead of generating a new one.
			if cp.Projection != nil {
				cp.Projection.RegisterPending(agentID)
			}

			return map[string]interface{}{
				"agent_id":        agentID,
				"status":          string(AgentRunSpawned),
				"requested_agent": requestedAgent,
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
					return map[string]interface{}{
						"agent_id": run.ID,
						"status":   string(run.Status),
						"result":   run.Result,
						"error":    run.Error,
					}, nil
				}

				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("agent wait: %w", ctx.Err())
				case <-deadline:
					return map[string]interface{}{
						"agent_id": agentID,
						"status":   string(run.Status),
						"timeout":  true,
					}, nil
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

// generateAgentRunID creates a random hex ID for an agent run.
func generateAgentRunID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate agent run ID: %w", err)
	}
	return "arun-" + hex.EncodeToString(b), nil
}
