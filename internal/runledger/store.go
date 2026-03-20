package runledger

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RunLedgerStore is the persistence interface for the RunLedger.
// The journal is the single source of truth; snapshots are cached projections.
type RunLedgerStore interface {
	// AppendJournalEvent appends an event to the journal. Seq is auto-assigned.
	AppendJournalEvent(ctx context.Context, event JournalEvent) error

	// GetJournalEvents returns all events for a run, ordered by seq.
	GetJournalEvents(ctx context.Context, runID string) ([]JournalEvent, error)

	// GetJournalEventsSince returns events with seq > afterSeq.
	GetJournalEventsSince(ctx context.Context, runID string, afterSeq int64) ([]JournalEvent, error)

	// MaterializeRunSnapshot replays the full journal to build a snapshot.
	MaterializeRunSnapshot(ctx context.Context, runID string) (*RunSnapshot, error)

	// RecordValidationResult appends a step_validation_passed or step_validation_failed event.
	RecordValidationResult(ctx context.Context, runID, stepID string, result ValidationResult) error

	// GetCachedSnapshot returns the last cached snapshot and its seq, or nil if uncached.
	GetCachedSnapshot(ctx context.Context, runID string) (*RunSnapshot, int64, error)

	// UpdateCachedSnapshot stores/updates the snapshot cache.
	UpdateCachedSnapshot(ctx context.Context, snapshot *RunSnapshot) error

	// ListRuns returns run IDs with their current status, ordered by creation time desc.
	ListRuns(ctx context.Context, limit int) ([]RunSummary, error)

	// GetRunSnapshot returns the most up-to-date snapshot, using cache + tail replay.
	GetRunSnapshot(ctx context.Context, runID string) (*RunSnapshot, error)

	// ListRunSummariesBySession returns recent run summaries for a session.
	ListRunSummariesBySession(ctx context.Context, sessionKey string, limit int) ([]RunSummary, error)
}

// MemoryStore is an in-memory implementation of RunLedgerStore for testing
// and for the initial shadow phase.
type MemoryStore struct {
	mu       sync.RWMutex
	journals map[string][]JournalEvent // runID -> events
	seqs     map[string]int64          // runID -> next seq
	cache    map[string]*RunSnapshot   // runID -> cached snapshot
}

// NewMemoryStore creates a new in-memory RunLedgerStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		journals: make(map[string][]JournalEvent),
		seqs:     make(map[string]int64),
		cache:    make(map[string]*RunSnapshot),
	}
}

func (m *MemoryStore) AppendJournalEvent(_ context.Context, event JournalEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	seq := m.seqs[event.RunID]
	seq++
	event.Seq = seq
	m.seqs[event.RunID] = seq

	m.journals[event.RunID] = append(m.journals[event.RunID], event)
	return nil
}

func (m *MemoryStore) GetJournalEvents(_ context.Context, runID string) ([]JournalEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := m.journals[runID]
	if len(events) == 0 {
		return nil, fmt.Errorf("run %q: no journal events", runID)
	}
	cp := make([]JournalEvent, len(events))
	copy(cp, events)
	return cp, nil
}

func (m *MemoryStore) GetJournalEventsSince(_ context.Context, runID string, afterSeq int64) ([]JournalEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := m.journals[runID]
	var tail []JournalEvent
	for i := range events {
		if events[i].Seq > afterSeq {
			tail = append(tail, events[i])
		}
	}
	return tail, nil
}

func (m *MemoryStore) MaterializeRunSnapshot(ctx context.Context, runID string) (*RunSnapshot, error) {
	events, err := m.GetJournalEvents(ctx, runID)
	if err != nil {
		return nil, err
	}
	return MaterializeFromJournal(events)
}

func (m *MemoryStore) RecordValidationResult(ctx context.Context, runID, stepID string, result ValidationResult) error {
	evType := EventStepValidationPassed
	if !result.Passed {
		evType = EventStepValidationFailed
	}

	var payload interface{}
	if result.Passed {
		payload = StepValidationPassedPayload{StepID: stepID, Result: result}
	} else {
		payload = StepValidationFailedPayload{StepID: stepID, Result: result}
	}

	return m.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    evType,
		Payload: marshalPayload(payload),
	})
}

func (m *MemoryStore) GetCachedSnapshot(_ context.Context, runID string) (*RunSnapshot, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snap, ok := m.cache[runID]
	if !ok {
		return nil, 0, nil
	}
	return snap, snap.LastJournalSeq, nil
}

func (m *MemoryStore) UpdateCachedSnapshot(_ context.Context, snapshot *RunSnapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache[snapshot.RunID] = snapshot
	return nil
}

func (m *MemoryStore) ListRuns(_ context.Context, limit int) ([]RunSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type runInfo struct {
		runID     string
		createdAt time.Time
	}
	var runs []runInfo
	for runID, events := range m.journals {
		if len(events) > 0 {
			runs = append(runs, runInfo{runID: runID, createdAt: events[0].Timestamp})
		}
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].createdAt.After(runs[j].createdAt)
	})

	if limit > 0 && len(runs) > limit {
		runs = runs[:limit]
	}

	result := make([]RunSummary, 0, len(runs))
	for _, r := range runs {
		snap, err := MaterializeFromJournal(m.journals[r.runID])
		if err != nil {
			continue
		}
		result = append(result, snap.ToSummary())
	}
	return result, nil
}

func (m *MemoryStore) GetRunSnapshot(ctx context.Context, runID string) (*RunSnapshot, error) {
	// Try cache first.
	cached, lastSeq, err := m.GetCachedSnapshot(ctx, runID)
	if err != nil {
		return nil, err
	}

	if cached != nil {
		// Apply journal tail.
		tail, err := m.GetJournalEventsSince(ctx, runID, lastSeq)
		if err != nil {
			return nil, err
		}
		if len(tail) == 0 {
			return cached, nil
		}
		if err := ApplyTail(cached, tail); err != nil {
			return nil, fmt.Errorf("apply tail: %w", err)
		}
		_ = m.UpdateCachedSnapshot(ctx, cached)
		return cached, nil
	}

	// Full materialize.
	snap, err := m.MaterializeRunSnapshot(ctx, runID)
	if err != nil {
		return nil, err
	}
	_ = m.UpdateCachedSnapshot(ctx, snap)
	return snap, nil
}

func (m *MemoryStore) ListRunSummariesBySession(ctx context.Context, sessionKey string, limit int) ([]RunSummary, error) {
	runs, err := m.ListRuns(ctx, limit)
	if err != nil {
		return nil, err
	}
	var filtered []RunSummary
	for _, run := range runs {
		snap, snapErr := m.GetRunSnapshot(ctx, run.RunID)
		if snapErr != nil {
			continue
		}
		if snap.SessionKey == sessionKey {
			filtered = append(filtered, run)
		}
	}
	return filtered, nil
}
