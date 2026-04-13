package runledger

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/toolchain"
)

func BenchmarkToolProfileGuard_WithCache(b *testing.B) {
	store := NewMemoryStore()
	ctx := context.Background()
	seedActiveRunForBench(store, "run-bench", []string{string(ToolProfileCoding)})

	tool := &agent.Tool{
		Name: "exec",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	guarded := toolchain.Chain(tool, ToolProfileGuard(store))
	runCtx := toolchain.WithAgentName(session.WithRunContext(
		WithSnapshotCache(ctx),
		session.RunContext{SessionType: "workflow", WorkflowID: "wf", RunID: "run-bench"},
	), "operator")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := guarded.Handler(runCtx, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkToolProfileGuard_NoCache(b *testing.B) {
	store := NewMemoryStore()
	ctx := context.Background()
	seedActiveRunForBench(store, "run-bench", []string{string(ToolProfileCoding)})

	tool := &agent.Tool{
		Name: "exec",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	guarded := toolchain.Chain(tool, ToolProfileGuard(store))
	runCtx := toolchain.WithAgentName(session.WithRunContext(
		ctx,
		session.RunContext{SessionType: "workflow", WorkflowID: "wf", RunID: "run-bench"},
	), "operator")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := guarded.Handler(runCtx, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func seedActiveRunForBench(store *MemoryStore, runID string, profiles []string) {
	ctx := context.Background()
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{SessionKey: "session-1", Goal: "bench"}),
	})
	_ = store.AppendJournalEvent(ctx, JournalEvent{
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
	})
	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    EventStepStarted,
		Payload: marshalPayload(StepStartedPayload{StepID: "step-1", OwnerAgent: "operator"}),
	})
}
