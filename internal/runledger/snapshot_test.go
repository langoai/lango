package runledger

import (
	"encoding/json"
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

func TestRunSnapshot_DeepCopy(t *testing.T) {
	now := time.Now()
	snap := &RunSnapshot{
		RunID:           "run-1",
		SessionKey:      "session-1",
		OriginalRequest: "original",
		Goal:            "goal",
		Status:          RunStatusRunning,
		CurrentStepID:   "s1",
		CurrentBlocker:  "none",
		AcceptanceState: []AcceptanceCriterion{{
			Description: "criterion",
			Validator: ValidatorSpec{
				Type:   ValidatorBuildPass,
				Target: "./...",
				Params: map[string]string{"k": "v"},
			},
			Met:   true,
			MetAt: &now,
		}},
		Steps: []Step{{
			StepID:     "s1",
			Goal:       "step",
			OwnerAgent: "operator",
			Status:     StepStatusInProgress,
			Evidence: []Evidence{{
				Type:    "file",
				Content: "a.go",
			}},
			Validator: ValidatorSpec{
				Type:   ValidatorBuildPass,
				Target: "./internal/runledger",
				Params: map[string]string{"mode": "fast"},
			},
			ToolProfile: []string{string(ToolProfileCoding)},
			DependsOn:   []string{"root"},
		}},
		Notes:          map[string]string{"note": "value"},
		LastJournalSeq: 7,
		UpdatedAt:      now,
	}

	cp := snap.DeepCopy()
	require.NotNil(t, cp)
	require.NotSame(t, snap, cp)
	require.NotSame(t, &snap.Steps[0], &cp.Steps[0])
	require.NotSame(t, snap.AcceptanceState[0].MetAt, cp.AcceptanceState[0].MetAt)

	cp.Steps[0].Evidence[0].Content = "b.go"
	cp.Steps[0].Validator.Params["mode"] = "slow"
	cp.Steps[0].DependsOn[0] = "other"
	cp.Steps[0].ToolProfile[0] = string(ToolProfileSupervisor)
	cp.AcceptanceState[0].Validator.Params["k"] = "changed"
	*cp.AcceptanceState[0].MetAt = now.Add(time.Hour)
	cp.Notes["note"] = "updated"
	cp.Steps = append(cp.Steps, Step{StepID: "s2"})

	assert.Equal(t, "a.go", snap.Steps[0].Evidence[0].Content)
	assert.Equal(t, "fast", snap.Steps[0].Validator.Params["mode"])
	assert.Equal(t, "root", snap.Steps[0].DependsOn[0])
	assert.Equal(t, string(ToolProfileCoding), snap.Steps[0].ToolProfile[0])
	assert.Equal(t, "v", snap.AcceptanceState[0].Validator.Params["k"])
	assert.True(t, snap.AcceptanceState[0].MetAt.Equal(now))
	assert.Equal(t, "value", snap.Notes["note"])
	assert.Len(t, snap.Steps, 1)
}

func TestDeepCopy_SourceDescriptor(t *testing.T) {
	original := &RunSnapshot{
		RunID:            "r1",
		SourceKind:       "workflow",
		SourceDescriptor: json.RawMessage(`{"name":"test-wf"}`),
		Notes:            map[string]string{},
	}
	cp := original.DeepCopy()

	// Verify values are equal
	assert.Equal(t, original.SourceKind, cp.SourceKind)
	assert.Equal(t, string(original.SourceDescriptor), string(cp.SourceDescriptor))

	// Verify backing arrays are independent
	cp.SourceDescriptor[0] = 'X'
	assert.NotEqual(t, original.SourceDescriptor[0], cp.SourceDescriptor[0],
		"modifying copy should not affect original")
}

func TestFindStep_AfterPlanAttached(t *testing.T) {
	now := time.Now()
	events := []JournalEvent{
		{
			RunID: "run-1", Seq: 1, Type: EventRunCreated, Timestamp: now,
			Payload: marshalPayload(RunCreatedPayload{
				SessionKey: "s1", OriginalRequest: "req", Goal: "goal",
			}),
		},
		{
			RunID: "run-1", Seq: 2, Type: EventPlanAttached, Timestamp: now.Add(time.Second),
			Payload: marshalPayload(PlanAttachedPayload{
				Steps: []Step{
					{StepID: "alpha", Index: 0, Goal: "first step", Status: StepStatusPending},
					{StepID: "beta", Index: 1, Goal: "second step", Status: StepStatusPending},
					{StepID: "gamma", Index: 2, Goal: "third step", Status: StepStatusPending},
				},
			}),
		},
	}

	snap, err := MaterializeFromJournal(events)
	require.NoError(t, err)

	tests := []struct {
		give     string
		wantGoal string
		wantNil  bool
	}{
		{give: "alpha", wantGoal: "first step"},
		{give: "beta", wantGoal: "second step"},
		{give: "gamma", wantGoal: "third step"},
		{give: "missing", wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			step := snap.FindStep(tt.give)
			if tt.wantNil {
				assert.Nil(t, step)
			} else {
				require.NotNil(t, step)
				assert.Equal(t, tt.wantGoal, step.Goal)
			}
		})
	}
}

func TestFindStep_AfterPolicyDecompose(t *testing.T) {
	now := time.Now()
	events := []JournalEvent{
		{
			RunID: "run-1", Seq: 1, Type: EventRunCreated, Timestamp: now,
			Payload: marshalPayload(RunCreatedPayload{
				SessionKey: "s1", OriginalRequest: "req", Goal: "goal",
			}),
		},
		{
			RunID: "run-1", Seq: 2, Type: EventPlanAttached, Timestamp: now.Add(time.Second),
			Payload: marshalPayload(PlanAttachedPayload{
				Steps: []Step{
					{StepID: "s1", Index: 0, Goal: "original", Status: StepStatusPending},
				},
			}),
		},
	}

	snap, err := MaterializeFromJournal(events)
	require.NoError(t, err)

	require.NotNil(t, snap.FindStep("s1"))

	applyPolicyToSnapshot(snap, "s1", &PolicyDecision{
		Action: PolicyDecompose,
		NewSteps: []Step{
			{StepID: "s1-sub-a", Goal: "decomposed A"},
			{StepID: "s1-sub-b", Goal: "decomposed B"},
		},
	})

	stepA := snap.FindStep("s1-sub-a")
	require.NotNil(t, stepA)
	assert.Equal(t, "decomposed A", stepA.Goal)

	stepB := snap.FindStep("s1-sub-b")
	require.NotNil(t, stepB)
	assert.Equal(t, "decomposed B", stepB.Goal)

	orig := snap.FindStep("s1")
	require.NotNil(t, orig)
	assert.Equal(t, StepStatusCompleted, orig.Status)
}

func TestFindStep_AfterDeepCopy(t *testing.T) {
	snap := &RunSnapshot{
		Steps: []Step{
			{StepID: "x1", Goal: "goal-x1"},
			{StepID: "x2", Goal: "goal-x2"},
		},
	}

	require.NotNil(t, snap.FindStep("x1"))

	cp := snap.DeepCopy()
	snap.Steps[0].Goal = "mutated"

	step := cp.FindStep("x1")
	require.NotNil(t, step)
	assert.Equal(t, "goal-x1", step.Goal, "copy must be independent of original mutations")

	step2 := cp.FindStep("x2")
	require.NotNil(t, step2)
	assert.Equal(t, "goal-x2", step2.Goal)

	assert.Nil(t, cp.FindStep("missing"))
}

func TestFindStep_AfterJSONRehydrate(t *testing.T) {
	original := &RunSnapshot{
		RunID:  "run-json",
		Status: RunStatusRunning,
		Steps: []Step{
			{StepID: "j1", Goal: "json-step-1"},
			{StepID: "j2", Goal: "json-step-2"},
		},
		Notes: map[string]string{"k": "v"},
	}

	require.NotNil(t, original.FindStep("j1"))

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var rehydrated RunSnapshot
	require.NoError(t, json.Unmarshal(data, &rehydrated))

	assert.Nil(t, rehydrated.stepIndex, "stepIndex should be nil after JSON unmarshal")

	step := rehydrated.FindStep("j1")
	require.NotNil(t, step)
	assert.Equal(t, "json-step-1", step.Goal)

	step2 := rehydrated.FindStep("j2")
	require.NotNil(t, step2)
	assert.Equal(t, "json-step-2", step2.Goal)

	assert.Nil(t, rehydrated.FindStep("nonexistent"))
}
