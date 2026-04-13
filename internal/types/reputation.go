package types

import "context"

// ReputationQuerier queries peer trust scores. The score is in [0, 1].
type ReputationQuerier func(ctx context.Context, peerDID string) (float64, error)
