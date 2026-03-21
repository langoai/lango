package provenanceproto

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushBundleRoundTrip(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server, err := libp2p.New()
	require.NoError(t, err)
	defer server.Close()

	client, err := libp2p.New()
	require.NoError(t, err)
	defer client.Close()

	require.NoError(t, client.Connect(ctx, peer.AddrInfo{
		ID:    server.ID(),
		Addrs: server.Addrs(),
	}))

	var imported []byte
	handler := NewHandler(HandlerConfig{
		Validator: func(token string) (string, bool) {
			return "did:lango:peer", token == "token"
		},
		Importer: func(_ context.Context, peerDID string, data []byte) error {
			assert.Equal(t, "did:lango:peer", peerDID)
			imported = append([]byte(nil), data...)
			return nil
		},
	})
	server.SetStreamHandler(ProtocolID, handler.StreamHandler())

	resp, err := PushBundle(ctx, client, server.ID(), "token", []byte("bundle-json"))
	require.NoError(t, err)
	assert.True(t, resp.Stored)
	assert.Equal(t, []byte("bundle-json"), imported)
}
