package testutil

import (
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/storage"
)

// FakeCfgLoader returns a cfgLoader func that always returns the given config.
func FakeCfgLoader(cfg *config.Config) func() (*config.Config, error) {
	return func() (*config.Config, error) {
		return cfg, nil
	}
}

// FailCfgLoader returns a cfgLoader func that always returns the given error.
func FailCfgLoader(err error) func() (*config.Config, error) {
	return func() (*config.Config, error) {
		return nil, err
	}
}

// FakeBootLoader returns a bootLoader func backed by a test storage facade.
func FakeBootLoader(t testing.TB, cfg *config.Config) func() (*bootstrap.Result, error) {
	t.Helper()

	return func() (*bootstrap.Result, error) {
		client := TestEntClient(t)
		return &bootstrap.Result{
			Config:  cfg,
			Storage: storage.NewFacade(nil, nil, storage.WithEntClient(client)),
		}, nil
	}
}

// FailBootLoader returns a bootLoader func that always returns the given error.
func FailBootLoader(err error) func() (*bootstrap.Result, error) {
	return func() (*bootstrap.Result, error) {
		return nil, err
	}
}
