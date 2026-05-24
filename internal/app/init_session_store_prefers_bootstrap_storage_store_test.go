package app

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storage"
)

func TestInitSessionStorePrefersBootstrapStorageStore(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "fallback.db")

	want := &stubSessionStore{}
	var gotOptionCount int
	facade := storage.NewFacade(nil, nil, storage.WithSessionStoreFactory(
		func(opts ...session.StoreOption) (session.Store, error) {
			gotOptionCount = len(opts)
			return want, nil
		},
	))

	got, err := initSessionStore(cfg, &bootstrap.Result{Storage: facade})

	require.NoError(t, err)
	assert.Same(t, want, got)
	assert.Positive(t, gotOptionCount, "configured session limits should be forwarded as store options")
}

func TestInitSessionStoreFallsBackWhenBootstrapStorageCannotOpenStore(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "fallback.db")

	var called bool
	facade := storage.NewFacade(nil, nil, storage.WithSessionStoreFactory(
		func(opts ...session.StoreOption) (session.Store, error) {
			called = true
			return nil, fmt.Errorf("bootstrap session store unavailable")
		},
	))

	got, err := initSessionStore(cfg, &bootstrap.Result{Storage: facade})
	t.Cleanup(func() {
		if got != nil {
			_ = got.Close()
		}
	})

	require.NoError(t, err)
	assert.True(t, called)
	assert.NotNil(t, got)
}

func TestInitSessionStoreReturnsFallbackOpenError(t *testing.T) {
	t.Parallel()

	dbPathAsDir := filepath.Join(t.TempDir(), "session.db")
	require.NoError(t, os.Mkdir(dbPathAsDir, 0o700))

	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = dbPathAsDir

	got, err := initSessionStore(cfg, nil)

	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "session store:")
	assert.ErrorContains(t, err, "unable to open database file")
}
