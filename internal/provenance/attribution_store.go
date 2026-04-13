package provenance

import "context"

// AttributionStore persists attribution records.
type AttributionStore interface {
	SaveAttribution(ctx context.Context, attr Attribution) error
	ListBySession(ctx context.Context, sessionKey string, limit int) ([]Attribution, error)
}
