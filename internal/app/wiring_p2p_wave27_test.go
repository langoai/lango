package app

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/config"
)

func TestWave27InitP2PMissingWalletShortCircuitsBeforeNodeSetup(t *testing.T) {
	t.Parallel()

	keyDir := filepath.Join(t.TempDir(), "p2p-keys")
	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.KeyDir = keyDir
	cfg.P2P.ListenAddrs = []string{"not-a-valid-multiaddr"}

	got := initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, "")

	assert.Nil(t, got)
	assert.NoDirExists(t, keyDir, "missing wallet should skip node/key setup entirely")
}
