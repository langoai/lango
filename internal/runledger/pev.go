package runledger

import (
	"context"
	"fmt"
	"time"
)

// Validator executes a specific validation strategy against step evidence.
type Validator interface {
	Validate(ctx context.Context, spec ValidatorSpec, evidence []Evidence) (*ValidationResult, error)
}

// PEVEngine is the Propose-Evidence-Verify engine.
// It runs typed validators against step results and records outcomes
// in the journal. It never modifies step status directly.
type PEVEngine struct {
	ledger     RunLedgerStore
	validators map[ValidatorType]Validator
	workspace  *WorkspaceManager // nil = no isolation (Phase 1 default)
	timeout    time.Duration
	maxHistory int
}

// NewPEVEngine creates a PEV engine with the provided store and validators.
func NewPEVEngine(ledger RunLedgerStore, validators map[ValidatorType]Validator) *PEVEngine {
	return &PEVEngine{
		ledger:     ledger,
		validators: validators,
	}
}

// WithWorkspace enables workspace isolation for coding steps.
// Phase 1 default is nil (no isolation). Phase 3 activates with:
//
//	pev.WithWorkspace(NewWorkspaceManager())
func (e *PEVEngine) WithWorkspace(ws *WorkspaceManager) *PEVEngine {
	e.workspace = ws
	return e
}

// WithTimeout configures a validator execution deadline.
func (e *PEVEngine) WithTimeout(timeout time.Duration) *PEVEngine {
	e.timeout = timeout
	return e
}

// WithMaxRunHistory configures how many runs should be retained in the store.
func (e *PEVEngine) WithMaxRunHistory(maxHistory int) *PEVEngine {
	e.maxHistory = maxHistory
	return e
}

// WorkspaceEnabled reports whether runtime workspace isolation is wired in.
func (e *PEVEngine) WorkspaceEnabled() bool {
	return e.workspace != nil
}

// Verify runs the step's validator and records the result in the journal.
// Returns a PolicyRequest if validation fails, nil if it passes.
func (e *PEVEngine) Verify(ctx context.Context, runID string, step *Step) (*PolicyRequest, error) {
	// Workspace isolation: prepare before validation.
	if e.workspace != nil {
		cleanup, wsErr := e.workspace.PrepareStepWorkspace(step, runID)
		if wsErr != nil {
			// Fail-closed: workspace creation failed -> return as PolicyRequest.
			return &PolicyRequest{
				RunID:    runID,
				StepID:   step.StepID,
				StepGoal: step.Goal,
				Failure: &ValidationResult{
					Passed: false,
					Reason: fmt.Sprintf("workspace isolation failed: %v", wsErr),
				},
				RetryCount: step.RetryCount,
				MaxRetries: step.MaxRetries,
			}, nil
		}
		defer cleanup()
	}

	v, ok := e.validators[step.Validator.Type]
	if !ok {
		return nil, fmt.Errorf("no validator registered for type %q", step.Validator.Type)
	}

	validateCtx := ctx
	cancel := func() {}
	if e.timeout > 0 {
		validateCtx, cancel = context.WithTimeout(ctx, e.timeout)
	}
	defer cancel()

	result, err := v.Validate(validateCtx, step.Validator, step.Evidence)
	if err != nil {
		return nil, fmt.Errorf("validator %q: %w", step.Validator.Type, err)
	}

	if err := e.ledger.RecordValidationResult(ctx, runID, step.StepID, *result); err != nil {
		return nil, fmt.Errorf("record validation result: %w", err)
	}

	if result.Passed {
		return nil, nil
	}

	return &PolicyRequest{
		RunID:      runID,
		StepID:     step.StepID,
		StepGoal:   step.Goal,
		Failure:    result,
		RetryCount: step.RetryCount,
		MaxRetries: step.MaxRetries,
	}, nil
}

// VerifyAcceptanceCriteria checks all acceptance criteria against the current state.
// It returns both unmet criteria and a fully evaluated copy.
func (e *PEVEngine) VerifyAcceptanceCriteria(
	ctx context.Context,
	criteria []AcceptanceCriterion,
) ([]AcceptanceCriterion, []AcceptanceCriterion, error) {
	evaluated := make([]AcceptanceCriterion, len(criteria))
	for i := range criteria {
		evaluated[i] = copyAcceptanceCriterion(criteria[i])
	}

	unmet := make([]AcceptanceCriterion, 0, len(evaluated))
	for i := range evaluated {
		if evaluated[i].Met {
			continue
		}
		v, ok := e.validators[evaluated[i].Validator.Type]
		if !ok {
			unmet = append(unmet, copyAcceptanceCriterion(evaluated[i]))
			continue
		}
		result, err := v.Validate(ctx, evaluated[i].Validator, nil)
		if err != nil {
			unmet = append(unmet, copyAcceptanceCriterion(evaluated[i]))
			continue
		}
		if result.Passed {
			now := time.Now()
			evaluated[i].Met = true
			evaluated[i].MetAt = &now
			continue
		}
		unmet = append(unmet, copyAcceptanceCriterion(evaluated[i]))
	}
	return unmet, evaluated, nil
}

func (e *PEVEngine) maybePruneRunHistory(ctx context.Context) error {
	if e.maxHistory <= 0 {
		return nil
	}
	return e.ledger.PruneOldRuns(ctx, e.maxHistory)
}
