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

type signerBundleAndAliasRegistrationBundleSigner struct {
	*mockSigner
	bundle *identity.IdentityBundle
	did    string
}

func (s *signerBundleAndAliasRegistrationBundleSigner) Bundle() *identity.IdentityBundle {
	return s.bundle
}

func (s *signerBundleAndAliasRegistrationBundleSigner) DID(ctx context.Context) (string, error) {
	if s.did != "" {
		return s.did, nil
	}
	return s.mockSigner.DID(ctx)
}

type signerBundleAndAliasRegistrationHandshakeStream struct {
	rw     io.ReadWriter
	conn   network.Conn
	proto  protocol.ID
	closed bool
}

func (s *signerBundleAndAliasRegistrationHandshakeStream) Read(p []byte) (int, error) {
	return s.rw.Read(p)
}
func (s *signerBundleAndAliasRegistrationHandshakeStream) Write(p []byte) (int, error) {
	return s.rw.Write(p)
}
func (s *signerBundleAndAliasRegistrationHandshakeStream) Close() error {
	s.closed = true
	if c, ok := s.rw.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
func (s *signerBundleAndAliasRegistrationHandshakeStream) CloseWrite() error { return nil }
func (s *signerBundleAndAliasRegistrationHandshakeStream) CloseRead() error  { return nil }
func (s *signerBundleAndAliasRegistrationHandshakeStream) Reset() error      { return s.Close() }
func (s *signerBundleAndAliasRegistrationHandshakeStream) ResetWithError(network.StreamErrorCode) error {
	return s.Close()
}
func (s *signerBundleAndAliasRegistrationHandshakeStream) SetDeadline(t time.Time) error {
	if d, ok := s.rw.(interface{ SetDeadline(time.Time) error }); ok {
		return d.SetDeadline(t)
	}
	return nil
}
func (s *signerBundleAndAliasRegistrationHandshakeStream) SetReadDeadline(t time.Time) error {
	if d, ok := s.rw.(interface{ SetReadDeadline(time.Time) error }); ok {
		return d.SetReadDeadline(t)
	}
	return nil
}
func (s *signerBundleAndAliasRegistrationHandshakeStream) SetWriteDeadline(t time.Time) error {
	if d, ok := s.rw.(interface{ SetWriteDeadline(time.Time) error }); ok {
		return d.SetWriteDeadline(t)
	}
	return nil
}
func (s *signerBundleAndAliasRegistrationHandshakeStream) ID() string {
	return "onChainEscrowToolsRunLifecycleAndQueryViews0-stream"
}
func (s *signerBundleAndAliasRegistrationHandshakeStream) Protocol() protocol.ID { return s.proto }
func (s *signerBundleAndAliasRegistrationHandshakeStream) SetProtocol(id protocol.ID) error {
	s.proto = id
	return nil
}
func (s *signerBundleAndAliasRegistrationHandshakeStream) Stat() network.Stats {
	return network.Stats{}
}
func (s *signerBundleAndAliasRegistrationHandshakeStream) Conn() network.Conn { return s.conn }
func (s *signerBundleAndAliasRegistrationHandshakeStream) Scope() network.StreamScope {
	return &network.NullScope{}
}

type signerBundleAndAliasRegistrationHandshakeConn struct {
	localPeer  peer.ID
	remotePeer peer.ID
	localAddr  ma.Multiaddr
	remoteAddr ma.Multiaddr
}

func (c *signerBundleAndAliasRegistrationHandshakeConn) Close() error { return nil }
func (c *signerBundleAndAliasRegistrationHandshakeConn) CloseWithError(network.ConnErrorCode) error {
	return nil
}
func (c *signerBundleAndAliasRegistrationHandshakeConn) ID() string {
	return "onChainEscrowToolsRunLifecycleAndQueryViews0-conn"
}
func (c *signerBundleAndAliasRegistrationHandshakeConn) NewStream(context.Context) (network.Stream, error) {
	return nil, errors.New("not implemented")
}
func (c *signerBundleAndAliasRegistrationHandshakeConn) GetStreams() []network.Stream   { return nil }
func (c *signerBundleAndAliasRegistrationHandshakeConn) IsClosed() bool                 { return false }
func (c *signerBundleAndAliasRegistrationHandshakeConn) As(any) bool                    { return false }
func (c *signerBundleAndAliasRegistrationHandshakeConn) LocalPeer() peer.ID             { return c.localPeer }
func (c *signerBundleAndAliasRegistrationHandshakeConn) RemotePeer() peer.ID            { return c.remotePeer }
func (c *signerBundleAndAliasRegistrationHandshakeConn) RemotePublicKey() crypto.PubKey { return nil }
func (c *signerBundleAndAliasRegistrationHandshakeConn) ConnState() network.ConnectionState {
	return network.ConnectionState{}
}
func (c *signerBundleAndAliasRegistrationHandshakeConn) LocalMultiaddr() ma.Multiaddr {
	return c.localAddr
}
func (c *signerBundleAndAliasRegistrationHandshakeConn) RemoteMultiaddr() ma.Multiaddr {
	return c.remoteAddr
}
func (c *signerBundleAndAliasRegistrationHandshakeConn) Stat() network.ConnStats {
	return network.ConnStats{}
}
func (c *signerBundleAndAliasRegistrationHandshakeConn) Scope() network.ConnScope {
	return &network.NullScope{}
}

func signerBundleAndAliasRegistrationSigner(t *testing.T, hexKey string) *mockSigner {
	t.Helper()
	key, err := ethcrypto.HexToECDSA(hexKey)
	require.NoError(t, err)
	return &mockSigner{privKeyBytes: ethcrypto.FromECDSA(key)}
}

func signerBundleAndAliasRegistrationBundle(t *testing.T, signer *mockSigner, legacyDID string) *identity.IdentityBundle {
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
		CreatedAt: time.Unix(1_700_000_000, 0).UTC(),
	}
}

func signerBundleAndAliasRegistrationHandshakePipe() (*signerBundleAndAliasRegistrationHandshakeStream, *signerBundleAndAliasRegistrationHandshakeStream) {
	a, b := net.Pipe()
	addrA := ma.StringCast("/ip4/127.0.0.1/tcp/10001")
	addrB := ma.StringCast("/ip4/127.0.0.1/tcp/10002")
	peerA := peer.ID("onChainEscrowToolsRunLifecycleAndQueryViews0-peer-a")
	peerB := peer.ID("onChainEscrowToolsRunLifecycleAndQueryViews0-peer-b")
	return &signerBundleAndAliasRegistrationHandshakeStream{
			rw:   a,
			conn: &signerBundleAndAliasRegistrationHandshakeConn{localPeer: peerA, remotePeer: peerB, localAddr: addrA, remoteAddr: addrB},
		}, &signerBundleAndAliasRegistrationHandshakeStream{
			rw:   b,
			conn: &signerBundleAndAliasRegistrationHandshakeConn{localPeer: peerB, remotePeer: peerA, localAddr: addrB, remoteAddr: addrA},
		}
}

func TestSignerBundleAndAliasRegistration(t *testing.T) {
	t.Parallel()

	signer := signerBundleAndAliasRegistrationSigner(t, "1111111111111111111111111111111111111111111111111111111111111111")
	bundle := signerBundleAndAliasRegistrationBundle(t, signer, "did:lango:legacy-onChainEscrowToolsRunLifecycleAndQueryViews0")
	bundleSigner := &signerBundleAndAliasRegistrationBundleSigner{mockSigner: signer, bundle: bundle}

	assert.Nil(t, signerBundle(signer))
	assert.Same(t, bundle, signerBundle(bundleSigner))

	alias := identity.NewDIDAlias()
	h := NewHandshaker(Config{DIDAlias: alias, Logger: zap.NewNop().Sugar()})
	h.registerAlias(nil, "did:lango:v2:nil")
	assert.Equal(t, "did:lango:v2:nil", h.canonicalDID("did:lango:v2:nil"))

	h.registerAlias(bundle, "did:lango:v2:registered")
	assert.Equal(t, "did:lango:legacy-onChainEscrowToolsRunLifecycleAndQueryViews0", h.canonicalDID("did:lango:v2:registered"))
}

func TestInitiateCachesV2BundleAndRegistersAlias(t *testing.T) {
	t.Parallel()

	localSigner := signerBundleAndAliasRegistrationSigner(t, "2222222222222222222222222222222222222222222222222222222222222222")
	localBundle := signerBundleAndAliasRegistrationBundle(t, localSigner, "did:lango:local-legacy")
	initiatorSigner := &signerBundleAndAliasRegistrationBundleSigner{mockSigner: localSigner, bundle: localBundle}

	remoteSigner := signerBundleAndAliasRegistrationSigner(t, "3333333333333333333333333333333333333333333333333333333333333333")
	remoteBundle := signerBundleAndAliasRegistrationBundle(t, remoteSigner, "did:lango:remote-legacy")
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

	client, server := signerBundleAndAliasRegistrationHandshakePipe()
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

func TestHandleIncomingApprovalBranches(t *testing.T) {
	t.Parallel()

	responderSigner := signerBundleAndAliasRegistrationSigner(t, "4444444444444444444444444444444444444444444444444444444444444444")
	initiatorSigner := signerBundleAndAliasRegistrationSigner(t, "5555555555555555555555555555555555555555555555555555555555555555")
	initiatorPub, err := initiatorSigner.PublicKey(context.Background())
	require.NoError(t, err)
	initiatorDID, err := initiatorSigner.DID(context.Background())
	require.NoError(t, err)

	newIncoming := func(t *testing.T, h *Handshaker) (*Session, error) {
		t.Helper()
		nonce := []byte("onChainEscrowToolsRunLifecycleAndQueryViews0-handle-incoming-nonce-0001")
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
		stream := &signerBundleAndAliasRegistrationHandshakeStream{
			rw: &input,
			conn: &signerBundleAndAliasRegistrationHandshakeConn{
				localPeer:  peer.ID("onChainEscrowToolsRunLifecycleAndQueryViews0-local"),
				remotePeer: peer.ID("onChainEscrowToolsRunLifecycleAndQueryViews0-remote"),
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

func TestFullHandshakeAutoApproveKnownSkipsApproval(t *testing.T) {
	t.Parallel()

	initiatorSigner := signerBundleAndAliasRegistrationSigner(t, "6666666666666666666666666666666666666666666666666666666666666666")
	responderSigner := signerBundleAndAliasRegistrationSigner(t, "7777777777777777777777777777777777777777777777777777777777777777")
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

	client, server := signerBundleAndAliasRegistrationHandshakePipe()
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
