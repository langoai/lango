// Package testutil provides shared test utilities, helpers, and mock
// implementations used across the Lango test suite.
package testutil

import (
	"testing"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/enttest"

	_ "github.com/mattn/go-sqlite3"
)

// NopLogger returns a no-op *zap.SugaredLogger suitable for tests.
func NopLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

// TestEntClient returns an in-memory Ent client with auto-migration.
// The client is automatically closed when the test completes.
func TestEntClient(t testing.TB) *ent.Client {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
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
