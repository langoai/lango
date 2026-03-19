package runledger

import (
	"context"
	"fmt"
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

	result, err := v.Validate(ctx, step.Validator, step.Evidence)
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
// Returns the list of unmet criteria.
func (e *PEVEngine) VerifyAcceptanceCriteria(ctx context.Context, criteria []AcceptanceCriterion) ([]AcceptanceCriterion, error) {
	var unmet []AcceptanceCriterion
	for i := range criteria {
		if criteria[i].Met {
			continue
		}
		v, ok := e.validators[criteria[i].Validator.Type]
		if !ok {
			unmet = append(unmet, criteria[i])
			continue
		}
		result, err := v.Validate(ctx, criteria[i].Validator, nil)
		if err != nil {
			unmet = append(unmet, criteria[i])
			continue
		}
		if result.Passed {
			criteria[i].Met = true
			now := ctx.Value(ctxKeyNow{})
			if now == nil {
				// fallback: don't set MetAt
			}
		} else {
			unmet = append(unmet, criteria[i])
		}
	}
	return unmet, nil
}

type ctxKeyNow struct{}
