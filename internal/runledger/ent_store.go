package runledger

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	locks  sync.Map
	cache  sync.Map
	opts   StoreOptions
}

// NewEntStore creates a new Ent-backed RunLedger store.
func NewEntStore(client *ent.Client, opts ...StoreOption) *EntStore {
	return &EntStore{
		client: client,
		opts:   applyStoreOptions(opts),
	}
}

func (s *EntStore) AppendJournalEvent(ctx context.Context, event JournalEvent) error {
	if event.RunID == "" {
		return fmt.Errorf("append journal event: run_id is required")
	}
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	var nextSeq int64
	for attempt := 0; attempt < 10; attempt++ {
		tx, err := s.client.Tx(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}

		commitErr := func() error {
			last, qErr := tx.RunJournal.Query().
				Where(entrunjournal.RunIDEQ(event.RunID)).
				Order(entrunjournal.BySeq(sql.OrderDesc())).
				First(ctx)
			if qErr != nil && !ent.IsNotFound(qErr) {
				return fmt.Errorf("query latest seq: %w", qErr)
			}

			nextSeq = int64(1)
			if last != nil {
				nextSeq = last.Seq + 1
			}

			if _, err := tx.RunJournal.Create().
				SetID(uuid.MustParse(event.ID)).
				SetRunID(event.RunID).
				SetSeq(nextSeq).
				SetType(entrunjournal.Type(event.Type)).
				SetTimestamp(event.Timestamp).
				SetPayload(string(event.Payload)).
				Save(ctx); err != nil {
				return fmt.Errorf("create run_journal: %w", err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("commit journal event: %w", err)
			}
			return nil
		}()

		if commitErr == nil {
			event.Seq = nextSeq
			// AppendHook is called after the transaction commits to avoid
			// deadlocks when the hook calls back into the store.
			if s.opts.AppendHook != nil {
				s.opts.AppendHook(event)
			}
			return nil
		}
		_ = tx.Rollback()
		if !shouldRetryAppendJournalError(commitErr) || attempt == 9 {
			return commitErr
		}
		time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond)
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
	lock := s.runLock(runID)
	lock.Lock()
	if cached, ok := s.cache.Load(runID); ok {
		snap := cached.(*RunSnapshot)
		lock.Unlock()
		return snap, snap.LastJournalSeq, nil
	}
	lock.Unlock()

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

	lock.Lock()
	s.cache.Store(runID, &snap)
	lock.Unlock()
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

	lock := s.runLock(snapshot.RunID)
	lock.Lock()
	s.cache.Store(snapshot.RunID, snapshot.DeepCopy())
	lock.Unlock()
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
			return cached.DeepCopy(), nil
		}
		snap := cached.DeepCopy()
		if err := ApplyTail(snap, tail); err != nil {
			return nil, fmt.Errorf("apply tail: %w", err)
		}
		if err := s.UpdateCachedSnapshot(ctx, snap); err != nil {
			return nil, err
		}
		return snap, nil
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

func (s *EntStore) ListRunSummariesBySession(ctx context.Context, sessionKey string, limit int) ([]RunSummary, error) {
	query := s.client.RunSnapshot.Query().
		Where(entrunsnapshot.SessionKeyEQ(sessionKey)).
		Order(entrunsnapshot.ByUpdatedAt(sql.OrderDesc()))
	if limit > 0 {
		query = query.Limit(limit)
	}

	rows, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list run summaries by session: %w", err)
	}

	result := make([]RunSummary, 0, len(rows))
	for _, row := range rows {
		var snap RunSnapshot
		if err := json.Unmarshal([]byte(row.SnapshotData), &snap); err != nil {
			return nil, fmt.Errorf("unmarshal session snapshot %q: %w", row.RunID, err)
		}
		snap.LastJournalSeq = row.LastJournalSeq
		result = append(result, snap.ToSummary())
	}
	return result, nil
}

func (s *EntStore) MaxJournalSeqForSession(ctx context.Context, sessionKey string) (int64, error) {
	rows, err := s.client.RunSnapshot.Query().
		Where(entrunsnapshot.SessionKeyEQ(sessionKey)).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("query max journal seq for session: %w", err)
	}

	var maxSeq int64
	for _, row := range rows {
		if row.LastJournalSeq > maxSeq {
			maxSeq = row.LastJournalSeq
		}
	}
	return maxSeq, nil
}

func (s *EntStore) PruneOldRuns(ctx context.Context, maxKeep int) error {
	if maxKeep <= 0 {
		return nil
	}

	total, err := s.client.RunSnapshot.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("count run snapshots: %w", err)
	}
	excess := total - maxKeep
	if excess <= 0 {
		return nil
	}

	rows, err := s.client.RunSnapshot.Query().
		Where(
			entrunsnapshot.StatusIn(
				entrunsnapshot.Status(string(RunStatusCompleted)),
				entrunsnapshot.Status(string(RunStatusFailed)),
			),
		).
		Order(entrunsnapshot.ByCreatedAt(sql.OrderAsc())).
		Limit(excess).
		All(ctx)
	if err != nil {
		return fmt.Errorf("query prune candidates: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin prune tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	runIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		runIDs = append(runIDs, row.RunID)
	}

	if _, err = tx.RunJournal.Delete().Where(entrunjournal.RunIDIn(runIDs...)).Exec(ctx); err != nil {
		err = fmt.Errorf("delete run journals: %w", err)
		return err
	}
	if _, err = tx.RunStep.Delete().Where(entrunstep.RunIDIn(runIDs...)).Exec(ctx); err != nil {
		err = fmt.Errorf("delete run steps: %w", err)
		return err
	}
	if _, err = tx.RunSnapshot.Delete().Where(entrunsnapshot.RunIDIn(runIDs...)).Exec(ctx); err != nil {
		err = fmt.Errorf("delete run snapshots: %w", err)
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit prune tx: %w", err)
	}

	for _, runID := range runIDs {
		s.cache.Delete(runID)
	}

	return nil
}

func (s *EntStore) runLock(runID string) *sync.Mutex {
	lock, _ := s.locks.LoadOrStore(runID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// SetAppendHook adds an append hook, chaining with any existing hook.
// Must be called before concurrent AppendJournalEvent calls (e.g., during app boot).
func (s *EntStore) SetAppendHook(h func(JournalEvent)) {
	prev := s.opts.AppendHook
	if prev == nil {
		s.opts.AppendHook = h
	} else {
		s.opts.AppendHook = func(e JournalEvent) { prev(e); h(e) }
	}
}

func shouldRetryAppendJournalError(err error) bool {
	if ent.IsConstraintError(err) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "database table is locked") ||
		strings.Contains(msg, "database is locked")
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
