// Package testutil provides shared test utilities, helpers, and mock
// implementations used across the Lango test suite.
package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/sqlitedriver"
)

var testDBSeq uint64

// NopLogger returns a no-op *zap.SugaredLogger suitable for tests.
func NopLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

// TestEntClient returns an in-memory Ent client with auto-migration.
// The client is automatically closed when the test completes.
func TestEntClient(t testing.TB) *ent.Client {
	t.Helper()
	name := strings.NewReplacer("/", "-", " ", "-", ":", "-").Replace(t.Name())
	seq := atomic.AddUint64(&testDBSeq, 1)
	db, err := sql.Open(sqlitedriver.DriverName(), sqlitedriver.MemoryDSN(fmt.Sprintf("%s-%d", name, seq)))
	if err != nil {
		t.Fatalf("open test sqlite db: %v", err)
	}
	if err := sqlitedriver.ConfigureConnection(db, false); err != nil {
		_ = db.Close()
		t.Fatalf("configure test sqlite db: %v", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		t.Fatalf("enable foreign keys for test sqlite db: %v", err)
	}
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	if err := client.Schema.Create(context.Background(), schema.WithForeignKeys(false)); err != nil {
		_ = client.Close()
		t.Fatalf("migrate test sqlite db: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

// SkipShort skips the test when running with -short flag.
func SkipShort(t testing.TB) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}
