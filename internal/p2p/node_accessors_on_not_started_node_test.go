package p2p

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/security"

	_ "github.com/mattn/go-sqlite3"
)

func TestNodeAccessorsOnNotStartedNode(t *testing.T) {
	t.Parallel()

	node, err := NewNode(nodeAccessorsOnNotStartedNodeNodeConfig{keyDir: t.TempDir()}, zap.NewNop().Sugar(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, node.Stop()) })

	assert.NotEmpty(t, node.PeerID())
	assert.NotNil(t, node.Host())
	assert.NotEmpty(t, node.Multiaddrs())
	assert.Empty(t, node.ConnectedPeers())

	ps1, err := node.PubSub()
	require.NoError(t, err)
	ps2, err := node.PubSub()
	require.NoError(t, err)
	assert.Same(t, ps1, ps2)
}

func TestConnectedPeersDeduplicatesConnections(t *testing.T) {
	t.Parallel()

	peerA := peer.ID("peer-a")
	peerB := peer.ID("peer-b")
	net := &nodeAccessorsOnNotStartedNodeNetwork{
		conns: []network.Conn{
			nodeAccessorsOnNotStartedNodeConn{remote: peerA},
			nodeAccessorsOnNotStartedNodeConn{remote: peerA},
			nodeAccessorsOnNotStartedNodeConn{remote: peerB},
		},
	}
	node := &Node{host: &nodeAccessorsOnNotStartedNodeNetworkHost{network: net}}

	assert.Equal(t, []peer.ID{peerA, peerB}, node.ConnectedPeers())
}

func TestNodeStopCancelsAndReturnsHostCloseError(t *testing.T) {
	t.Parallel()

	var canceled bool
	host := &nodeAccessorsOnNotStartedNodeCloseHost{closeErr: errors.New("close failed")}
	node := &Node{
		host:   host,
		logger: zap.NewNop().Sugar(),
		cancel: func() { canceled = true },
	}

	err := node.Stop()
	require.Error(t, err)
	assert.ErrorContains(t, err, "close host")
	assert.ErrorContains(t, err, "close failed")
	assert.True(t, canceled)
	assert.True(t, host.closed)
}

func TestNodeSetStreamHandlerServesLocalStream(t *testing.T) {
	t.Parallel()

	node, err := NewNode(nodeAccessorsOnNotStartedNodeNodeConfig{keyDir: t.TempDir()}, zap.NewNop().Sugar(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, node.Stop()) })

	gotRemote := make(chan peer.ID, 1)
	node.SetStreamHandler("/lango/contextHelpersRoundTripAgentAndChildSession1/1.0.0", func(stream network.Stream) {
		defer stream.Close()
		gotRemote <- stream.Conn().RemotePeer()
		_, _ = stream.Write([]byte("contextHelpersRoundTripAgentAndChildSession1-ok\n"))
	})

	client, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Connect(ctx, peer.AddrInfo{ID: node.PeerID(), Addrs: node.Multiaddrs()})
	require.NoError(t, err)

	stream, err := client.NewStream(ctx, node.PeerID(), "/lango/contextHelpersRoundTripAgentAndChildSession1/1.0.0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = stream.Close() })

	line, err := bufio.NewReader(stream).ReadString('\n')
	require.NoError(t, err)
	assert.Equal(t, "contextHelpersRoundTripAgentAndChildSession1-ok\n", line)
	assert.Equal(t, client.ID(), <-gotRemote)
	assert.Equal(t, []peer.ID{node.PeerID()}, client.Network().Peers())
}

func TestMDNSNotifeeIgnoresSelfAndConnectsDiscoveredPeer(t *testing.T) {
	t.Parallel()

	owner := &nodeAccessorsOnNotStartedNodeMDNSHost{id: peer.ID("owner-peer")}
	notifee := &mdnsNotifee{
		host:   owner,
		ctx:    context.Background(),
		logger: zap.NewNop().Sugar(),
	}

	notifee.HandlePeerFound(peer.AddrInfo{ID: owner.ID(), Addrs: nil})
	assert.Zero(t, owner.connectCalls)

	discovered := peer.AddrInfo{ID: peer.ID("discovered-peer")}
	notifee.HandlePeerFound(discovered)
	assert.Equal(t, 1, owner.connectCalls)
	assert.Equal(t, discovered, owner.connected[0])
}

func TestLoadOrGenerateKeyRejectsUnreadableLegacyPath(t *testing.T) {
	t.Parallel()

	keyDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(keyDir, nodeKeyFile), 0o700))

	key, err := loadOrGenerateKey(keyDir, nil, zap.NewNop().Sugar())
	require.Error(t, err)
	assert.Nil(t, key)
	assert.ErrorContains(t, err, "read node key")
}

func TestLoadOrGenerateKeyMigratesLegacyFileToSecretsStore(t *testing.T) {
	t.Parallel()

	keyDir := t.TempDir()
	legacyKey, _, err := crypto.GenerateEd25519Key(nil)
	require.NoError(t, err)
	raw, err := crypto.MarshalPrivateKey(legacyKey)
	require.NoError(t, err)
	keyPath := filepath.Join(keyDir, nodeKeyFile)
	require.NoError(t, os.WriteFile(keyPath, raw, 0o600))

	secrets := newNodeAccessorsOnNotStartedNodeSecretsStore(t)
	loaded, err := loadOrGenerateKey(keyDir, secrets, zap.NewNop().Sugar())
	require.NoError(t, err)

	loadedRaw, err := crypto.MarshalPrivateKey(loaded)
	require.NoError(t, err)
	assert.Equal(t, raw, loadedRaw)
	assert.NoFileExists(t, keyPath)

	stored, err := secrets.Get(context.Background(), nodeKeySecret)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(raw, stored))

	loadedAgain, err := loadOrGenerateKey(keyDir, secrets, zap.NewNop().Sugar())
	require.NoError(t, err)
	loadedAgainRaw, err := crypto.MarshalPrivateKey(loadedAgain)
	require.NoError(t, err)
	assert.Equal(t, raw, loadedAgainRaw)
}

type nodeAccessorsOnNotStartedNodeNodeConfig struct {
	keyDir string
}

type nodeAccessorsOnNotStartedNodeMDNSHost struct {
	host.Host
	id           peer.ID
	connectCalls int
	connected    []peer.AddrInfo
}

func (h *nodeAccessorsOnNotStartedNodeMDNSHost) ID() peer.ID {
	return h.id
}

func (h *nodeAccessorsOnNotStartedNodeMDNSHost) Connect(_ context.Context, pi peer.AddrInfo) error {
	h.connectCalls++
	h.connected = append(h.connected, pi)
	return nil
}

type nodeAccessorsOnNotStartedNodeNetworkHost struct {
	host.Host
	network network.Network
}

func (h *nodeAccessorsOnNotStartedNodeNetworkHost) Network() network.Network {
	return h.network
}

type nodeAccessorsOnNotStartedNodeNetwork struct {
	network.Network
	conns []network.Conn
}

func (n *nodeAccessorsOnNotStartedNodeNetwork) Conns() []network.Conn {
	return n.conns
}

type nodeAccessorsOnNotStartedNodeConn struct {
	network.Conn
	remote peer.ID
}

func (c nodeAccessorsOnNotStartedNodeConn) RemotePeer() peer.ID {
	return c.remote
}

type nodeAccessorsOnNotStartedNodeCloseHost struct {
	host.Host
	closeErr error
	closed   bool
}

func (h *nodeAccessorsOnNotStartedNodeCloseHost) Close() error {
	h.closed = true
	return h.closeErr
}

func (cfg nodeAccessorsOnNotStartedNodeNodeConfig) GetKeyDir() string {
	return cfg.keyDir
}

func (nodeAccessorsOnNotStartedNodeNodeConfig) GetMaxPeers() int {
	return 10
}

func (nodeAccessorsOnNotStartedNodeNodeConfig) GetListenAddrs() []string {
	return []string{"/ip4/127.0.0.1/tcp/0"}
}

func (nodeAccessorsOnNotStartedNodeNodeConfig) GetEnableRelay() bool {
	return false
}

func (nodeAccessorsOnNotStartedNodeNodeConfig) GetBootstrapPeers() []string {
	return nil
}

func (nodeAccessorsOnNotStartedNodeNodeConfig) GetEnableMDNS() bool {
	return false
}

type nodeAccessorsOnNotStartedNodeCryptoProvider struct{}

func (nodeAccessorsOnNotStartedNodeCryptoProvider) Sign(_ context.Context, _ string, payload []byte) ([]byte, error) {
	return append([]byte("sig:"), payload...), nil
}

func (nodeAccessorsOnNotStartedNodeCryptoProvider) Encrypt(_ context.Context, _ string, plaintext []byte) ([]byte, error) {
	return append([]byte("enc:"), plaintext...), nil
}

func (nodeAccessorsOnNotStartedNodeCryptoProvider) Decrypt(_ context.Context, _ string, ciphertext []byte) ([]byte, error) {
	return bytes.TrimPrefix(ciphertext, []byte("enc:")), nil
}

func newNodeAccessorsOnNotStartedNodeSecretsStore(t *testing.T) *security.SecretsStore {
	t.Helper()

	client := enttest.Open(t, "sqlite3", "file:contextHelpersRoundTripAgentAndChildSession1?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	registry := security.NewKeyRegistry(client)
	_, err := registry.RegisterKey(
		context.Background(),
		"contextHelpersRoundTripAgentAndChildSession1-default",
		"local",
		security.KeyTypeEncryption,
	)
	require.NoError(t, err)

	return security.NewSecretsStore(client, registry, nodeAccessorsOnNotStartedNodeCryptoProvider{})
}
