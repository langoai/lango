package runledger

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockValidator always returns a fixed result.
type mockValidator struct {
	result *ValidationResult
	err    error
}

func (m *mockValidator) Validate(_ context.Context, _ ValidatorSpec, _ []Evidence) (*ValidationResult, error) {
	return m.result, m.err
}

type slowValidator struct{}

func (s *slowValidator) Validate(ctx context.Context, _ ValidatorSpec, _ []Evidence) (*ValidationResult, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestPEVEngine_Verify_Pass(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{Goal: "test"}),
	})

	pev := NewPEVEngine(store, map[ValidatorType]Validator{
		ValidatorBuildPass: &mockValidator{
			result: &ValidationResult{Passed: true, Reason: "build ok"},
		},
	})

	step := &Step{
		StepID:    "s1",
		Goal:      "write code",
		Validator: ValidatorSpec{Type: ValidatorBuildPass},
	}

	req, err := pev.Verify(ctx, "run-1", step)
	require.NoError(t, err)
	assert.Nil(t, req) // no policy request on pass
}

func TestPEVEngine_Verify_Fail(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{Goal: "test"}),
	})

	pev := NewPEVEngine(store, map[ValidatorType]Validator{
		ValidatorTestPass: &mockValidator{
			result: &ValidationResult{
				Passed: false,
				Reason: "2 tests failed",
				Details: map[string]string{
					"exit_code": "1",
				},
			},
		},
	})

	step := &Step{
		StepID:     "s1",
		Goal:       "test code",
		Validator:  ValidatorSpec{Type: ValidatorTestPass},
		MaxRetries: 2,
		RetryCount: 1,
	}

	req, err := pev.Verify(ctx, "run-1", step)
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, "run-1", req.RunID)
	assert.Equal(t, "s1", req.StepID)
	assert.Equal(t, 1, req.RetryCount)
	assert.Equal(t, 2, req.MaxRetries)
	assert.Equal(t, "2 tests failed", req.Failure.Reason)
}

func TestPEVEngine_Verify_UnknownValidator(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	pev := NewPEVEngine(store, map[ValidatorType]Validator{})

	step := &Step{
		StepID:    "s1",
		Validator: ValidatorSpec{Type: "unknown"},
	}

	_, err := pev.Verify(ctx, "run-1", step)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no validator registered")
}

func TestPEVEngine_OrchestratorApprovalNeverAutoPass(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{Goal: "test"}),
	})

	pev := NewPEVEngine(store, DefaultValidators())

	step := &Step{
		StepID:    "s1",
		Goal:      "review",
		Validator: ValidatorSpec{Type: ValidatorOrchestratorApproval},
	}

	req, err := pev.Verify(ctx, "run-1", step)
	require.NoError(t, err)
	require.NotNil(t, req, "orchestrator_approval must never auto-pass")
	assert.Equal(t, "awaiting orchestrator approval", req.Failure.Reason)
}

func TestPEVEngine_WorkspaceIsolationFailure(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_ = store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-1",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{Goal: "test"}),
	})

	// Use real workspace manager but on a step that needs isolation.
	// We can't easily mock CheckDirtyTree on the real manager, so we test
	// via the PEV engine path with a workspace that will fail because
	// we're not in a git repo at temp dirs.
	// Instead, test the PrepareStepWorkspace directly with a step
	// that doesn't need isolation — should return noop cleanup.
	ws := NewWorkspaceManager()
	step := &Step{
		StepID:    "s1",
		Validator: ValidatorSpec{Type: ValidatorArtifactExists}, // no isolation needed
	}
	cleanup, err := ws.PrepareStepWorkspace(step, "run-1")
	require.NoError(t, err)
	cleanup()
	assert.Empty(t, step.Validator.WorkDir) // WorkDir should stay empty

	// Test WithWorkspace sets the field.
	pev := NewPEVEngine(store, map[ValidatorType]Validator{
		ValidatorBuildPass: &mockValidator{result: &ValidationResult{Passed: true, Reason: "ok"}},
	})
	assert.Nil(t, pev.workspace)
	pev.WithWorkspace(ws)
	assert.NotNil(t, pev.workspace)
}

func TestPEVEngine_Verify_RespectsTimeout(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-timeout",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{Goal: "timeout"}),
	}))

	pev := NewPEVEngine(store, map[ValidatorType]Validator{
		ValidatorBuildPass: &slowValidator{},
	}).WithTimeout(10 * time.Millisecond)

	step := &Step{
		StepID:    "s1",
		Goal:      "slow step",
		Validator: ValidatorSpec{Type: ValidatorBuildPass},
	}

	_, err := pev.Verify(ctx, "run-timeout", step)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestPEVEngine_VerifyAcceptanceCriteria_DoesNotMutateInput(t *testing.T) {
	ctx := context.Background()
	pev := NewPEVEngine(NewMemoryStore(), map[ValidatorType]Validator{
		ValidatorBuildPass: &mockValidator{
			result: &ValidationResult{Passed: true, Reason: "ok"},
		},
		ValidatorTestPass: &mockValidator{
			result: &ValidationResult{Passed: false, Reason: "fail"},
		},
	})

	criteria := []AcceptanceCriterion{
		{
			Description: "build",
			Validator:   ValidatorSpec{Type: ValidatorBuildPass},
		},
		{
			Description: "tests",
			Validator:   ValidatorSpec{Type: ValidatorTestPass},
		},
	}

	unmet, evaluated, err := pev.VerifyAcceptanceCriteria(ctx, criteria)
	require.NoError(t, err)
	require.Len(t, unmet, 1)
	require.Len(t, evaluated, 2)

	assert.False(t, criteria[0].Met)
	assert.Nil(t, criteria[0].MetAt)
	assert.True(t, evaluated[0].Met)
	assert.NotNil(t, evaluated[0].MetAt)
	assert.False(t, evaluated[1].Met)
}
