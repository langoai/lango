package app

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestWave28InitGatewayBuildsRouterWithoutListener(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Server.HTTPEnabled = true
	cfg.Server.WebSocketEnabled = true
	cfg.Agent.RequestTimeout = 0
	cfg.Agent.IdleTimeout = 0
	cfg.Agent.MaxRequestTimeout = 0

	server := initGateway(cfg, nil, &stubSessionStore{}, nil)
	require.NotNil(t, server)
	require.NotNil(t, server.Router())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	server.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
}

func TestWave28InitP2PEarlyExitBranchesAvoidListenerBinding(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = false
	cfg.P2P.KeyDir = filepath.Join(t.TempDir(), "disabled-keys")
	cfg.P2P.ListenAddrs = []string{"not-a-multiaddr"}
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))

	cfg.P2P.Enabled = true
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))

	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	cfg.P2P.KeyDir = filepath.Join(t.TempDir(), "invalid-listen-keys")
	wallet := &wiringP2PWallet{publicKey: ethcrypto.CompressPubkey(&key.PublicKey)}

	assert.Nil(t, initP2P(cfg, wallet, nil, nil, nil, nil, nil, nil, ""))
}
