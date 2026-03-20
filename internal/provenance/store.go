package provenance

import "context"

// CheckpointStore is the persistence interface for provenance checkpoints.
type CheckpointStore interface {
	// SaveCheckpoint persists a new checkpoint.
	SaveCheckpoint(ctx context.Context, cp Checkpoint) error

	// GetCheckpoint returns a checkpoint by ID.
	GetCheckpoint(ctx context.Context, id string) (*Checkpoint, error)

	// ListByRun returns checkpoints for a run, ordered by journal_seq asc.
	ListByRun(ctx context.Context, runID string) ([]Checkpoint, error)

	// ListBySession returns checkpoints for a session, ordered by created_at desc.
	ListBySession(ctx context.Context, sessionKey string, limit int) ([]Checkpoint, error)

	// CountBySession returns the number of checkpoints for a session.
	CountBySession(ctx context.Context, sessionKey string) (int, error)

	// DeleteCheckpoint removes a checkpoint by ID.
	DeleteCheckpoint(ctx context.Context, id string) error
}
