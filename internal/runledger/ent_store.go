package runledger

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/langoai/lango/internal/ent"
	entrunjournal "github.com/langoai/lango/internal/ent/runjournal"
	entrunsnapshot "github.com/langoai/lango/internal/ent/runsnapshot"
	entrunstep "github.com/langoai/lango/internal/ent/runstep"
)

var _ RunLedgerStore = (*EntStore)(nil)

// EntStore persists RunLedger journal events and cached projections in Ent.
type EntStore struct {
	client *ent.Client
	mu     sync.Mutex
	cache  map[string]*RunSnapshot
}

// NewEntStore creates a new Ent-backed RunLedger store.
func NewEntStore(client *ent.Client) *EntStore {
	return &EntStore{
		client: client,
		cache:  make(map[string]*RunSnapshot),
	}
}

func (s *EntStore) AppendJournalEvent(ctx context.Context, event JournalEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.RunID == "" {
		return fmt.Errorf("append journal event: run_id is required")
	}
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	last, qErr := tx.RunJournal.Query().
		Where(entrunjournal.RunIDEQ(event.RunID)).
		Order(entrunjournal.BySeq(sql.OrderDesc())).
		First(ctx)
	if qErr != nil && !ent.IsNotFound(qErr) {
		err = fmt.Errorf("query latest seq: %w", qErr)
		return err
	}

	event.Seq = 1
	if last != nil {
		event.Seq = last.Seq + 1
	}

	if _, err = tx.RunJournal.Create().
		SetID(uuid.MustParse(event.ID)).
		SetRunID(event.RunID).
		SetSeq(event.Seq).
		SetType(entrunjournal.Type(event.Type)).
		SetTimestamp(event.Timestamp).
		SetPayload(string(event.Payload)).
		Save(ctx); err != nil {
		err = fmt.Errorf("create run_journal: %w", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit journal event: %w", err)
	}
	return nil
}

func (s *EntStore) GetJournalEvents(ctx context.Context, runID string) ([]JournalEvent, error) {
	rows, err := s.client.RunJournal.Query().
		Where(entrunjournal.RunIDEQ(runID)).
		Order(entrunjournal.BySeq()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query journal events %q: %w", runID, err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("run %q: no journal events", runID)
	}

	return decodeRunJournalRows(rows)
}

func (s *EntStore) GetJournalEventsSince(ctx context.Context, runID string, afterSeq int64) ([]JournalEvent, error) {
	rows, err := s.client.RunJournal.Query().
		Where(
			entrunjournal.RunIDEQ(runID),
			entrunjournal.SeqGT(afterSeq),
		).
		Order(entrunjournal.BySeq()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query journal events since %d: %w", afterSeq, err)
	}

	return decodeRunJournalRows(rows)
}

func (s *EntStore) MaterializeRunSnapshot(ctx context.Context, runID string) (*RunSnapshot, error) {
	events, err := s.GetJournalEvents(ctx, runID)
	if err != nil {
		return nil, err
	}
	return MaterializeFromJournal(events)
}

func (s *EntStore) RecordValidationResult(ctx context.Context, runID, stepID string, result ValidationResult) error {
	evType := EventStepValidationPassed
	payload := marshalPayload(StepValidationPassedPayload{
		StepID: stepID,
		Result: result,
	})
	if !result.Passed {
		evType = EventStepValidationFailed
		payload = marshalPayload(StepValidationFailedPayload{
			StepID: stepID,
			Result: result,
		})
	}
	return s.AppendJournalEvent(ctx, JournalEvent{
		RunID:   runID,
		Type:    evType,
		Payload: payload,
	})
}

func (s *EntStore) GetCachedSnapshot(ctx context.Context, runID string) (*RunSnapshot, int64, error) {
	s.mu.Lock()
	if snap, ok := s.cache[runID]; ok {
		s.mu.Unlock()
		return snap, snap.LastJournalSeq, nil
	}
	s.mu.Unlock()

	row, err := s.client.RunSnapshot.Query().
		Where(entrunsnapshot.RunIDEQ(runID)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, fmt.Errorf("query cached snapshot %q: %w", runID, err)
	}

	var snap RunSnapshot
	if err := json.Unmarshal([]byte(row.SnapshotData), &snap); err != nil {
		return nil, 0, fmt.Errorf("unmarshal snapshot %q: %w", runID, err)
	}
	snap.LastJournalSeq = row.LastJournalSeq
	if snap.RunID == "" {
		snap.RunID = runID
	}

	s.mu.Lock()
	s.cache[runID] = &snap
	s.mu.Unlock()
	return &snap, row.LastJournalSeq, nil
}

func (s *EntStore) UpdateCachedSnapshot(ctx context.Context, snapshot *RunSnapshot) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin snapshot tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	existing, qErr := tx.RunSnapshot.Query().
		Where(entrunsnapshot.RunIDEQ(snapshot.RunID)).
		Only(ctx)
	if qErr != nil && !ent.IsNotFound(qErr) {
		err = fmt.Errorf("query snapshot row: %w", qErr)
		return err
	}

	if existing == nil {
		_, err = tx.RunSnapshot.Create().
			SetRunID(snapshot.RunID).
			SetSessionKey(snapshot.SessionKey).
			SetStatus(entrunsnapshot.Status(snapshot.Status)).
			SetGoal(snapshot.Goal).
			SetSnapshotData(string(data)).
			SetLastJournalSeq(snapshot.LastJournalSeq).
			Save(ctx)
	} else {
		_, err = tx.RunSnapshot.UpdateOne(existing).
			SetSessionKey(snapshot.SessionKey).
			SetStatus(entrunsnapshot.Status(snapshot.Status)).
			SetGoal(snapshot.Goal).
			SetSnapshotData(string(data)).
			SetLastJournalSeq(snapshot.LastJournalSeq).
			Save(ctx)
	}
	if err != nil {
		err = fmt.Errorf("upsert snapshot: %w", err)
		return err
	}

	if _, err = tx.RunStep.Delete().
		Where(entrunstep.RunIDEQ(snapshot.RunID)).
		Exec(ctx); err != nil {
		err = fmt.Errorf("delete old run steps: %w", err)
		return err
	}

	for _, step := range snapshot.Steps {
		evidence, marshalErr := json.Marshal(step.Evidence)
		if marshalErr != nil {
			err = fmt.Errorf("marshal step evidence: %w", marshalErr)
			return err
		}
		validator, marshalErr := json.Marshal(step.Validator)
		if marshalErr != nil {
			err = fmt.Errorf("marshal validator spec: %w", marshalErr)
			return err
		}
		if _, err = tx.RunStep.Create().
			SetRunID(snapshot.RunID).
			SetStepID(step.StepID).
			SetStepIndex(step.Index).
			SetGoal(step.Goal).
			SetOwnerAgent(step.OwnerAgent).
			SetStatus(entrunstep.Status(step.Status)).
			SetResult(step.Result).
			SetEvidence(string(evidence)).
			SetValidatorSpec(string(validator)).
			SetRetryCount(step.RetryCount).
			SetMaxRetries(step.MaxRetries).
			Save(ctx); err != nil {
			err = fmt.Errorf("create run step %q: %w", step.StepID, err)
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit snapshot tx: %w", err)
	}

	s.mu.Lock()
	s.cache[snapshot.RunID] = snapshot
	s.mu.Unlock()
	return nil
}

func (s *EntStore) ListRuns(ctx context.Context, limit int) ([]RunSummary, error) {
	query := s.client.RunSnapshot.Query().
		Order(entrunsnapshot.ByUpdatedAt(sql.OrderDesc()))
	if limit > 0 {
		query = query.Limit(limit)
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list run snapshots: %w", err)
	}

	result := make([]RunSummary, 0, len(rows))
	for _, row := range rows {
		var snap RunSnapshot
		if err := json.Unmarshal([]byte(row.SnapshotData), &snap); err != nil {
			return nil, fmt.Errorf("unmarshal listed snapshot %q: %w", row.RunID, err)
		}
		snap.LastJournalSeq = row.LastJournalSeq
		result = append(result, snap.ToSummary())
	}
	return result, nil
}

func (s *EntStore) GetRunSnapshot(ctx context.Context, runID string) (*RunSnapshot, error) {
	cached, lastSeq, err := s.GetCachedSnapshot(ctx, runID)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		tail, err := s.GetJournalEventsSince(ctx, runID, lastSeq)
		if err != nil {
			return nil, err
		}
		if len(tail) == 0 {
			return cached, nil
		}
		if err := ApplyTail(cached, tail); err != nil {
			return nil, fmt.Errorf("apply tail: %w", err)
		}
		if err := s.UpdateCachedSnapshot(ctx, cached); err != nil {
			return nil, err
		}
		return cached, nil
	}

	snap, err := s.MaterializeRunSnapshot(ctx, runID)
	if err != nil {
		return nil, err
	}
	if err := s.UpdateCachedSnapshot(ctx, snap); err != nil {
		return nil, err
	}
	return snap, nil
}

func decodeRunJournalRows(rows []*ent.RunJournal) ([]JournalEvent, error) {
	events := make([]JournalEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, JournalEvent{
			ID:        row.ID.String(),
			RunID:     row.RunID,
			Seq:       row.Seq,
			Type:      JournalEventType(row.Type),
			Timestamp: row.Timestamp,
			Payload:   json.RawMessage(row.Payload),
		})
	}
	return events, nil
}
