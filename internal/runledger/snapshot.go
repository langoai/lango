package runledger

import (
	"encoding/json"
	"fmt"
	"time"
)

// RunSnapshot is a materialized view derived entirely from the journal.
// It is a read cache — never the source of truth.
type RunSnapshot struct {
	RunID            string                `json:"run_id"`
	SessionKey       string                `json:"session_key"`
	OriginalRequest  string                `json:"original_request"`
	Goal             string                `json:"goal"`
	Status           RunStatus             `json:"status"`
	CurrentStepID    string                `json:"current_step_id,omitempty"`
	CurrentBlocker   string                `json:"current_blocker,omitempty"`
	AcceptanceState  []AcceptanceCriterion `json:"acceptance_state"`
	Steps            []Step                `json:"steps"`
	Notes            map[string]string     `json:"notes"`
	SourceKind       string                `json:"source_kind,omitempty"`
	SourceDescriptor json.RawMessage       `json:"source_descriptor,omitempty"`
	LastJournalSeq   int64                 `json:"last_journal_seq"`
	UpdatedAt        time.Time             `json:"updated_at"`
	stepIndex        map[string]int        `json:"-"` // lazy-built, nil = needs rebuild
}

// DeepCopy returns a fully independent copy of the snapshot.
func (s *RunSnapshot) DeepCopy() *RunSnapshot {
	if s == nil {
		return nil
	}

	cp := *s
	cp.Steps = make([]Step, len(s.Steps))
	for i := range s.Steps {
		cp.Steps[i] = copyStep(s.Steps[i])
	}

	cp.AcceptanceState = make([]AcceptanceCriterion, len(s.AcceptanceState))
	for i := range s.AcceptanceState {
		cp.AcceptanceState[i] = copyAcceptanceCriterion(s.AcceptanceState[i])
	}

	if s.Notes != nil {
		cp.Notes = make(map[string]string, len(s.Notes))
		for k, v := range s.Notes {
			cp.Notes[k] = v
		}
	}

	if s.SourceDescriptor != nil {
		cp.SourceDescriptor = make(json.RawMessage, len(s.SourceDescriptor))
		copy(cp.SourceDescriptor, s.SourceDescriptor)
	}

	cp.stepIndex = nil // lazy rebuild on next FindStep

	return &cp
}

// ensureStepIndex lazily builds the stepID-to-index map on first access.
func (s *RunSnapshot) ensureStepIndex() {
	if s.stepIndex != nil {
		return
	}
	s.stepIndex = make(map[string]int, len(s.Steps))
	for i := range s.Steps {
		s.stepIndex[s.Steps[i].StepID] = i
	}
}

// invalidateStepIndex forces a rebuild on the next FindStep call.
func (s *RunSnapshot) invalidateStepIndex() { s.stepIndex = nil }

// CompletedSteps counts how many steps have StepStatusCompleted.
func (s *RunSnapshot) CompletedSteps() int {
	n := 0
	for i := range s.Steps {
		if s.Steps[i].Status == StepStatusCompleted {
			n++
		}
	}
	return n
}

// FindStep returns the step with the given ID, or nil.
func (s *RunSnapshot) FindStep(stepID string) *Step {
	s.ensureStepIndex()
	if idx, ok := s.stepIndex[stepID]; ok && idx < len(s.Steps) {
		return &s.Steps[idx]
	}
	return nil
}

// NextExecutableStep returns the first step that is pending and has all
// dependencies completed, or nil if no step is ready.
func (s *RunSnapshot) NextExecutableStep() *Step {
	completed := make(map[string]bool, len(s.Steps))
	for i := range s.Steps {
		if s.Steps[i].Status == StepStatusCompleted {
			completed[s.Steps[i].StepID] = true
		}
	}
	for i := range s.Steps {
		st := &s.Steps[i]
		if st.Status != StepStatusPending {
			continue
		}
		ready := true
		for _, dep := range st.DependsOn {
			if !completed[dep] {
				ready = false
				break
			}
		}
		if ready {
			return st
		}
	}
	return nil
}

// AllStepsTerminal returns true if every step is completed, failed, or interrupted.
func (s *RunSnapshot) AllStepsTerminal() bool {
	for i := range s.Steps {
		switch s.Steps[i].Status {
		case StepStatusCompleted, StepStatusFailed, StepStatusInterrupted:
			continue
		default:
			return false
		}
	}
	return true
}

// AllStepsSuccessful returns true if every step is completed.
// Unlike AllStepsTerminal, failed/interrupted steps make this return false.
func (s *RunSnapshot) AllStepsSuccessful() bool {
	for i := range s.Steps {
		if s.Steps[i].Status != StepStatusCompleted {
			return false
		}
	}
	return len(s.Steps) > 0
}

// AllCriteriaMet returns true if every acceptance criterion is met.
func (s *RunSnapshot) AllCriteriaMet() bool {
	for i := range s.AcceptanceState {
		if !s.AcceptanceState[i].Met {
			return false
		}
	}
	return true
}

// ToSummary produces a compact RunSummary for context injection.
func (s *RunSnapshot) ToSummary() RunSummary {
	var unmet []string
	for i := range s.AcceptanceState {
		if !s.AcceptanceState[i].Met {
			unmet = append(unmet, s.AcceptanceState[i].Description)
		}
	}
	summary := RunSummary{
		RunID:          s.RunID,
		Goal:           s.Goal,
		Status:         s.Status,
		TotalSteps:     len(s.Steps),
		CompletedSteps: s.CompletedSteps(),
		UnmetCriteria:  unmet,
	}
	if step := s.FindStep(s.CurrentStepID); step != nil {
		summary.CurrentStepGoal = step.Goal
		summary.CurrentStepStatus = string(step.Status)
	}
	summary.CurrentBlocker = s.CurrentBlocker
	return summary
}

// MaterializeFromJournal builds a RunSnapshot from scratch by replaying
// all journal events in sequence order.
func MaterializeFromJournal(events []JournalEvent) (*RunSnapshot, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("empty journal")
	}
	snap := &RunSnapshot{
		RunID: events[0].RunID,
		Notes: make(map[string]string),
	}
	for i := range events {
		if err := applyEvent(snap, &events[i]); err != nil {
			return nil, fmt.Errorf("apply event seq %d: %w", events[i].Seq, err)
		}
	}
	return snap, nil
}

// ApplyTail applies only events after the snapshot's LastJournalSeq.
func ApplyTail(snap *RunSnapshot, events []JournalEvent) error {
	for i := range events {
		if events[i].Seq <= snap.LastJournalSeq {
			continue
		}
		if err := applyEvent(snap, &events[i]); err != nil {
			return fmt.Errorf("apply event seq %d: %w", events[i].Seq, err)
		}
	}
	return nil
}

func applyEvent(snap *RunSnapshot, ev *JournalEvent) error {
	snap.LastJournalSeq = ev.Seq
	snap.UpdatedAt = ev.Timestamp

	switch ev.Type {
	case EventRunCreated:
		var p RunCreatedPayload
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return fmt.Errorf("unmarshal run_created: %w", err)
		}
		snap.SessionKey = p.SessionKey
		snap.OriginalRequest = p.OriginalRequest
		snap.Goal = p.Goal
		snap.SourceKind = p.SourceKind
		snap.SourceDescriptor = p.SourceDescriptor
		snap.Status = RunStatusPlanning

	case EventPlanAttached:
		var p PlanAttachedPayload
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return fmt.Errorf("unmarshal plan_attached: %w", err)
		}
		snap.Steps = p.Steps
		snap.invalidateStepIndex()
		snap.AcceptanceState = p.AcceptanceCriteria
		snap.Status = RunStatusRunning

	case EventStepStarted:
		var p StepStartedPayload
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return fmt.Errorf("unmarshal step_started: %w", err)
		}
		if step := snap.FindStep(p.StepID); step != nil {
			step.Status = StepStatusInProgress
		}
		snap.CurrentStepID = p.StepID

	case EventStepResultProposed:
		var p StepResultProposedPayload
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return fmt.Errorf("unmarshal step_result_proposed: %w", err)
		}
		if step := snap.FindStep(p.StepID); step != nil {
			step.Status = StepStatusVerifyPending
			step.Result = p.Result
			step.Evidence = p.Evidence
		}

	case EventStepValidationPassed:
		var p StepValidationPassedPayload
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return fmt.Errorf("unmarshal step_validation_passed: %w", err)
		}
		if step := snap.FindStep(p.StepID); step != nil {
			step.Status = StepStatusCompleted
		}

	case EventStepValidationFailed:
		var p StepValidationFailedPayload
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return fmt.Errorf("unmarshal step_validation_failed: %w", err)
		}
		if step := snap.FindStep(p.StepID); step != nil {
			step.Status = StepStatusFailed
			snap.CurrentBlocker = p.Result.Reason
		}

	case EventPolicyDecisionApplied:
		var p PolicyDecisionAppliedPayload
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return fmt.Errorf("unmarshal policy_decision_applied: %w", err)
		}
		applyPolicyToSnapshot(snap, p.StepID, &p.Decision)

	case EventNoteWritten:
		var p NoteWrittenPayload
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return fmt.Errorf("unmarshal note_written: %w", err)
		}
		snap.Notes[p.Key] = p.Value

	case EventRunPaused:
		snap.Status = RunStatusPaused

	case EventRunResumed:
		snap.Status = RunStatusRunning
		snap.CurrentBlocker = ""

	case EventRunCompleted:
		snap.Status = RunStatusCompleted

	case EventRunFailed:
		snap.Status = RunStatusFailed

	case EventCriterionMet:
		var p CriterionMetPayload
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return fmt.Errorf("unmarshal criterion_met: %w", err)
		}
		if p.Index >= 0 && p.Index < len(snap.AcceptanceState) {
			snap.AcceptanceState[p.Index].Met = true
			snap.AcceptanceState[p.Index].MetAt = &ev.Timestamp
		}

	case EventProjectionSynced:
		// no-op for snapshot

	default:
		// unknown event types are ignored for forward compatibility
	}

	return nil
}

func applyPolicyToSnapshot(snap *RunSnapshot, stepID string, decision *PolicyDecision) {
	step := snap.FindStep(stepID)
	switch decision.Action {
	case PolicyRetry:
		if step != nil {
			step.Status = StepStatusPending
			step.RetryCount++
		}
		snap.CurrentBlocker = ""

	case PolicyDecompose:
		// Insert new steps after the failed step.
		if step != nil {
			step.Status = StepStatusCompleted // original marked done
		}
		for i, ns := range decision.NewSteps {
			ns.Index = len(snap.Steps) + i
			snap.Steps = append(snap.Steps, ns)
		}
		snap.invalidateStepIndex()
		snap.CurrentBlocker = ""

	case PolicyChangeAgent:
		if step != nil {
			step.OwnerAgent = decision.NewAgent
			step.Status = StepStatusPending
		}
		snap.CurrentBlocker = ""

	case PolicyChangeValidator:
		if step != nil && decision.NewValidator != nil {
			step.Validator = *decision.NewValidator
			step.Status = StepStatusPending
		}
		snap.CurrentBlocker = ""

	case PolicySkip:
		if step != nil {
			step.Status = StepStatusCompleted // skip = treat as done
		}
		snap.CurrentBlocker = ""

	case PolicyAbort:
		snap.Status = RunStatusFailed
		snap.CurrentBlocker = decision.Reason

	case PolicyEscalate:
		snap.CurrentBlocker = "escalated: " + decision.Reason
	}
}

func copyStep(step Step) Step {
	cp := step
	cp.Evidence = copyEvidenceSlice(step.Evidence)
	cp.Validator = copyValidatorSpec(step.Validator)
	cp.ToolProfile = copyStringSlice(step.ToolProfile)
	cp.DependsOn = copyStringSlice(step.DependsOn)
	return cp
}

func copyValidatorSpec(spec ValidatorSpec) ValidatorSpec {
	cp := spec
	if spec.Params != nil {
		cp.Params = make(map[string]string, len(spec.Params))
		for k, v := range spec.Params {
			cp.Params[k] = v
		}
	}
	return cp
}

func copyAcceptanceCriterion(criterion AcceptanceCriterion) AcceptanceCriterion {
	cp := criterion
	cp.Validator = copyValidatorSpec(criterion.Validator)
	if criterion.MetAt != nil {
		ts := *criterion.MetAt
		cp.MetAt = &ts
	}
	return cp
}

func copyEvidenceSlice(src []Evidence) []Evidence {
	if src == nil {
		return nil
	}
	dst := make([]Evidence, len(src))
	copy(dst, src)
	return dst
}

func copyStringSlice(src []string) []string {
	if src == nil {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}
