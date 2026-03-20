package runledger

import (
	"encoding/json"
	"time"
)

// JournalEventType identifies the kind of journal event.
type JournalEventType string

const (
	EventRunCreated            JournalEventType = "run_created"
	EventPlanAttached          JournalEventType = "plan_attached"
	EventStepStarted           JournalEventType = "step_started"
	EventStepResultProposed    JournalEventType = "step_result_proposed"
	EventStepValidationPassed  JournalEventType = "step_validation_passed"
	EventStepValidationFailed  JournalEventType = "step_validation_failed"
	EventPolicyDecisionApplied JournalEventType = "policy_decision_applied"
	EventNoteWritten           JournalEventType = "note_written"
	EventRunPaused             JournalEventType = "run_paused"
	EventRunResumed            JournalEventType = "run_resumed"
	EventRunCompleted          JournalEventType = "run_completed"
	EventRunFailed             JournalEventType = "run_failed"
	EventProjectionSynced      JournalEventType = "projection_synced"
	EventCriterionMet          JournalEventType = "criterion_met"
)

// JournalEvent is a single append-only record in the RunLedger journal.
// The journal is the sole source of truth — all other state is derived from it.
type JournalEvent struct {
	ID        string           `json:"id"`
	RunID     string           `json:"run_id"`
	Seq       int64            `json:"seq"`
	Type      JournalEventType `json:"type"`
	Timestamp time.Time        `json:"timestamp"`
	Payload   json.RawMessage  `json:"payload"`
}

// RunCreatedPayload is the payload for EventRunCreated.
type RunCreatedPayload struct {
	SessionKey      string `json:"session_key"`
	OriginalRequest string `json:"original_request"`
	Goal            string `json:"goal"`
}

// PlanAttachedPayload is the payload for EventPlanAttached.
type PlanAttachedPayload struct {
	Steps              []Step              `json:"steps"`
	AcceptanceCriteria []AcceptanceCriterion `json:"acceptance_criteria"`
}

// StepStartedPayload is the payload for EventStepStarted.
type StepStartedPayload struct {
	StepID     string `json:"step_id"`
	OwnerAgent string `json:"owner_agent"`
}

// StepResultProposedPayload is the payload for EventStepResultProposed.
type StepResultProposedPayload struct {
	StepID   string     `json:"step_id"`
	Result   string     `json:"result"`
	Evidence []Evidence `json:"evidence,omitempty"`
}

// StepValidationPassedPayload is the payload for EventStepValidationPassed.
type StepValidationPassedPayload struct {
	StepID string           `json:"step_id"`
	Result ValidationResult `json:"result"`
}

// StepValidationFailedPayload is the payload for EventStepValidationFailed.
type StepValidationFailedPayload struct {
	StepID string           `json:"step_id"`
	Result ValidationResult `json:"result"`
}

// PolicyDecisionAppliedPayload is the payload for EventPolicyDecisionApplied.
type PolicyDecisionAppliedPayload struct {
	StepID   string         `json:"step_id"`
	Decision PolicyDecision `json:"decision"`
}

// NoteWrittenPayload is the payload for EventNoteWritten.
type NoteWrittenPayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RunPausedPayload is the payload for EventRunPaused.
type RunPausedPayload struct {
	Reason string `json:"reason"`
}

// RunResumedPayload is the payload for EventRunResumed.
type RunResumedPayload struct {
	ResumedBy string `json:"resumed_by"`
}

// RunCompletedPayload is the payload for EventRunCompleted.
type RunCompletedPayload struct {
	Summary string `json:"summary"`
}

// RunFailedPayload is the payload for EventRunFailed.
type RunFailedPayload struct {
	Reason string `json:"reason"`
}

// CriterionMetPayload is the payload for EventCriterionMet.
type CriterionMetPayload struct {
	Index       int    `json:"index"`
	Description string `json:"description"`
}

// ProjectionSyncPayload describes the result of syncing a projection target.
type ProjectionSyncPayload struct {
	Target string `json:"target"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
