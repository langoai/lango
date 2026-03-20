package runledger

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolProfileGuard_AllowsCodingToolsForCodingProfile(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	seedActiveRun(t, store, "run-1", []string{string(ToolProfileCoding)})

	tool := &agent.Tool{
		Name: "exec_command",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	guarded := toolchain.Chain(tool, ToolProfileGuard(store))

	runCtx := toolchain.WithAgentName(session.WithSessionKey(ctx, "workflow:wf:run-1:step-1"), "operator")
	result, err := guarded.Handler(runCtx, nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", result)
}

func TestToolProfileGuard_BlocksDisallowedTools(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	seedActiveRun(t, store, "run-2", []string{string(ToolProfileSupervisor)})

	tool := &agent.Tool{
		Name: "exec_command",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	guarded := toolchain.Chain(tool, ToolProfileGuard(store))

	runCtx := toolchain.WithAgentName(session.WithSessionKey(ctx, "workflow:wf:run-2:step-1"), "operator")
	_, err := guarded.Handler(runCtx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestToolProfileGuard_AllowsRunInspectionToolsForSupervisorProfile(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	seedActiveRun(t, store, "run-3", []string{string(ToolProfileSupervisor)})

	tool := &agent.Tool{
		Name: "run_read",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	guarded := toolchain.Chain(tool, ToolProfileGuard(store))

	runCtx := toolchain.WithAgentName(session.WithSessionKey(ctx, "workflow:wf:run-3:step-1"), "operator")
	result, err := guarded.Handler(runCtx, nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", result)
}

func seedActiveRun(t *testing.T, store *MemoryStore, runID string, profiles []string) {
	t.Helper()
	ctx := context.Background()

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "session-1", Goal: "guard"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID,
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{
				StepID:      "step-1",
				Goal:        "work",
				OwnerAgent:  "operator",
				Status:      StepStatusPending,
				Validator:   ValidatorSpec{Type: ValidatorBuildPass},
				ToolProfile: profiles,
				MaxRetries:  DefaultMaxRetries,
			}},
		}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "step-1", OwnerAgent: "operator"}),
	}))
}
