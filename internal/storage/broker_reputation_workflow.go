package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/storagebroker"
	"github.com/langoai/lango/internal/workflow"
)

type brokerReputationStore struct {
	broker storagebroker.API
}

func (s *brokerReputationStore) GetDetails(ctx context.Context, peerDID string) (*reputation.PeerDetails, error) {
	result, err := s.broker.ReputationGet(ctx, peerDID)
	if err != nil {
		return nil, err
	}
	if !result.Found {
		return nil, nil
	}
	return &reputation.PeerDetails{
		PeerDID:             result.PeerDID,
		TrustScore:          result.TrustScore,
		SuccessfulExchanges: result.SuccessfulExchanges,
		FailedExchanges:     result.FailedExchanges,
		TimeoutCount:        result.TimeoutCount,
		FirstSeen:           result.FirstSeen,
		LastInteraction:     result.LastInteraction,
	}, nil
}

func (s *brokerReputationStore) GetScore(ctx context.Context, peerDID string) (float64, error) {
	result, err := s.broker.ReputationGet(ctx, peerDID)
	if err != nil {
		return 0, err
	}
	if !result.Found {
		return 0, nil
	}
	return result.TrustScore, nil
}

type brokerWorkflowRunStore struct {
	broker storagebroker.API
}

func (s *brokerWorkflowRunStore) CreateRun(ctx context.Context, w *workflow.Workflow) (string, error) {
	return "", fmt.Errorf("workflow mutators unavailable in reader-only broker runtime")
}

func (s *brokerWorkflowRunStore) UpdateRunStatus(ctx context.Context, runID string, status string) error {
	return fmt.Errorf("workflow mutators unavailable in reader-only broker runtime")
}

func (s *brokerWorkflowRunStore) CompleteRun(ctx context.Context, runID string, status string, errMsg string) error {
	return fmt.Errorf("workflow mutators unavailable in reader-only broker runtime")
}

func (s *brokerWorkflowRunStore) CreateStepRun(ctx context.Context, runID string, step workflow.Step, renderedPrompt string) error {
	return fmt.Errorf("workflow mutators unavailable in reader-only broker runtime")
}

func (s *brokerWorkflowRunStore) UpdateStepStatus(ctx context.Context, runID string, stepID string, status string, result string, errMsg string) error {
	return fmt.Errorf("workflow mutators unavailable in reader-only broker runtime")
}

func (s *brokerWorkflowRunStore) GetRunStatus(ctx context.Context, runID string) (*workflow.RunStatus, error) {
	runs, err := s.broker.WorkflowRuns(ctx, 100)
	if err != nil {
		return nil, err
	}
	for _, run := range runs.Runs {
		if run.RunID == runID {
			return &workflow.RunStatus{
				RunID:          run.RunID,
				WorkflowName:   run.WorkflowName,
				Status:         run.Status,
				TotalSteps:     run.TotalSteps,
				CompletedSteps: run.CompletedSteps,
				StartedAt:      run.StartedAt,
			}, nil
		}
	}
	return nil, fmt.Errorf("workflow run %q not found", runID)
}

func (s *brokerWorkflowRunStore) GetStepResults(ctx context.Context, runID string) (map[string]string, error) {
	return nil, fmt.Errorf("workflow step results unavailable in reader-only broker runtime")
}

func (s *brokerWorkflowRunStore) ListRuns(ctx context.Context, limit int) ([]workflow.RunStatus, error) {
	result, err := s.broker.WorkflowRuns(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]workflow.RunStatus, 0, len(result.Runs))
	for _, run := range result.Runs {
		out = append(out, workflow.RunStatus{
			RunID:          run.RunID,
			WorkflowName:   run.WorkflowName,
			Status:         run.Status,
			TotalSteps:     run.TotalSteps,
			CompletedSteps: run.CompletedSteps,
			StartedAt:      run.StartedAt,
			StepStatuses:   nil,
		})
	}
	return out, nil
}

var _ workflow.RunStore = (*brokerWorkflowRunStore)(nil)

func _() {
	_ = time.Time{}
}
