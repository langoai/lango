package provenance

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/runledger"
)

// CheckpointService manages checkpoint creation and lifecycle.
type CheckpointService struct {
	store       CheckpointStore
	ledger      runledger.RunLedgerStore
	cfg         config.CheckpointConfig
}

// NewCheckpointService creates a new checkpoint service.
func NewCheckpointService(store CheckpointStore, ledger runledger.RunLedgerStore, cfg config.CheckpointConfig) *CheckpointService {
	return &CheckpointService{
		store:  store,
		ledger: ledger,
		cfg:    cfg,
	}
}

// CreateManual creates a manually-triggered checkpoint.
func (s *CheckpointService) CreateManual(ctx context.Context, sessionKey, runID, label string) (*Checkpoint, error) {
	if label == "" {
		return nil, ErrInvalidLabel
	}
	if runID == "" {
		return nil, ErrInvalidRunID
	}
	return s.create(ctx, sessionKey, runID, label, TriggerManual)
}

// OnJournalEvent is the append hook callback for automatic checkpoint creation.
// It checks whether the event type warrants a checkpoint based on config.
func (s *CheckpointService) OnJournalEvent(event runledger.JournalEvent) {
	ctx := context.Background()

	switch event.Type {
	case runledger.EventStepValidationPassed:
		if !s.cfg.AutoOnStepComplete {
			return
		}
		label := fmt.Sprintf("step_validated_%d", event.Seq)
		_, _ = s.create(ctx, "", event.RunID, label, TriggerStepComplete)

	case runledger.EventPolicyDecisionApplied:
		if !s.cfg.AutoOnPolicy {
			return
		}
		label := fmt.Sprintf("policy_applied_%d", event.Seq)
		_, _ = s.create(ctx, "", event.RunID, label, TriggerPolicy)
	}
}

func (s *CheckpointService) create(ctx context.Context, sessionKey, runID, label string, trigger CheckpointTrigger) (*Checkpoint, error) {
	// Resolve session key from run snapshot if not provided.
	if sessionKey == "" && s.ledger != nil {
		snap, err := s.ledger.GetRunSnapshot(ctx, runID)
		if err == nil && snap != nil {
			sessionKey = snap.SessionKey
		}
	}

	// Check max checkpoints per session.
	if sessionKey != "" && s.cfg.MaxPerSession > 0 {
		count, err := s.store.CountBySession(ctx, sessionKey)
		if err != nil {
			return nil, fmt.Errorf("count checkpoints: %w", err)
		}
		if count >= s.cfg.MaxPerSession {
			return nil, ErrMaxCheckpoints
		}
	}

	// Resolve current journal seq.
	var journalSeq int64
	if s.ledger != nil {
		events, err := s.ledger.GetJournalEvents(ctx, runID)
		if err == nil && len(events) > 0 {
			journalSeq = events[len(events)-1].Seq
		}
	}

	cp := Checkpoint{
		ID:         uuid.New().String(),
		SessionKey: sessionKey,
		RunID:      runID,
		Label:      label,
		Trigger:    trigger,
		JournalSeq: journalSeq,
		CreatedAt:  time.Now(),
	}

	if err := s.store.SaveCheckpoint(ctx, cp); err != nil {
		return nil, fmt.Errorf("save checkpoint: %w", err)
	}
	return &cp, nil
}
