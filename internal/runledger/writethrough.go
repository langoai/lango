package runledger

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/langoai/lango/internal/workflow"
)

// RolloutStage controls how deeply the RunLedger is integrated.
type RolloutStage int

const (
	// StageShadow: journal records only, existing systems unaffected.
	StageShadow RolloutStage = iota
	// StageWriteThrough: all creates/updates go through ledger first, then mirror to projections.
	StageWriteThrough
	// StageAuthoritativeRead: state reads come from ledger snapshots only.
	StageAuthoritativeRead
	// StageProjectionRetired: legacy direct writes removed.
	StageProjectionRetired
)

// RolloutConfig holds the current rollout stage configuration.
type RolloutConfig struct {
	Stage RolloutStage
}

// IsShadow returns true if only shadow journaling is active.
func (c RolloutConfig) IsShadow() bool {
	return c.Stage == StageShadow
}

// IsWriteThrough returns true if write-through is active.
func (c RolloutConfig) IsWriteThrough() bool {
	return c.Stage >= StageWriteThrough
}

// IsAuthoritativeRead returns true if reads should come from ledger.
func (c RolloutConfig) IsAuthoritativeRead() bool {
	return c.Stage >= StageAuthoritativeRead
}

// WorkflowProjectionStore is the subset of workflow state persistence that
// RunLedger write-through needs to mirror workflow state.
type WorkflowProjectionStore interface {
	CreateRun(ctx context.Context, w *workflow.Workflow) (string, error)
	CreateRunWithID(ctx context.Context, runID string, w *workflow.Workflow) error
	UpdateRunStatus(ctx context.Context, runID string, status string) error
	CompleteRun(ctx context.Context, runID string, status string, errMsg string) error
	CreateStepRun(ctx context.Context, runID string, step workflow.Step, renderedPrompt string) error
	UpdateStepStatus(ctx context.Context, runID string, stepID string, status string, result string, errMsg string) error
	GetRunStatus(ctx context.Context, runID string) (*workflow.RunStatus, error)
	GetStepResults(ctx context.Context, runID string) (map[string]string, error)
	ListRuns(ctx context.Context, limit int) ([]workflow.RunStatus, error)
}

// WorkflowWriteThrough creates canonical run IDs in RunLedger and mirrors state
// to the legacy workflow projection store.
type WorkflowWriteThrough struct {
	ledger   RunLedgerStore
	original WorkflowProjectionStore
	enabled  bool
}

// NewWorkflowWriteThrough creates a workflow projection adapter backed by RunLedger.
func NewWorkflowWriteThrough(
	ledger RunLedgerStore,
	original WorkflowProjectionStore,
	cfg RolloutConfig,
) *WorkflowWriteThrough {
	return &WorkflowWriteThrough{
		ledger:   ledger,
		original: original,
		enabled:  cfg.IsWriteThrough(),
	}
}

func (w *WorkflowWriteThrough) CreateRun(ctx context.Context, wf *workflow.Workflow) (string, error) {
	if !w.enabled {
		return w.original.CreateRun(ctx, wf)
	}

	runID := uuid.NewString()
	if err := w.ledger.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID,
		Type:  EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{
			SessionKey:      "",
			OriginalRequest: wf.Description,
			Goal:            wf.Name,
		}),
	}); err != nil {
		return "", fmt.Errorf("append run_created: %w", err)
	}

	steps := make([]Step, 0, len(wf.Steps))
	for i, step := range wf.Steps {
		steps = append(steps, Step{
			StepID:     step.ID,
			Index:      i,
			Goal:       step.Prompt,
			OwnerAgent: step.Agent,
			Status:     StepStatusPending,
			Validator: ValidatorSpec{
				Type: ValidatorOrchestratorApproval,
			},
			MaxRetries: DefaultMaxRetries,
			DependsOn:  step.DependsOn,
		})
	}

	if err := w.ledger.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID,
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps:              steps,
			AcceptanceCriteria: nil,
		}),
	}); err != nil {
		return "", fmt.Errorf("append plan_attached: %w", err)
	}

	if _, err := w.ledger.GetRunSnapshot(ctx, runID); err != nil {
		return "", fmt.Errorf("materialize workflow run snapshot: %w", err)
	}

	if err := w.original.CreateRunWithID(ctx, runID, wf); err != nil {
		return "", fmt.Errorf("create workflow projection: %w", err)
	}
	if err := w.ledger.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID,
		Type:  EventProjectionSynced,
		Payload: marshalPayload(struct {
			Target string `json:"target"`
		}{Target: "workflow"}),
	}); err != nil {
		return "", fmt.Errorf("append projection_synced: %w", err)
	}
	return runID, nil
}

func (w *WorkflowWriteThrough) UpdateRunStatus(ctx context.Context, runID string, status string) error {
	return w.original.UpdateRunStatus(ctx, runID, status)
}

func (w *WorkflowWriteThrough) CompleteRun(ctx context.Context, runID string, status string, errMsg string) error {
	if w.enabled {
		eventType := EventRunCompleted
		payload := marshalPayload(RunCompletedPayload{Summary: "workflow projection completed"})
		if status != "completed" {
			eventType = EventRunFailed
			payload = marshalPayload(RunFailedPayload{Reason: errMsg})
		}
		if err := w.ledger.AppendJournalEvent(ctx, JournalEvent{
			RunID:   runID,
			Type:    eventType,
			Payload: payload,
		}); err != nil {
			return fmt.Errorf("append workflow completion event: %w", err)
		}
	}
	return w.original.CompleteRun(ctx, runID, status, errMsg)
}

func (w *WorkflowWriteThrough) CreateStepRun(ctx context.Context, runID string, step workflow.Step, renderedPrompt string) error {
	return w.original.CreateStepRun(ctx, runID, step, renderedPrompt)
}

func (w *WorkflowWriteThrough) UpdateStepStatus(
	ctx context.Context,
	runID string,
	stepID string,
	status string,
	result string,
	errMsg string,
) error {
	if w.enabled {
		switch status {
		case "running":
			if err := w.ledger.AppendJournalEvent(ctx, JournalEvent{
				RunID:   runID,
				Type:    EventStepStarted,
				Payload: marshalPayload(StepStartedPayload{StepID: stepID}),
			}); err != nil {
				return fmt.Errorf("append workflow step_started: %w", err)
			}
		case "completed":
			if err := w.ledger.AppendJournalEvent(ctx, JournalEvent{
				RunID: runID,
				Type:  EventStepResultProposed,
				Payload: marshalPayload(StepResultProposedPayload{
					StepID: stepID,
					Result: result,
				}),
			}); err != nil {
				return fmt.Errorf("append workflow step_result_proposed: %w", err)
			}
			if err := w.ledger.RecordValidationResult(ctx, runID, stepID, ValidationResult{
				Passed: true,
				Reason: "workflow step completed",
			}); err != nil {
				return fmt.Errorf("record workflow validation pass: %w", err)
			}
		case "failed", "skipped":
			if err := w.ledger.RecordValidationResult(ctx, runID, stepID, ValidationResult{
				Passed: false,
				Reason: errMsg,
			}); err != nil {
				return fmt.Errorf("record workflow validation failure: %w", err)
			}
		}
	}
	return w.original.UpdateStepStatus(ctx, runID, stepID, status, result, errMsg)
}

func (w *WorkflowWriteThrough) GetRunStatus(ctx context.Context, runID string) (*workflow.RunStatus, error) {
	return w.original.GetRunStatus(ctx, runID)
}

func (w *WorkflowWriteThrough) GetStepResults(ctx context.Context, runID string) (map[string]string, error) {
	return w.original.GetStepResults(ctx, runID)
}

func (w *WorkflowWriteThrough) ListRuns(ctx context.Context, limit int) ([]workflow.RunStatus, error) {
	return w.original.ListRuns(ctx, limit)
}
