package schemautil

import (
	"context"

	"entgo.io/ent/dialect/sql/schema"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/storeutil/schemaexec"
)

// RunExclusive serializes test schema operations that read or mutate ent's
// generated migration metadata.
func RunExclusive(fn func() error) error {
	return schemaexec.RunExclusive(fn)
}

// CreateSchema serializes ent Atlas-backed schema creation for tests.
// Some migration paths mutate shared metadata and can panic with concurrent
// map writes when many tests bootstrap databases in parallel.
func CreateSchema(ctx context.Context, client *ent.Client) error {
	if client == nil {
		return nil
	}
	return RunExclusive(func() error {
		return client.Schema.Create(ctx, schema.WithForeignKeys(false))
	})
}
