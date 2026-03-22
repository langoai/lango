package provenance

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/langoai/lango/internal/ent"
	entattr "github.com/langoai/lango/internal/ent/provenanceattribution"
)

var _ AttributionStore = (*EntAttributionStore)(nil)

// EntAttributionStore persists attribution rows in Ent.
type EntAttributionStore struct {
	client *ent.Client
}

// NewEntAttributionStore creates a new Ent-backed attribution store.
func NewEntAttributionStore(client *ent.Client) *EntAttributionStore {
	return &EntAttributionStore{client: client}
}

func (s *EntAttributionStore) SaveAttribution(ctx context.Context, attr Attribution) error {
	if attr.ID == "" {
		attr.ID = uuid.New().String()
	}
	id, err := uuid.Parse(attr.ID)
	if err != nil {
		return fmt.Errorf("parse attribution id: %w", err)
	}

	existing, err := s.client.ProvenanceAttribution.Get(ctx, id)
	if err == nil && existing != nil {
		return nil
	}
	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("get attribution: %w", err)
	}

	builder := s.client.ProvenanceAttribution.Create().
		SetID(id).
		SetSessionKey(attr.SessionKey).
		SetAuthorType(string(attr.AuthorType)).
		SetAuthorID(attr.AuthorID).
		SetSource(string(attr.Source)).
		SetLinesAdded(attr.LinesAdded).
		SetLinesRemoved(attr.LinesRemoved).
		SetCreatedAt(attr.CreatedAt)
	if attr.RunID != "" {
		builder = builder.SetRunID(attr.RunID)
	}
	if attr.WorkspaceID != "" {
		builder = builder.SetWorkspaceID(attr.WorkspaceID)
	}
	if attr.FilePath != "" {
		builder = builder.SetFilePath(attr.FilePath)
	}
	if attr.CommitHash != "" {
		builder = builder.SetCommitHash(attr.CommitHash)
	}
	if attr.StepID != "" {
		builder = builder.SetStepID(attr.StepID)
	}
	if _, err := builder.Save(ctx); err != nil {
		return fmt.Errorf("save attribution: %w", err)
	}
	return nil
}

func (s *EntAttributionStore) ListBySession(ctx context.Context, sessionKey string, limit int) ([]Attribution, error) {
	query := s.client.ProvenanceAttribution.Query().
		Where(entattr.SessionKeyEQ(sessionKey)).
		Order(entattr.ByCreatedAt(sql.OrderDesc()))
	if limit > 0 {
		query = query.Limit(limit)
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list attributions: %w", err)
	}
	out := make([]Attribution, 0, len(rows))
	for _, row := range rows {
		out = append(out, entRowToAttribution(row))
	}
	return out, nil
}

func entRowToAttribution(row *ent.ProvenanceAttribution) Attribution {
	return Attribution{
		ID:           row.ID.String(),
		SessionKey:   row.SessionKey,
		RunID:        row.RunID,
		WorkspaceID:  row.WorkspaceID,
		AuthorType:   AuthorType(row.AuthorType),
		AuthorID:     row.AuthorID,
		FilePath:     row.FilePath,
		CommitHash:   row.CommitHash,
		StepID:       row.StepID,
		Source:       AttributionSource(row.Source),
		LinesAdded:   row.LinesAdded,
		LinesRemoved: row.LinesRemoved,
		CreatedAt:    row.CreatedAt,
	}
}
