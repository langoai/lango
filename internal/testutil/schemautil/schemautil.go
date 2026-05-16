package schemautil

import (
	"context"
	"sync"

	"entgo.io/ent/dialect/sql/schema"

	"github.com/langoai/lango/internal/ent"
)

var schemaCreateMu sync.Mutex

// CreateSchema serializes ent Atlas-backed schema creation for tests.
// Some migration paths mutate shared metadata and can panic with concurrent
// map writes when many tests bootstrap databases in parallel.
func CreateSchema(ctx context.Context, client *ent.Client) error {
	if client == nil {
		return nil
	}
	schemaCreateMu.Lock()
	defer schemaCreateMu.Unlock()
	return client.Schema.Create(ctx, schema.WithForeignKeys(false))
}
