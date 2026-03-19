package runledger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaterializeFromJournal_EmptyEvents(t *testing.T) {
	_, err := MaterializeFromJournal(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty journal")
}

func TestMaterializeFromJournal_BasicFlow(t *testing.T) {
	now := time.Now()
	events := []JournalEvent{
		{
			RunID:     "run-1",
			Seq:       1,
			Type:      EventRunCreated,
			Timestamp: now,
			Payload: marshalPayload(RunCreatedPayload{
				SessionKey:      "session-1",
				OriginalRequest: "build a feature",
				Goal:            "build feature X",
			}),
		},
		{
			RunID:     "run-1",
			Seq:       2,
			Type:      EventPlanAttached,
			Timestamp: now.Add(time.Second),
			Payload: marshalPayload(PlanAttachedPayload{
				Steps: []Step{
					{StepID: "s1", Index: 0, Goal: "write code", OwnerAgent: "operator", Status: StepStatusPending, MaxRetries: 2},
					{StepID: "s2", Index: 1, Goal: "test code", OwnerAgent: "operator", Status: StepStatusPending, MaxRetries: 2, DependsOn: []string{"s1"}},
				},
				AcceptanceCriteria: []AcceptanceCriterion{
					{Description: "build passes", Validator: ValidatorSpec{Type: ValidatorBuildPass}},
				},
			}),
		},
		{
			RunID:     "run-1",
			Seq:       3,
			Type:      EventStepStarted,
			Timestamp: now.Add(2 * time.Second),
			Payload:   marshalPayload(StepStartedPayload{StepID: "s1", OwnerAgent: "operator"}),
		},
	}

	snap, err := MaterializeFromJournal(events)
	require.NoError(t, err)

	assert.Equal(t, "run-1", snap.RunID)
	assert.Equal(t, "session-1", snap.SessionKey)
	assert.Equal(t, RunStatusRunning, snap.Status)
	assert.Equal(t, "build feature X", snap.Goal)
	assert.Len(t, snap.Steps, 2)
	assert.Equal(t, StepStatusInProgress, snap.Steps[0].Status)
	assert.Equal(t, StepStatusPending, snap.Steps[1].Status)
	assert.Equal(t, "s1", snap.CurrentStepID)
	assert.Equal(t, int64(3), snap.LastJournalSeq)
}

func TestRunSnapshot_NextExecutableStep(t *testing.T) {
	snap := &RunSnapshot{
		Steps: []Step{
			{StepID: "s1", Status: StepStatusCompleted},
			{StepID: "s2", Status: StepStatusPending, DependsOn: []string{"s1"}},
			{StepID: "s3", Status: StepStatusPending, DependsOn: []string{"s2"}},
		},
	}

	next := snap.NextExecutableStep()
	require.NotNil(t, next)
	assert.Equal(t, "s2", next.StepID)
}

func TestRunSnapshot_NextExecutableStep_NoneReady(t *testing.T) {
	snap := &RunSnapshot{
		Steps: []Step{
			{StepID: "s1", Status: StepStatusInProgress},
			{StepID: "s2", Status: StepStatusPending, DependsOn: []string{"s1"}},
		},
	}

	next := snap.NextExecutableStep()
	assert.Nil(t, next)
}

func TestRunSnapshot_AllStepsTerminal(t *testing.T) {
	tests := []struct {
		give []Step
		want bool
	}{
		{
			give: []Step{
				{Status: StepStatusCompleted},
				{Status: StepStatusFailed},
				{Status: StepStatusInterrupted},
			},
			want: true,
		},
		{
			give: []Step{
				{Status: StepStatusCompleted},
				{Status: StepStatusPending},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		snap := &RunSnapshot{Steps: tt.give}
		assert.Equal(t, tt.want, snap.AllStepsTerminal())
	}
}

func TestRunSnapshot_ToSummary(t *testing.T) {
	snap := &RunSnapshot{
		RunID:         "run-1",
		Goal:          "test goal",
		Status:        RunStatusRunning,
		CurrentStepID: "s2",
		Steps: []Step{
			{StepID: "s1", Goal: "first", Status: StepStatusCompleted},
			{StepID: "s2", Goal: "second", Status: StepStatusInProgress},
			{StepID: "s3", Goal: "third", Status: StepStatusPending},
		},
		AcceptanceState: []AcceptanceCriterion{
			{Description: "build passes", Met: true},
			{Description: "tests pass", Met: false},
		},
	}

	summary := snap.ToSummary()
	assert.Equal(t, "run-1", summary.RunID)
	assert.Equal(t, 3, summary.TotalSteps)
	assert.Equal(t, 1, summary.CompletedSteps)
	assert.Equal(t, "second", summary.CurrentStepGoal)
	assert.Equal(t, []string{"tests pass"}, summary.UnmetCriteria)
}

func TestApplyTail(t *testing.T) {
	// Start with a snapshot at seq 2.
	snap := &RunSnapshot{
		RunID:          "run-1",
		Status:         RunStatusRunning,
		LastJournalSeq: 2,
		Steps: []Step{
			{StepID: "s1", Status: StepStatusInProgress},
		},
		Notes: make(map[string]string),
	}

	tail := []JournalEvent{
		{
			RunID: "run-1", Seq: 1, // should be skipped (seq <= 2)
			Type:    EventNoteWritten,
			Payload: marshalPayload(NoteWrittenPayload{Key: "old", Value: "skip"}),
		},
		{
			RunID: "run-1", Seq: 3,
			Type:      EventStepResultProposed,
			Timestamp: time.Now(),
			Payload: marshalPayload(StepResultProposedPayload{
				StepID: "s1", Result: "done",
			}),
		},
	}

	err := ApplyTail(snap, tail)
	require.NoError(t, err)
	assert.Equal(t, int64(3), snap.LastJournalSeq)
	assert.Equal(t, StepStatusVerifyPending, snap.Steps[0].Status)
	assert.Empty(t, snap.Notes) // old note should be skipped
}

func TestApplyPolicyToSnapshot_Retry(t *testing.T) {
	snap := &RunSnapshot{
		Steps: []Step{
			{StepID: "s1", Status: StepStatusFailed, RetryCount: 0},
		},
		CurrentBlocker: "some error",
	}

	applyPolicyToSnapshot(snap, "s1", &PolicyDecision{Action: PolicyRetry})

	assert.Equal(t, StepStatusPending, snap.Steps[0].Status)
	assert.Equal(t, 1, snap.Steps[0].RetryCount)
	assert.Empty(t, snap.CurrentBlocker)
}

func TestApplyPolicyToSnapshot_Decompose(t *testing.T) {
	snap := &RunSnapshot{
		Steps: []Step{
			{StepID: "s1", Status: StepStatusFailed},
		},
	}

	applyPolicyToSnapshot(snap, "s1", &PolicyDecision{
		Action: PolicyDecompose,
		NewSteps: []Step{
			{StepID: "s1a", Goal: "sub-task A"},
			{StepID: "s1b", Goal: "sub-task B"},
		},
	})

	assert.Equal(t, StepStatusCompleted, snap.Steps[0].Status)
	assert.Len(t, snap.Steps, 3) // original + 2 new
	assert.Equal(t, "s1a", snap.Steps[1].StepID)
}

func TestApplyPolicyToSnapshot_Abort(t *testing.T) {
	snap := &RunSnapshot{
		Status: RunStatusRunning,
		Steps: []Step{
			{StepID: "s1", Status: StepStatusFailed},
		},
	}

	applyPolicyToSnapshot(snap, "s1", &PolicyDecision{
		Action: PolicyAbort,
		Reason: "unrecoverable",
	})

	assert.Equal(t, RunStatusFailed, snap.Status)
	assert.Equal(t, "unrecoverable", snap.CurrentBlocker)
}
