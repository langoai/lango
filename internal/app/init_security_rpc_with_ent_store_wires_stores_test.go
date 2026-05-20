package app

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/testutil"
)

func TestInitSecurityRPCWithEntStoreWiresStores(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = "rpc"
	store := session.NewEntStoreWithClient(testutil.TestEntClient(t))

	crypto, keys, secrets, err := initSecurity(cfg, store, nil)

	require.NoError(t, err)
	require.NotNil(t, crypto)
	require.NotNil(t, keys)
	require.NotNil(t, secrets)
}

func TestBuildProvenanceAgentOptionsHandlesMissingMetadata(t *testing.T) {
	t.Parallel()

	pv := &provenanceValues{
		sessionTree: provenance.NewSessionTree(provenance.NewMemoryTreeStore()),
	}

	opts := buildProvenanceAgentOptions(pv, nil)

	require.Len(t, opts, 2)
	assert.Nil(t, pv.configMetadata)
}

func TestInitP2PEarlyExitBranchesAvoidRuntimeStartup(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.KeyDir = filepath.Join(t.TempDir(), "p2p-keys")
	cfg.P2P.ListenAddrs = []string{"/ip4/127.0.0.1/tcp/0"}

	cfg.P2P.Enabled = false
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))

	cfg.P2P.Enabled = true
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))

	cfg.P2P.ListenAddrs = []string{"not-a-multiaddr"}
	assert.Nil(t, initP2P(cfg, &wiringP2PWallet{}, nil, nil, nil, nil, nil, nil, ""))
}
