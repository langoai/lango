package p2p

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNodeStart_CanceledParentReturnsBeforeStartingNetwork(t *testing.T) {
	t.Parallel()

	parent, cancel := context.WithCancel(context.Background())
	cancel()

	node, err := NewNode(nodeStartTestConfig{keyDir: t.TempDir()}, zap.NewNop().Sugar(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = node.Stop() })

	err = node.Start(parent, nil)
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, node.dht)
}

type nodeStartTestConfig struct {
	keyDir string
}

func (cfg nodeStartTestConfig) GetKeyDir() string {
	return cfg.keyDir
}

func (nodeStartTestConfig) GetMaxPeers() int {
	return 10
}

func (nodeStartTestConfig) GetListenAddrs() []string {
	return []string{"/ip4/127.0.0.1/tcp/0"}
}

func (nodeStartTestConfig) GetEnableRelay() bool {
	return false
}

func (nodeStartTestConfig) GetBootstrapPeers() []string {
	return nil
}

func (nodeStartTestConfig) GetEnableMDNS() bool {
	return false
}
