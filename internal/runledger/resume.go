package runledger

import (
	"context"
	"strings"
	"time"
)

// ResumeCandidate represents a run that can be resumed.
type ResumeCandidate struct {
	RunID       string    `json:"run_id"`
	Goal        string    `json:"goal"`
	Status      RunStatus `json:"status"`
	LastUpdated time.Time `json:"last_updated"`
	StepSummary string    `json:"step_summary"`
}

// ResumeManager handles detection and execution of run resumption.
// Resume is always opt-in — no automatic revival.
type ResumeManager struct {
	store    RunLedgerStore
	staleTTL time.Duration // default 1h — runs older than this are not resumable
}

// NewResumeManager creates a new ResumeManager.
func NewResumeManager(store RunLedgerStore, staleTTL time.Duration) *ResumeManager {
	if staleTTL == 0 {
		staleTTL = time.Hour
	}
	return &ResumeManager{
		store:    store,
		staleTTL: staleTTL,
	}
}

// FindCandidates returns runs that can be resumed for the given session.
func (m *ResumeManager) FindCandidates(ctx context.Context, sessionKey string) ([]ResumeCandidate, error) {
	runs, err := m.store.ListRuns(ctx, 50)
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().Add(-m.staleTTL)
	var candidates []ResumeCandidate
	for _, r := range runs {
		if r.Status != RunStatusPaused {
			continue
		}
		// Get full snapshot to check session and freshness.
		snap, err := m.store.GetRunSnapshot(ctx, r.RunID)
		if err != nil {
			continue
		}
		if snap.SessionKey != sessionKey {
			continue
		}
		if snap.UpdatedAt.Before(cutoff) {
			continue
		}
		candidates = append(candidates, ResumeCandidate{
			RunID:       snap.RunID,
			Goal:        snap.Goal,
			Status:      snap.Status,
			LastUpdated: snap.UpdatedAt,
			StepSummary: buildStepSummary(snap),
		})
	}
	return candidates, nil
}

// DetectResumeIntent checks if the user's message contains resume keywords.
func DetectResumeIntent(message string) bool {
	lower := strings.ToLower(message)
	keywords := []string{
		"계속", "이어서", "resume", "continue",
		"다시 시작", "재개", "이어하", "마저",
	}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// Resume re-opens a paused run by appending a run_resumed event.
func (m *ResumeManager) Resume(ctx context.Context, runID, resumedBy string) (*RunSnapshot, error) {
	snap, err := m.store.GetRunSnapshot(ctx, runID)
	if err != nil {
		return nil, err
	}
	if snap.Status != RunStatusPaused {
		return nil, ErrRunNotPaused
	}

	err = m.store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    EventRunResumed,
		Payload: marshalPayload(RunResumedPayload{ResumedBy: resumedBy}),
	})
	if err != nil {
		return nil, err
	}

	return m.store.GetRunSnapshot(ctx, runID)
}

func buildStepSummary(snap *RunSnapshot) string {
	completed := snap.CompletedSteps()
	total := len(snap.Steps)
	current := ""
	if step := snap.FindStep(snap.CurrentStepID); step != nil {
		current = step.Goal
	}
	var b strings.Builder
	b.WriteString(strings.Repeat("=", completed))
	b.WriteString(strings.Repeat("-", total-completed))
	if current != "" {
		b.WriteString(" | current: ")
		b.WriteString(current)
	}
	return b.String()
}
