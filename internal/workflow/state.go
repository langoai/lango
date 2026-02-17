package workflow

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/ent"
	"github.com/langowarny/lango/internal/ent/workflowrun"
	"github.com/langowarny/lango/internal/ent/workflowsteprun"
)

// StateStore persists workflow execution state for resume capability.
type StateStore struct {
	client *ent.Client
	logger *zap.SugaredLogger
}

// NewStateStore creates a new StateStore backed by the given ent client.
func NewStateStore(client *ent.Client, logger *zap.SugaredLogger) *StateStore {
	return &StateStore{
		client: client,
		logger: logger,
	}
}

// CreateRun creates a new workflow run record and returns its ID.
func (s *StateStore) CreateRun(ctx context.Context, w *Workflow) (string, error) {
	now := time.Now()
	run, err := s.client.WorkflowRun.Create().
		SetWorkflowName(w.Name).
		SetDescription(w.Description).
		SetStatus(workflowrun.StatusPending).
		SetTotalSteps(len(w.Steps)).
		SetCompletedSteps(0).
		SetStartedAt(now).
		Save(ctx)
	if err != nil {
		return "", fmt.Errorf("create workflow run: %w", err)
	}
	return run.ID.String(), nil
}

// UpdateRunStatus updates the status of a workflow run.
func (s *StateStore) UpdateRunStatus(ctx context.Context, runID string, status string) error {
	uid, err := uuid.Parse(runID)
	if err != nil {
		return fmt.Errorf("parse run ID %q: %w", runID, err)
	}
	return s.client.WorkflowRun.Update().
		Where(workflowrun.ID(uid)).
		SetStatus(workflowrun.Status(status)).
		Exec(ctx)
}

// CompleteRun marks a workflow run as finished with a final status and optional error message.
func (s *StateStore) CompleteRun(ctx context.Context, runID string, status string, errMsg string) error {
	uid, err := uuid.Parse(runID)
	if err != nil {
		return fmt.Errorf("parse run ID %q: %w", runID, err)
	}
	now := time.Now()
	builder := s.client.WorkflowRun.Update().
		Where(workflowrun.ID(uid)).
		SetStatus(workflowrun.Status(status)).
		SetCompletedAt(now)
	if errMsg != "" {
		builder = builder.SetErrorMessage(errMsg)
	}
	return builder.Exec(ctx)
}

// CreateStepRun creates a new step run record for a workflow run.
func (s *StateStore) CreateStepRun(ctx context.Context, runID string, step Step, renderedPrompt string) error {
	uid, err := uuid.Parse(runID)
	if err != nil {
		return fmt.Errorf("parse run ID %q: %w", runID, err)
	}
	builder := s.client.WorkflowStepRun.Create().
		SetRunID(uid).
		SetStepID(step.ID).
		SetPrompt(renderedPrompt).
		SetStatus(workflowsteprun.StatusPending)
	if step.Agent != "" {
		builder = builder.SetAgent(step.Agent)
	}
	return builder.Exec(ctx)
}

// UpdateStepStatus updates the status, result, and error message of a step run.
func (s *StateStore) UpdateStepStatus(ctx context.Context, runID string, stepID string, status string, result string, errMsg string) error {
	uid, err := uuid.Parse(runID)
	if err != nil {
		return fmt.Errorf("parse run ID %q: %w", runID, err)
	}

	now := time.Now()
	builder := s.client.WorkflowStepRun.Update().
		Where(
			workflowsteprun.RunID(uid),
			workflowsteprun.StepID(stepID),
		).
		SetStatus(workflowsteprun.Status(status))

	switch status {
	case "running":
		builder = builder.SetStartedAt(now)
	case "completed", "failed", "skipped":
		builder = builder.SetCompletedAt(now)
	}

	if result != "" {
		builder = builder.SetResult(result)
	}
	if errMsg != "" {
		builder = builder.SetErrorMessage(errMsg)
	}

	// Also increment completed_steps on the parent run when a step finishes.
	if status == "completed" || status == "failed" || status == "skipped" {
		if updateErr := s.client.WorkflowRun.Update().
			Where(workflowrun.ID(uid)).
			AddCompletedSteps(1).
			Exec(ctx); updateErr != nil {
			s.logger.Warnw("increment completed_steps", "runID", runID, "error", updateErr)
		}
	}

	return builder.Exec(ctx)
}

// GetRunStatus returns the current status of a workflow run including all step statuses.
func (s *StateStore) GetRunStatus(ctx context.Context, runID string) (*RunStatus, error) {
	uid, err := uuid.Parse(runID)
	if err != nil {
		return nil, fmt.Errorf("parse run ID %q: %w", runID, err)
	}

	run, err := s.client.WorkflowRun.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get workflow run %q: %w", runID, err)
	}

	steps, err := s.client.WorkflowStepRun.Query().
		Where(workflowsteprun.RunID(uid)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query step runs for %q: %w", runID, err)
	}

	statuses := make([]StepStatus, 0, len(steps))
	for _, st := range steps {
		statuses = append(statuses, StepStatus{
			StepID: st.StepID,
			Agent:  st.Agent,
			Status: string(st.Status),
			Error:  st.ErrorMessage,
		})
	}

	return &RunStatus{
		RunID:          runID,
		WorkflowName:   run.WorkflowName,
		Status:         string(run.Status),
		TotalSteps:     run.TotalSteps,
		CompletedSteps: run.CompletedSteps,
		StartedAt:      run.StartedAt,
		StepStatuses:   statuses,
	}, nil
}

// GetStepResults returns a map of stepID -> result for all completed steps.
func (s *StateStore) GetStepResults(ctx context.Context, runID string) (map[string]string, error) {
	uid, err := uuid.Parse(runID)
	if err != nil {
		return nil, fmt.Errorf("parse run ID %q: %w", runID, err)
	}

	steps, err := s.client.WorkflowStepRun.Query().
		Where(
			workflowsteprun.RunID(uid),
			workflowsteprun.StatusEQ(workflowsteprun.StatusCompleted),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query step results for %q: %w", runID, err)
	}

	results := make(map[string]string, len(steps))
	for _, st := range steps {
		results[st.StepID] = st.Result
	}
	return results, nil
}

// ListRuns returns the most recent workflow runs, ordered by start time descending.
func (s *StateStore) ListRuns(ctx context.Context, limit int) ([]RunStatus, error) {
	runs, err := s.client.WorkflowRun.Query().
		Order(workflowrun.ByStartedAt(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list workflow runs: %w", err)
	}

	result := make([]RunStatus, 0, len(runs))
	for _, r := range runs {
		result = append(result, RunStatus{
			RunID:          r.ID.String(),
			WorkflowName:   r.WorkflowName,
			Status:         string(r.Status),
			TotalSteps:     r.TotalSteps,
			CompletedSteps: r.CompletedSteps,
			StartedAt:      r.StartedAt,
		})
	}
	return result, nil
}
