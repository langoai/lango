package p2p

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/security"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewNodeReusesPersistedKeyAndRejectsInvalidListenAddr(t *testing.T) {
	t.Parallel()

	keyDir := t.TempDir()
	cfg := newNodeReusesPersistedKeyAndRejectsInvalidListenAddrNodeConfig{keyDir: keyDir}
	node, err := NewNode(cfg, zap.NewNop().Sugar(), nil)
	require.NoError(t, err)
	peerID := node.PeerID()
	require.NoError(t, node.Stop())

	reloaded, err := NewNode(cfg, zap.NewNop().Sugar(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, reloaded.Stop()) })
	assert.Equal(t, peerID, reloaded.PeerID())

	invalid, err := NewNode(
		newNodeReusesPersistedKeyAndRejectsInvalidListenAddrNodeConfig{keyDir: t.TempDir(), listenAddrs: []string{"/invalid"}},
		zap.NewNop().Sugar(),
		nil,
	)
	require.Error(t, err)
	assert.Nil(t, invalid)
	assert.ErrorContains(t, err, "new libp2p host")
}

func TestNodeStartStopSkipsInvalidBootstrapAndExposesAccessors(t *testing.T) {
	t.Parallel()

	node, err := NewNode(newNodeReusesPersistedKeyAndRejectsInvalidListenAddrNodeConfig{
		keyDir:         t.TempDir(),
		bootstrapPeers: []string{"not-a-multiaddr", "/ip4/127.0.0.1/tcp/1"},
	}, zap.NewNop().Sugar(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = node.Stop() })

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, node.Start(ctx, &wg))
	wg.Wait()

	assert.NotNil(t, node.Host())
	assert.NotNil(t, node.dht)
	assert.NotEmpty(t, node.PeerID())
	assert.NotEmpty(t, node.Multiaddrs())
	assert.Empty(t, node.ConnectedPeers())
	assert.Nil(t, node.mdnsSvc)

	require.NoError(t, node.Stop())
}

func TestLoadOrGenerateKeyCoversSecretLegacyAndInvalidPaths(t *testing.T) {
	t.Parallel()

	log := zap.NewNop().Sugar()

	t.Run("invalid legacy key", func(t *testing.T) {
		t.Parallel()

		keyDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(keyDir, nodeKeyFile), []byte("not a key"), 0o600))
		key, err := loadOrGenerateKey(keyDir, nil, log)
		require.Error(t, err)
		assert.Nil(t, key)
		assert.ErrorContains(t, err, "unmarshal node key")
	})

	t.Run("secrets store key wins over legacy file", func(t *testing.T) {
		t.Parallel()

		secrets := newNewNodeReusesPersistedKeyAndRejectsInvalidListenAddrSecretsStore(t)
		secretKey, _, err := crypto.GenerateEd25519Key(nil)
		require.NoError(t, err)
		secretRaw, err := crypto.MarshalPrivateKey(secretKey)
		require.NoError(t, err)
		require.NoError(t, secrets.Store(context.Background(), nodeKeySecret, secretRaw))

		legacyKey, _, err := crypto.GenerateEd25519Key(nil)
		require.NoError(t, err)
		legacyRaw, err := crypto.MarshalPrivateKey(legacyKey)
		require.NoError(t, err)
		keyDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(keyDir, nodeKeyFile), legacyRaw, 0o600))

		loaded, err := loadOrGenerateKey(keyDir, secrets, log)
		require.NoError(t, err)
		loadedRaw, err := crypto.MarshalPrivateKey(loaded)
		require.NoError(t, err)
		assert.True(t, bytes.Equal(secretRaw, loadedRaw))
		assert.FileExists(t, filepath.Join(keyDir, nodeKeyFile))
	})

	t.Run("new key stored in secrets without plaintext file", func(t *testing.T) {
		t.Parallel()

		secrets := newNewNodeReusesPersistedKeyAndRejectsInvalidListenAddrSecretsStore(t)
		keyDir := t.TempDir()
		loaded, err := loadOrGenerateKey(keyDir, secrets, log)
		require.NoError(t, err)
		loadedRaw, err := crypto.MarshalPrivateKey(loaded)
		require.NoError(t, err)
		stored, err := secrets.Get(context.Background(), nodeKeySecret)
		require.NoError(t, err)
		assert.True(t, bytes.Equal(loadedRaw, stored))
		assert.NoFileExists(t, filepath.Join(keyDir, nodeKeyFile))
	})
}

func TestMigrateKeyToSecretsStoresWhenLegacyFileAlreadyGone(t *testing.T) {
	t.Parallel()

	secrets := newNewNodeReusesPersistedKeyAndRejectsInvalidListenAddrSecretsStore(t)
	key, _, err := crypto.GenerateEd25519Key(nil)
	require.NoError(t, err)
	raw, err := crypto.MarshalPrivateKey(key)
	require.NoError(t, err)
	missingPath := filepath.Join(t.TempDir(), nodeKeyFile)

	require.NoError(t, migrateKeyToSecrets(secrets, raw, missingPath, zap.NewNop().Sugar()))
	stored, err := secrets.Get(context.Background(), nodeKeySecret)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(raw, stored))
}

func TestMigrateKeyToSecretsKeepsUnreadableLegacyDirectoryAfterStore(t *testing.T) {
	t.Parallel()

	secrets := newNewNodeReusesPersistedKeyAndRejectsInvalidListenAddrSecretsStore(t)
	key, _, err := crypto.GenerateEd25519Key(nil)
	require.NoError(t, err)
	raw, err := crypto.MarshalPrivateKey(key)
	require.NoError(t, err)

	legacyPath := filepath.Join(t.TempDir(), nodeKeyFile)
	require.NoError(t, os.Mkdir(legacyPath, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(legacyPath, "child"), []byte("legacy"), 0o600))

	require.NoError(t, migrateKeyToSecrets(secrets, raw, legacyPath, zap.NewNop().Sugar()))
	stored, err := secrets.Get(context.Background(), nodeKeySecret)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(raw, stored))
	assert.DirExists(t, legacyPath)
}

func TestExpandHomeAndMDNSConnectErrorBranches(t *testing.T) {
	t.Parallel()

	home, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Equal(t, home, expandHome("~"))
	assert.Equal(t, filepath.Join(home, "nested", "keydir"), expandHome("~/nested/keydir"))

	owner := &newNodeReusesPersistedKeyAndRejectsInvalidListenAddrMDNSHost{id: peer.ID("owner"), connectErr: errors.New("dial refused")}
	notifee := &mdnsNotifee{
		host:   owner,
		ctx:    context.Background(),
		logger: zap.NewNop().Sugar(),
	}

	notifee.HandlePeerFound(peer.AddrInfo{ID: owner.ID()})
	assert.Zero(t, owner.connectCalls)

	discovered := peer.AddrInfo{ID: peer.ID("other")}
	notifee.HandlePeerFound(discovered)
	assert.Equal(t, 1, owner.connectCalls)
	assert.Equal(t, discovered, owner.connected[0])
}

type newNodeReusesPersistedKeyAndRejectsInvalidListenAddrNodeConfig struct {
	keyDir         string
	listenAddrs    []string
	bootstrapPeers []string
	enableMDNS     bool
}

func (cfg newNodeReusesPersistedKeyAndRejectsInvalidListenAddrNodeConfig) GetKeyDir() string {
	return cfg.keyDir
}

func (newNodeReusesPersistedKeyAndRejectsInvalidListenAddrNodeConfig) GetMaxPeers() int {
	return 10
}

func (cfg newNodeReusesPersistedKeyAndRejectsInvalidListenAddrNodeConfig) GetListenAddrs() []string {
	if cfg.listenAddrs != nil {
		return cfg.listenAddrs
	}
	return []string{"/ip4/127.0.0.1/tcp/0"}
}

func (newNodeReusesPersistedKeyAndRejectsInvalidListenAddrNodeConfig) GetEnableRelay() bool {
	return false
}

func (cfg newNodeReusesPersistedKeyAndRejectsInvalidListenAddrNodeConfig) GetBootstrapPeers() []string {
	return cfg.bootstrapPeers
}

func (cfg newNodeReusesPersistedKeyAndRejectsInvalidListenAddrNodeConfig) GetEnableMDNS() bool {
	return cfg.enableMDNS
}

type newNodeReusesPersistedKeyAndRejectsInvalidListenAddrMDNSHost struct {
	nodeAccessorsOnNotStartedNodeMDNSHost
	id           peer.ID
	connectErr   error
	connectCalls int
	connected    []peer.AddrInfo
}

func (h *newNodeReusesPersistedKeyAndRejectsInvalidListenAddrMDNSHost) ID() peer.ID {
	return h.id
}

func (h *newNodeReusesPersistedKeyAndRejectsInvalidListenAddrMDNSHost) Connect(_ context.Context, pi peer.AddrInfo) error {
	h.connectCalls++
	h.connected = append(h.connected, pi)
	return h.connectErr
}

type newNodeReusesPersistedKeyAndRejectsInvalidListenAddrCryptoProvider struct{}

func (newNodeReusesPersistedKeyAndRejectsInvalidListenAddrCryptoProvider) Sign(_ context.Context, _ string, payload []byte) ([]byte, error) {
	return append([]byte("sig:"), payload...), nil
}

func (newNodeReusesPersistedKeyAndRejectsInvalidListenAddrCryptoProvider) Encrypt(_ context.Context, _ string, plaintext []byte) ([]byte, error) {
	return append([]byte("enc:"), plaintext...), nil
}

func (newNodeReusesPersistedKeyAndRejectsInvalidListenAddrCryptoProvider) Decrypt(_ context.Context, _ string, ciphertext []byte) ([]byte, error) {
	return bytes.TrimPrefix(ciphertext, []byte("enc:")), nil
}

func newNewNodeReusesPersistedKeyAndRejectsInvalidListenAddrSecretsStore(t *testing.T) *security.SecretsStore {
	t.Helper()

	client := enttest.Open(t, "sqlite3", "file:initSecurityDeterministicDisabledAndErrorBranches4?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	registry := security.NewKeyRegistry(client)
	_, err := registry.RegisterKey(
		context.Background(),
		"initSecurityDeterministicDisabledAndErrorBranches4-default",
		"local",
		security.KeyTypeEncryption,
	)
	require.NoError(t, err)

	return security.NewSecretsStore(client, registry, newNodeReusesPersistedKeyAndRejectsInvalidListenAddrCryptoProvider{})
}
