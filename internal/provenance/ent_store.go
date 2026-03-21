package provenance

import (
	"context"
	"encoding/json"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/langoai/lango/internal/ent"
	entpc "github.com/langoai/lango/internal/ent/provenancecheckpoint"
)

var _ CheckpointStore = (*EntCheckpointStore)(nil)

// EntCheckpointStore is an Ent-backed CheckpointStore for persistent checkpoints.
type EntCheckpointStore struct {
	client *ent.Client
}

// NewEntCheckpointStore creates a new Ent-backed CheckpointStore.
func NewEntCheckpointStore(client *ent.Client) *EntCheckpointStore {
	return &EntCheckpointStore{client: client}
}

func (s *EntCheckpointStore) SaveCheckpoint(ctx context.Context, cp Checkpoint) error {
	id, err := uuid.Parse(cp.ID)
	if err != nil {
		return fmt.Errorf("parse checkpoint id: %w", err)
	}

	var metadataStr string
	if len(cp.Metadata) > 0 {
		data, mErr := json.Marshal(cp.Metadata)
		if mErr != nil {
			return fmt.Errorf("marshal checkpoint metadata: %w", mErr)
		}
		metadataStr = string(data)
	}

	builder := s.client.ProvenanceCheckpoint.Create().
		SetID(id).
		SetLabel(cp.Label).
		SetTrigger(entpc.Trigger(cp.Trigger)).
		SetJournalSeq(cp.JournalSeq).
		SetCreatedAt(cp.CreatedAt)

	if cp.SessionKey != "" {
		builder = builder.SetSessionKey(cp.SessionKey)
	}
	if cp.RunID != "" {
		builder = builder.SetRunID(cp.RunID)
	}
	if cp.GitRef != "" {
		builder = builder.SetGitRef(cp.GitRef)
	}
	if metadataStr != "" {
		builder = builder.SetMetadata(metadataStr)
	}

	if _, err := builder.Save(ctx); err != nil {
		return fmt.Errorf("save checkpoint: %w", err)
	}
	return nil
}

func (s *EntCheckpointStore) GetCheckpoint(ctx context.Context, id string) (*Checkpoint, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("parse checkpoint id: %w", err)
	}

	row, err := s.client.ProvenanceCheckpoint.Get(ctx, parsed)
	if ent.IsNotFound(err) {
		return nil, ErrCheckpointNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get checkpoint: %w", err)
	}

	return entRowToCheckpoint(row)
}

func (s *EntCheckpointStore) ListByRun(ctx context.Context, runID string) ([]Checkpoint, error) {
	rows, err := s.client.ProvenanceCheckpoint.Query().
		Where(entpc.RunIDEQ(runID)).
		Order(entpc.ByJournalSeq()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list checkpoints by run: %w", err)
	}

	return entRowsToCheckpoints(rows)
}

func (s *EntCheckpointStore) ListBySession(ctx context.Context, sessionKey string, limit int) ([]Checkpoint, error) {
	query := s.client.ProvenanceCheckpoint.Query().
		Where(entpc.SessionKeyEQ(sessionKey)).
		Order(entpc.ByCreatedAt(sql.OrderDesc()))
	if limit > 0 {
		query = query.Limit(limit)
	}

	rows, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list checkpoints by session: %w", err)
	}

	return entRowsToCheckpoints(rows)
}

func (s *EntCheckpointStore) CountBySession(ctx context.Context, sessionKey string) (int, error) {
	count, err := s.client.ProvenanceCheckpoint.Query().
		Where(entpc.SessionKeyEQ(sessionKey)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count checkpoints by session: %w", err)
	}
	return count, nil
}

func (s *EntCheckpointStore) DeleteCheckpoint(ctx context.Context, id string) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("parse checkpoint id: %w", err)
	}

	err = s.client.ProvenanceCheckpoint.DeleteOneID(parsed).Exec(ctx)
	if ent.IsNotFound(err) {
		return ErrCheckpointNotFound
	}
	if err != nil {
		return fmt.Errorf("delete checkpoint: %w", err)
	}
	return nil
}

func entRowToCheckpoint(row *ent.ProvenanceCheckpoint) (*Checkpoint, error) {
	cp := &Checkpoint{
		ID:         row.ID.String(),
		SessionKey: row.SessionKey,
		RunID:      row.RunID,
		Label:      row.Label,
		Trigger:    CheckpointTrigger(row.Trigger),
		JournalSeq: row.JournalSeq,
		GitRef:     row.GitRef,
		CreatedAt:  row.CreatedAt,
	}

	if row.Metadata != "" {
		var meta map[string]string
		if err := json.Unmarshal([]byte(row.Metadata), &meta); err != nil {
			return nil, fmt.Errorf("unmarshal checkpoint metadata: %w", err)
		}
		cp.Metadata = meta
	}

	return cp, nil
}

func entRowsToCheckpoints(rows []*ent.ProvenanceCheckpoint) ([]Checkpoint, error) {
	result := make([]Checkpoint, 0, len(rows))
	for _, row := range rows {
		cp, err := entRowToCheckpoint(row)
		if err != nil {
			return nil, err
		}
		result = append(result, *cp)
	}
	return result, nil
}
