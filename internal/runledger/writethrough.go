package runledger

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/workflow"
)

func logProjectionSyncWarning(runID string, err error) {
	if err == nil {
		return
	}
	log.Printf("WARN projection sync %s: %v", runID, err)
}

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
	maxKeep  int
}

// ProjectionDrift describes a mismatch between RunLedger and a projection target.
type ProjectionDrift struct {
	RunID  string
	Target string
	Reason string
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

// WithMaxHistory configures pruning of old terminal runs after completion.
func (w *WorkflowWriteThrough) WithMaxHistory(maxKeep int) *WorkflowWriteThrough {
	w.maxKeep = maxKeep
	return w
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
		logProjectionSyncWarning(runID, appendProjectionSyncEvent(
			ctx, w.ledger, runID, "workflow", "degraded", err,
		))
		return "", fmt.Errorf("create workflow projection: %w", err)
	}
	if err := w.original.UpdateRunStatus(ctx, runID, "running"); err != nil {
		logProjectionSyncWarning(runID, appendProjectionSyncEvent(
			ctx, w.ledger, runID, "workflow", "degraded", err,
		))
		return "", fmt.Errorf("set workflow projection running: %w", err)
	}
	if err := appendProjectionSyncEvent(ctx, w.ledger, runID, "workflow", "synced", nil); err != nil {
		return "", fmt.Errorf("append projection_synced: %w", err)
	}
	return runID, nil
}

func (w *WorkflowWriteThrough) UpdateRunStatus(ctx context.Context, runID string, status string) error {
	if err := w.original.UpdateRunStatus(ctx, runID, status); err != nil {
		if w.enabled {
			logProjectionSyncWarning(runID, appendProjectionSyncEvent(
				ctx, w.ledger, runID, "workflow", "degraded", err,
			))
		}
		return err
	}
	if w.enabled {
		logProjectionSyncWarning(runID, appendProjectionSyncEvent(
			ctx, w.ledger, runID, "workflow", "synced", nil,
		))
	}
	return nil
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
		if w.maxKeep > 0 {
			if err := w.ledger.PruneOldRuns(ctx, w.maxKeep); err != nil {
				return fmt.Errorf("prune old runs: %w", err)
			}
		}
	}
	if err := w.original.CompleteRun(ctx, runID, status, errMsg); err != nil {
		if w.enabled {
			logProjectionSyncWarning(runID, appendProjectionSyncEvent(
				ctx, w.ledger, runID, "workflow", "degraded", err,
			))
		}
		return err
	}
	if w.enabled {
		logProjectionSyncWarning(runID, appendProjectionSyncEvent(
			ctx, w.ledger, runID, "workflow", "synced", nil,
		))
	}
	return nil
}

func (w *WorkflowWriteThrough) CreateStepRun(ctx context.Context, runID string, step workflow.Step, renderedPrompt string) error {
	if err := w.original.CreateStepRun(ctx, runID, step, renderedPrompt); err != nil {
		if w.enabled {
			logProjectionSyncWarning(runID, appendProjectionSyncEvent(
				ctx, w.ledger, runID, "workflow", "degraded", err,
			))
		}
		return err
	}
	if w.enabled {
		logProjectionSyncWarning(runID, appendProjectionSyncEvent(
			ctx, w.ledger, runID, "workflow", "synced", nil,
		))
	}
	return nil
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
	if err := w.original.UpdateStepStatus(ctx, runID, stepID, status, result, errMsg); err != nil {
		if w.enabled {
			logProjectionSyncWarning(runID, appendProjectionSyncEvent(
				ctx, w.ledger, runID, "workflow", "degraded", err,
			))
		}
		return err
	}
	if w.enabled {
		logProjectionSyncWarning(runID, appendProjectionSyncEvent(
			ctx, w.ledger, runID, "workflow", "synced", nil,
		))
	}
	return nil
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

// DetectWorkflowProjectionDrift compares a RunLedger snapshot against the workflow
// projection store and reports the first mismatch found.
func DetectWorkflowProjectionDrift(
	ctx context.Context,
	ledger RunLedgerStore,
	projection WorkflowProjectionStore,
	runID string,
) (*ProjectionDrift, error) {
	snap, err := ledger.GetRunSnapshot(ctx, runID)
	if err != nil {
		return nil, err
	}
	status, err := projection.GetRunStatus(ctx, runID)
	if err != nil {
		return &ProjectionDrift{
			RunID:  runID,
			Target: "workflow",
			Reason: "workflow projection missing",
		}, nil
	}

	if string(snap.Status) != status.Status {
		return &ProjectionDrift{
			RunID:  runID,
			Target: "workflow",
			Reason: fmt.Sprintf("status mismatch: ledger=%s projection=%s", snap.Status, status.Status),
		}, nil
	}
	if len(snap.Steps) != len(status.StepStatuses) {
		return &ProjectionDrift{
			RunID:  runID,
			Target: "workflow",
			Reason: fmt.Sprintf("step count mismatch: ledger=%d projection=%d", len(snap.Steps), len(status.StepStatuses)),
		}, nil
	}

	projectionSteps := make(map[string]string, len(status.StepStatuses))
	for _, step := range status.StepStatuses {
		projectionSteps[step.StepID] = step.Status
	}
	for _, step := range snap.Steps {
		if projectionSteps[step.StepID] != string(step.Status) {
			return &ProjectionDrift{
				RunID:  runID,
				Target: "workflow",
				Reason: fmt.Sprintf("step %s mismatch: ledger=%s projection=%s", step.StepID, step.Status, projectionSteps[step.StepID]),
			}, nil
		}
	}

	return nil, nil
}

// ReplayWorkflowProjection rebuilds the workflow projection from the RunLedger snapshot.
func ReplayWorkflowProjection(
	ctx context.Context,
	ledger RunLedgerStore,
	projection WorkflowProjectionStore,
	runID string,
	wf *workflow.Workflow,
) error {
	snap, err := ledger.GetRunSnapshot(ctx, runID)
	if err != nil {
		return err
	}

	if _, err := projection.GetRunStatus(ctx, runID); err != nil {
		if createErr := projection.CreateRunWithID(ctx, runID, wf); createErr != nil {
			return createErr
		}
	}

	prompts := make(map[string]string, len(wf.Steps))
	agents := make(map[string]string, len(wf.Steps))
	for _, step := range wf.Steps {
		prompts[step.ID] = step.Prompt
		agents[step.ID] = step.Agent
		_ = projection.CreateStepRun(ctx, runID, step, step.Prompt)
	}

	for _, step := range snap.Steps {
		status := mapRunStepStatus(step.Status)
		errMsg := ""
		if step.Status == StepStatusFailed || step.Status == StepStatusInterrupted {
			errMsg = step.Result
			if errMsg == "" {
				errMsg = snap.CurrentBlocker
			}
		}
		if _, ok := prompts[step.StepID]; !ok {
			_ = projection.CreateStepRun(ctx, runID, workflow.Step{
				ID:     step.StepID,
				Agent:  agents[step.StepID],
				Prompt: step.Goal,
			}, step.Goal)
		}
		if err := projection.UpdateStepStatus(ctx, runID, step.StepID, status, step.Result, errMsg); err != nil {
			return err
		}
	}

	if err := projection.UpdateRunStatus(ctx, runID, string(snap.Status)); err != nil {
		return err
	}
	return nil
}

func appendProjectionSyncEvent(
	ctx context.Context,
	ledger RunLedgerStore,
	runID string,
	target string,
	status string,
	syncErr error,
) error {
	payload := ProjectionSyncPayload{
		Target: target,
		Status: status,
	}
	if syncErr != nil {
		payload.Error = syncErr.Error()
	}
	return ledger.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    EventProjectionSynced,
		Payload: marshalPayload(payload),
	})
}

func mapRunStepStatus(status StepStatus) string {
	switch status {
	case StepStatusCompleted:
		return "completed"
	case StepStatusFailed:
		return "failed"
	case StepStatusInProgress:
		return "running"
	case StepStatusInterrupted:
		return "skipped"
	case StepStatusVerifyPending:
		return "running"
	case StepStatusPending:
		return "pending"
	default:
		return strings.ToLower(string(status))
	}
}

// BackgroundWriteThrough creates canonical task IDs in RunLedger and mirrors
// background task lifecycle transitions into the ledger.
type BackgroundWriteThrough struct {
	ledger  RunLedgerStore
	enabled bool
	maxKeep int
}

// NewBackgroundWriteThrough creates a background projection adapter backed by RunLedger.
func NewBackgroundWriteThrough(ledger RunLedgerStore, cfg RolloutConfig) *BackgroundWriteThrough {
	return &BackgroundWriteThrough{
		ledger:  ledger,
		enabled: cfg.IsWriteThrough(),
	}
}

// WithMaxHistory configures pruning of old terminal runs after completion.
func (b *BackgroundWriteThrough) WithMaxHistory(maxKeep int) *BackgroundWriteThrough {
	b.maxKeep = maxKeep
	return b
}

func (b *BackgroundWriteThrough) PrepareTask(
	ctx context.Context,
	prompt string,
	origin background.Origin,
) (string, error) {
	runID := uuid.NewString()
	if !b.enabled {
		return runID, nil
	}

	if err := b.ledger.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID,
		Type:  EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{
			SessionKey:      origin.Session,
			OriginalRequest: prompt,
			Goal:            prompt,
		}),
	}); err != nil {
		return "", fmt.Errorf("append background run_created: %w", err)
	}

	if err := b.ledger.AppendJournalEvent(ctx, JournalEvent{
		RunID: runID,
		Type:  EventPlanAttached,
		Payload: marshalPayload(PlanAttachedPayload{
			Steps: []Step{{
				StepID:     "background-task",
				Index:      0,
				Goal:       prompt,
				OwnerAgent: "automator",
				Status:     StepStatusPending,
				Validator: ValidatorSpec{
					Type:   ValidatorCommandPass,
					Target: "background-task",
				},
				MaxRetries: DefaultMaxRetries,
			}},
		}),
	}); err != nil {
		return "", fmt.Errorf("append background plan_attached: %w", err)
	}

	if _, err := b.ledger.GetRunSnapshot(ctx, runID); err != nil {
		return "", fmt.Errorf("materialize background snapshot: %w", err)
	}

	if err := appendProjectionSyncEvent(ctx, b.ledger, runID, "background", "synced", nil); err != nil {
		return "", fmt.Errorf("append background projection_synced: %w", err)
	}
	return runID, nil
}

func (b *BackgroundWriteThrough) SyncTask(ctx context.Context, snap background.TaskSnapshot) error {
	if !b.enabled {
		return nil
	}

	switch snap.StatusText {
	case "pending":
		return appendProjectionSyncEvent(ctx, b.ledger, snap.ID, "background", "synced", nil)
	case "running":
		if err := b.ledger.AppendJournalEvent(ctx, JournalEvent{
			RunID:   snap.ID,
			Type:    EventStepStarted,
			Payload: marshalPayload(StepStartedPayload{StepID: "background-task", OwnerAgent: "automator"}),
		}); err != nil {
			return err
		}
	case "done":
		if err := b.ledger.AppendJournalEvent(ctx, JournalEvent{
			RunID: snap.ID,
			Type:  EventStepResultProposed,
			Payload: marshalPayload(StepResultProposedPayload{
				StepID: "background-task",
				Result: snap.Result,
			}),
		}); err != nil {
			return err
		}
		if err := b.ledger.RecordValidationResult(ctx, snap.ID, "background-task", ValidationResult{
			Passed: true,
			Reason: "background task completed",
			Details: map[string]string{
				"status": snap.StatusText,
			},
		}); err != nil {
			return err
		}
		if err := b.ledger.AppendJournalEvent(ctx, JournalEvent{
			RunID:   snap.ID,
			Type:    EventRunCompleted,
			Payload: marshalPayload(RunCompletedPayload{Summary: snap.Result}),
		}); err != nil {
			return err
		}
		if b.maxKeep > 0 {
			if err := b.ledger.PruneOldRuns(ctx, b.maxKeep); err != nil {
				return err
			}
		}
	case "failed":
		if err := b.ledger.RecordValidationResult(ctx, snap.ID, "background-task", ValidationResult{
			Passed: false,
			Reason: snap.Error,
			Details: map[string]string{
				"status": snap.StatusText,
			},
		}); err != nil {
			return err
		}
		if err := b.ledger.AppendJournalEvent(ctx, JournalEvent{
			RunID:   snap.ID,
			Type:    EventRunFailed,
			Payload: marshalPayload(RunFailedPayload{Reason: snap.Error}),
		}); err != nil {
			return err
		}
		if b.maxKeep > 0 {
			if err := b.ledger.PruneOldRuns(ctx, b.maxKeep); err != nil {
				return err
			}
		}
	case "cancelled":
		if err := b.ledger.AppendJournalEvent(ctx, JournalEvent{
			RunID:   snap.ID,
			Type:    EventRunFailed,
			Payload: marshalPayload(RunFailedPayload{Reason: "background task cancelled"}),
		}); err != nil {
			return err
		}
		if b.maxKeep > 0 {
			if err := b.ledger.PruneOldRuns(ctx, b.maxKeep); err != nil {
				return err
			}
		}
	}

	return appendProjectionSyncEvent(ctx, b.ledger, snap.ID, "background", "synced", nil)
}
