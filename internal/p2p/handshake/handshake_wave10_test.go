package handshake

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/p2p/identity"
)

type wave10BundleSigner struct {
	*mockSigner
	bundle *identity.IdentityBundle
	did    string
}

func (s *wave10BundleSigner) Bundle() *identity.IdentityBundle {
	return s.bundle
}

func (s *wave10BundleSigner) DID(ctx context.Context) (string, error) {
	if s.did != "" {
		return s.did, nil
	}
	return s.mockSigner.DID(ctx)
}

type wave10HandshakeStream struct {
	rw     io.ReadWriter
	conn   network.Conn
	proto  protocol.ID
	closed bool
}

func (s *wave10HandshakeStream) Read(p []byte) (int, error)  { return s.rw.Read(p) }
func (s *wave10HandshakeStream) Write(p []byte) (int, error) { return s.rw.Write(p) }
func (s *wave10HandshakeStream) Close() error {
	s.closed = true
	if c, ok := s.rw.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
func (s *wave10HandshakeStream) CloseWrite() error { return nil }
func (s *wave10HandshakeStream) CloseRead() error  { return nil }
func (s *wave10HandshakeStream) Reset() error      { return s.Close() }
func (s *wave10HandshakeStream) ResetWithError(network.StreamErrorCode) error {
	return s.Close()
}
func (s *wave10HandshakeStream) SetDeadline(t time.Time) error {
	if d, ok := s.rw.(interface{ SetDeadline(time.Time) error }); ok {
		return d.SetDeadline(t)
	}
	return nil
}
func (s *wave10HandshakeStream) SetReadDeadline(t time.Time) error {
	if d, ok := s.rw.(interface{ SetReadDeadline(time.Time) error }); ok {
		return d.SetReadDeadline(t)
	}
	return nil
}
func (s *wave10HandshakeStream) SetWriteDeadline(t time.Time) error {
	if d, ok := s.rw.(interface{ SetWriteDeadline(time.Time) error }); ok {
		return d.SetWriteDeadline(t)
	}
	return nil
}
func (s *wave10HandshakeStream) ID() string                       { return "wave10-stream" }
func (s *wave10HandshakeStream) Protocol() protocol.ID            { return s.proto }
func (s *wave10HandshakeStream) SetProtocol(id protocol.ID) error { s.proto = id; return nil }
func (s *wave10HandshakeStream) Stat() network.Stats              { return network.Stats{} }
func (s *wave10HandshakeStream) Conn() network.Conn               { return s.conn }
func (s *wave10HandshakeStream) Scope() network.StreamScope {
	return &network.NullScope{}
}

type wave10HandshakeConn struct {
	localPeer  peer.ID
	remotePeer peer.ID
	localAddr  ma.Multiaddr
	remoteAddr ma.Multiaddr
}

func (c *wave10HandshakeConn) Close() error                               { return nil }
func (c *wave10HandshakeConn) CloseWithError(network.ConnErrorCode) error { return nil }
func (c *wave10HandshakeConn) ID() string                                 { return "wave10-conn" }
func (c *wave10HandshakeConn) NewStream(context.Context) (network.Stream, error) {
	return nil, errors.New("not implemented")
}
func (c *wave10HandshakeConn) GetStreams() []network.Stream       { return nil }
func (c *wave10HandshakeConn) IsClosed() bool                     { return false }
func (c *wave10HandshakeConn) As(any) bool                        { return false }
func (c *wave10HandshakeConn) LocalPeer() peer.ID                 { return c.localPeer }
func (c *wave10HandshakeConn) RemotePeer() peer.ID                { return c.remotePeer }
func (c *wave10HandshakeConn) RemotePublicKey() crypto.PubKey     { return nil }
func (c *wave10HandshakeConn) ConnState() network.ConnectionState { return network.ConnectionState{} }
func (c *wave10HandshakeConn) LocalMultiaddr() ma.Multiaddr       { return c.localAddr }
func (c *wave10HandshakeConn) RemoteMultiaddr() ma.Multiaddr      { return c.remoteAddr }
func (c *wave10HandshakeConn) Stat() network.ConnStats            { return network.ConnStats{} }
func (c *wave10HandshakeConn) Scope() network.ConnScope           { return &network.NullScope{} }

func wave10Signer(t *testing.T, hexKey string) *mockSigner {
	t.Helper()
	key, err := ethcrypto.HexToECDSA(hexKey)
	require.NoError(t, err)
	return &mockSigner{privKeyBytes: ethcrypto.FromECDSA(key)}
}

func wave10Bundle(t *testing.T, signer *mockSigner, legacyDID string) *identity.IdentityBundle {
	t.Helper()
	pub, err := signer.PublicKey(context.Background())
	require.NoError(t, err)
	return &identity.IdentityBundle{
		Version: 1,
		SigningKey: identity.PublicKeyEntry{
			Algorithm: signer.Algorithm(),
			PublicKey: pub,
		},
		SettlementKey: identity.PublicKeyEntry{
			Algorithm: signer.Algorithm(),
			PublicKey: pub,
		},
		LegacyDID: legacyDID,
		CreatedAt: time.Unix(1_700_000_000, 0),
	}
}

func wave10HandshakePipe() (*wave10HandshakeStream, *wave10HandshakeStream) {
	a, b := net.Pipe()
	addrA := ma.StringCast("/ip4/127.0.0.1/tcp/10001")
	addrB := ma.StringCast("/ip4/127.0.0.1/tcp/10002")
	peerA := peer.ID("wave10-peer-a")
	peerB := peer.ID("wave10-peer-b")
	return &wave10HandshakeStream{
			rw:   a,
			conn: &wave10HandshakeConn{localPeer: peerA, remotePeer: peerB, localAddr: addrA, remoteAddr: addrB},
		}, &wave10HandshakeStream{
			rw:   b,
			conn: &wave10HandshakeConn{localPeer: peerB, remotePeer: peerA, localAddr: addrB, remoteAddr: addrA},
		}
}

func TestWave10SignerBundleAndAliasRegistration(t *testing.T) {
	t.Parallel()

	signer := wave10Signer(t, "1111111111111111111111111111111111111111111111111111111111111111")
	bundle := wave10Bundle(t, signer, "did:lango:legacy-wave10")
	bundleSigner := &wave10BundleSigner{mockSigner: signer, bundle: bundle}

	assert.Nil(t, signerBundle(signer))
	assert.Same(t, bundle, signerBundle(bundleSigner))

	alias := identity.NewDIDAlias()
	h := NewHandshaker(Config{DIDAlias: alias, Logger: zap.NewNop().Sugar()})
	h.registerAlias(nil, "did:lango:v2:nil")
	assert.Equal(t, "did:lango:v2:nil", h.canonicalDID("did:lango:v2:nil"))

	h.registerAlias(bundle, "did:lango:v2:registered")
	assert.Equal(t, "did:lango:legacy-wave10", h.canonicalDID("did:lango:v2:registered"))
}

func TestWave10InitiateCachesV2BundleAndRegistersAlias(t *testing.T) {
	t.Parallel()

	localSigner := wave10Signer(t, "2222222222222222222222222222222222222222222222222222222222222222")
	localBundle := wave10Bundle(t, localSigner, "did:lango:local-legacy")
	initiatorSigner := &wave10BundleSigner{mockSigner: localSigner, bundle: localBundle}

	remoteSigner := wave10Signer(t, "3333333333333333333333333333333333333333333333333333333333333333")
	remoteBundle := wave10Bundle(t, remoteSigner, "did:lango:remote-legacy")
	remoteDID, err := identity.ComputeDIDv2(remoteBundle)
	require.NoError(t, err)

	sessions, err := NewSessionStore(time.Hour)
	require.NoError(t, err)
	cache := identity.NewMemoryBundleCache()
	alias := identity.NewDIDAlias()
	h := NewHandshaker(Config{
		Signer:      initiatorSigner,
		Sessions:    sessions,
		BundleCache: cache,
		DIDAlias:    alias,
		Timeout:     time.Second,
		Logger:      zap.NewNop().Sugar(),
	})

	client, server := wave10HandshakePipe()
	defer server.Close()
	defer client.Close()

	remoteErr := make(chan error, 1)
	go func() {
		var challenge Challenge
		if err := json.NewDecoder(server).Decode(&challenge); err != nil {
			remoteErr <- err
			return
		}
		if !assert.ObjectsAreEqual(localBundle, challenge.Bundle) {
			remoteErr <- errors.New("initiator challenge did not include signer bundle")
			return
		}
		sig, err := remoteSigner.SignMessage(context.Background(), responseCanonicalPayload(challenge.Nonce, nil))
		if err != nil {
			remoteErr <- err
			return
		}
		pub, err := remoteSigner.PublicKey(context.Background())
		if err != nil {
			remoteErr <- err
			return
		}
		resp := ChallengeResponse{
			Nonce:              challenge.Nonce,
			Signature:          sig,
			DID:                remoteDID,
			PublicKey:          pub,
			SignatureAlgorithm: remoteSigner.Algorithm(),
			Bundle:             remoteBundle,
		}
		if err := json.NewEncoder(server).Encode(resp); err != nil {
			remoteErr <- err
			return
		}
		var ack SessionAck
		if err := json.NewDecoder(server).Decode(&ack); err != nil {
			remoteErr <- err
			return
		}
		if ack.Token == "" || ack.ExpiresAt == 0 {
			remoteErr <- errors.New("empty session ack")
			return
		}
		remoteErr <- nil
	}()

	sess, err := h.Initiate(context.Background(), client, "ignored-local-did")
	require.NoError(t, err)
	require.NoError(t, <-remoteErr)
	assert.Equal(t, "did:lango:remote-legacy", sess.PeerDID)
	assert.True(t, sessions.Validate("did:lango:remote-legacy", sess.Token))
	assert.Equal(t, "did:lango:remote-legacy", h.canonicalDID(remoteDID))

	cached, err := cache.ResolveBundle(remoteDID)
	require.NoError(t, err)
	assert.Equal(t, remoteBundle, cached)
}

func TestWave10HandleIncomingApprovalBranches(t *testing.T) {
	t.Parallel()

	responderSigner := wave10Signer(t, "4444444444444444444444444444444444444444444444444444444444444444")
	initiatorSigner := wave10Signer(t, "5555555555555555555555555555555555555555555555555555555555555555")
	initiatorPub, err := initiatorSigner.PublicKey(context.Background())
	require.NoError(t, err)
	initiatorDID, err := initiatorSigner.DID(context.Background())
	require.NoError(t, err)

	newIncoming := func(t *testing.T, h *Handshaker) (*Session, error) {
		t.Helper()
		nonce := []byte("wave10-handle-incoming-nonce-0001")
		ts := time.Now().Unix()
		payload := challengeCanonicalPayload(nonce, ts, initiatorDID, "", nil)
		sig, err := initiatorSigner.SignMessage(context.Background(), payload)
		require.NoError(t, err)

		challenge := Challenge{
			Nonce:              nonce,
			Timestamp:          ts,
			SenderDID:          initiatorDID,
			PublicKey:          initiatorPub,
			Signature:          sig,
			SignatureAlgorithm: initiatorSigner.Algorithm(),
		}
		var input bytes.Buffer
		require.NoError(t, json.NewEncoder(&input).Encode(challenge))
		stream := &wave10HandshakeStream{
			rw: &input,
			conn: &wave10HandshakeConn{
				localPeer:  peer.ID("wave10-local"),
				remotePeer: peer.ID("wave10-remote"),
				localAddr:  ma.StringCast("/ip4/127.0.0.1/tcp/10003"),
				remoteAddr: ma.StringCast("/ip4/127.0.0.1/tcp/10004"),
			},
		}
		return h.HandleIncoming(context.Background(), stream)
	}

	t.Run("approval denial stops before response", func(t *testing.T) {
		sessions, err := NewSessionStore(time.Hour)
		require.NoError(t, err)
		h := NewHandshaker(Config{
			Signer:   responderSigner,
			Sessions: sessions,
			ApprovalFn: func(context.Context, *PendingHandshake) (bool, error) {
				return false, nil
			},
			Timeout: time.Second,
			Logger:  zap.NewNop().Sugar(),
		})

		_, err = newIncoming(t, h)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "handshake denied by user")
	})

	t.Run("approval error is surfaced", func(t *testing.T) {
		sessions, err := NewSessionStore(time.Hour)
		require.NoError(t, err)
		h := NewHandshaker(Config{
			Signer:   responderSigner,
			Sessions: sessions,
			ApprovalFn: func(context.Context, *PendingHandshake) (bool, error) {
				return false, errors.New("approval queue closed")
			},
			Timeout: time.Second,
			Logger:  zap.NewNop().Sugar(),
		})

		_, err = newIncoming(t, h)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "approval request: approval queue closed")
	})
}

func TestWave10FullHandshakeAutoApproveKnownSkipsApproval(t *testing.T) {
	t.Parallel()

	initiatorSigner := wave10Signer(t, "6666666666666666666666666666666666666666666666666666666666666666")
	responderSigner := wave10Signer(t, "7777777777777777777777777777777777777777777777777777777777777777")
	initiatorDID, err := initiatorSigner.DID(context.Background())
	require.NoError(t, err)

	initSessions, err := NewSessionStore(time.Hour)
	require.NoError(t, err)
	respSessions, err := NewSessionStore(time.Hour)
	require.NoError(t, err)
	_, err = respSessions.Create(initiatorDID, false)
	require.NoError(t, err)

	initiator := NewHandshaker(Config{
		Signer:   initiatorSigner,
		Sessions: initSessions,
		Timeout:  time.Second,
		Logger:   zap.NewNop().Sugar(),
	})
	responder := NewHandshaker(Config{
		Signer:           responderSigner,
		Sessions:         respSessions,
		AutoApproveKnown: true,
		ApprovalFn: func(context.Context, *PendingHandshake) (bool, error) {
			return false, errors.New("approval should have been skipped")
		},
		Timeout: time.Second,
		Logger:  zap.NewNop().Sugar(),
	})

	client, server := wave10HandshakePipe()
	defer client.Close()
	defer server.Close()

	incoming := make(chan struct {
		sess *Session
		err  error
	}, 1)
	go func() {
		sess, err := responder.HandleIncoming(context.Background(), server)
		incoming <- struct {
			sess *Session
			err  error
		}{sess: sess, err: err}
	}()

	initSess, err := initiator.Initiate(context.Background(), client, "ignored")
	require.NoError(t, err)
	var got struct {
		sess *Session
		err  error
	}
	select {
	case got = <-incoming:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for incoming handshake")
	}
	require.NoError(t, got.err)
	require.NotNil(t, got.sess)
	assert.Equal(t, initiatorDID, got.sess.PeerDID)
	assert.Equal(t, got.sess.Token, initSess.Token)
}
