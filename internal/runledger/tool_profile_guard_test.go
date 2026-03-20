package runledger

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countingRunLedgerStore struct {
	*MemoryStore
	getCalls atomic.Int32
}

func (s *countingRunLedgerStore) GetRunSnapshot(ctx context.Context, runID string) (*RunSnapshot, error) {
	s.getCalls.Add(1)
	return s.MemoryStore.GetRunSnapshot(ctx, runID)
}

func TestToolProfileGuard_AllowsCodingToolsForCodingProfile(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	seedActiveRun(t, store, "run-1", []string{string(ToolProfileCoding)})

	tool := &agent.Tool{
		Name: "exec",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	guarded := toolchain.Chain(tool, ToolProfileGuard(store))

	runCtx := toolchain.WithAgentName(session.WithRunContext(ctx, session.RunContext{
		SessionType: "workflow",
		WorkflowID:  "wf",
		RunID:       "run-1",
	}), "operator")
	result, err := guarded.Handler(runCtx, nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", result)
}

func TestToolProfileGuard_BlocksDisallowedTools(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	seedActiveRun(t, store, "run-2", []string{string(ToolProfileSupervisor)})

	tool := &agent.Tool{
		Name: "exec",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	guarded := toolchain.Chain(tool, ToolProfileGuard(store))

	runCtx := toolchain.WithAgentName(session.WithRunContext(ctx, session.RunContext{
		SessionType: "workflow",
		WorkflowID:  "wf",
		RunID:       "run-2",
	}), "operator")
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

	runCtx := toolchain.WithAgentName(session.WithRunContext(ctx, session.RunContext{
		SessionType: "workflow",
		WorkflowID:  "wf",
		RunID:       "run-3",
	}), "operator")
	result, err := guarded.Handler(runCtx, nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", result)
}

func TestToolProfileGuard_ExecutionAgentDeniedOrchestratorRunTool(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	seedActiveRun(t, store, "run-4", []string{string(ToolProfileCoding)})

	tool := &agent.Tool{
		Name: "run_create",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	guarded := toolchain.Chain(tool, ToolProfileGuard(store))

	runCtx := toolchain.WithAgentName(session.WithRunContext(ctx, session.RunContext{
		SessionType: "workflow",
		WorkflowID:  "wf",
		RunID:       "run-4",
	}), "operator")
	_, err := guarded.Handler(runCtx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestToolProfileGuard_ExactMatchDoesNotAllowUnrelatedTool(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	seedActiveRun(t, store, "run-5", []string{string(ToolProfileCoding)})

	tool := &agent.Tool{
		Name: "execute_payment",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	guarded := toolchain.Chain(tool, ToolProfileGuard(store))

	runCtx := toolchain.WithAgentName(session.WithRunContext(ctx, session.RunContext{
		SessionType: "workflow",
		WorkflowID:  "wf",
		RunID:       "run-5",
	}), "operator")
	_, err := guarded.Handler(runCtx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestRunIDFromSessionContext_UsesRunContext(t *testing.T) {
	ctx := session.WithRunContext(context.Background(), session.RunContext{
		SessionType: "workflow",
		WorkflowID:  "wf:with:colon",
		RunID:       "run-6",
	})

	assert.Equal(t, "run-6", runIDFromSessionContext(ctx))
}

func TestToolProfileGuard_SnapshotCacheHit(t *testing.T) {
	store := &countingRunLedgerStore{MemoryStore: NewMemoryStore()}
	ctx := context.Background()
	seedActiveRun(t, store.MemoryStore, "run-7", []string{string(ToolProfileCoding)})

	tool := &agent.Tool{
		Name: "exec",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	guarded := toolchain.Chain(tool, ToolProfileGuard(store))

	runCtx := toolchain.WithAgentName(session.WithRunContext(
		WithSnapshotCache(ctx),
		session.RunContext{SessionType: "workflow", WorkflowID: "wf", RunID: "run-7"},
	), "operator")

	_, err := guarded.Handler(runCtx, nil)
	require.NoError(t, err)
	_, err = guarded.Handler(runCtx, nil)
	require.NoError(t, err)
	assert.Equal(t, int32(1), store.getCalls.Load())
}

func TestToolProfileGuard_SnapshotCacheCrossTurnIsolation(t *testing.T) {
	store := &countingRunLedgerStore{MemoryStore: NewMemoryStore()}
	ctx := context.Background()
	seedActiveRun(t, store.MemoryStore, "run-8", []string{string(ToolProfileCoding)})

	tool := &agent.Tool{
		Name: "exec",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	guarded := toolchain.Chain(tool, ToolProfileGuard(store))

	runCtx1 := toolchain.WithAgentName(session.WithRunContext(
		WithSnapshotCache(ctx),
		session.RunContext{SessionType: "workflow", WorkflowID: "wf", RunID: "run-8"},
	), "operator")
	_, err := guarded.Handler(runCtx1, nil)
	require.NoError(t, err)

	runCtx2 := toolchain.WithAgentName(session.WithRunContext(
		WithSnapshotCache(ctx),
		session.RunContext{SessionType: "workflow", WorkflowID: "wf", RunID: "run-8"},
	), "operator")
	_, err = guarded.Handler(runCtx2, nil)
	require.NoError(t, err)

	assert.Equal(t, int32(2), store.getCalls.Load())
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
