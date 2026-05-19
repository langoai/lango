package handshake

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/security"
)

type wave36ErrorSigner struct {
	publicKeyErr error
	didErr       error
	signErr      error
	publicKey    []byte
	did          string
	algorithm    string
	signCalls    int
}

func (s *wave36ErrorSigner) SignMessage(context.Context, []byte) ([]byte, error) {
	s.signCalls++
	if s.signErr != nil {
		return nil, s.signErr
	}
	return []byte("wave36-signature"), nil
}

func (s *wave36ErrorSigner) PublicKey(context.Context) ([]byte, error) {
	if s.publicKeyErr != nil {
		return nil, s.publicKeyErr
	}
	if s.publicKey != nil {
		return s.publicKey, nil
	}
	return []byte("wave36-public-key"), nil
}

func (s *wave36ErrorSigner) Algorithm() string {
	if s.algorithm != "" {
		return s.algorithm
	}
	return security.AlgorithmSecp256k1Keccak256
}

func (s *wave36ErrorSigner) DID(context.Context) (string, error) {
	if s.didErr != nil {
		return "", s.didErr
	}
	if s.did != "" {
		return s.did, nil
	}
	return "did:lango:wave36-local", nil
}

func wave36Sessions(t *testing.T) *SessionStore {
	t.Helper()
	sessions, err := NewSessionStore(time.Hour)
	require.NoError(t, err)
	return sessions
}

func wave36StreamWithChallenge(t *testing.T, challenge Challenge) *wave10HandshakeStream {
	t.Helper()

	var input bytes.Buffer
	require.NoError(t, json.NewEncoder(&input).Encode(challenge))
	return wave36Stream(&input)
}

func wave36Stream(rw *bytes.Buffer) *wave10HandshakeStream {
	return &wave10HandshakeStream{
		rw: rw,
		conn: &wave10HandshakeConn{
			localPeer:  peer.ID("wave36-local"),
			remotePeer: peer.ID("wave36-remote"),
			localAddr:  ma.StringCast("/ip4/127.0.0.1/tcp/36001"),
			remoteAddr: ma.StringCast("/ip4/127.0.0.1/tcp/36002"),
		},
	}
}

func wave36UnsignedChallenge(nonce []byte) Challenge {
	return Challenge{
		Nonce:     nonce,
		Timestamp: time.Now().Unix(),
		SenderDID: "did:lango:wave36-remote",
	}
}

func TestWave36StreamHandlersCloseStreamsOnEarlyFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		giveName string
		handler  func(*Handshaker) network.StreamHandler
	}{
		{
			giveName: "legacy stream handler",
			handler: func(h *Handshaker) network.StreamHandler {
				return h.StreamHandler()
			},
		},
		{
			giveName: "v1.1 stream handler",
			handler: func(h *Handshaker) network.StreamHandler {
				return h.StreamHandlerV11()
			},
		},
		{
			giveName: "v1.2 stream handler",
			handler: func(h *Handshaker) network.StreamHandler {
				return h.StreamHandlerV12()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.giveName, func(t *testing.T) {
			h := NewHandshaker(Config{
				Signer:   &wave36ErrorSigner{},
				Sessions: wave36Sessions(t),
				Timeout:  time.Second,
				Logger:   zap.NewNop().Sugar(),
			})
			stream := wave36Stream(bytes.NewBufferString("{not-json"))

			tt.handler(h)(stream)

			assert.True(t, stream.closed)
		})
	}
}

func TestWave36PreferredProtocolsExactOrder(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{ProtocolIDv12, ProtocolIDv11, ProtocolID}, PreferredProtocols(true))
	assert.Equal(t, []string{ProtocolIDv11, ProtocolID}, PreferredProtocols(false))
}

func TestWave36SelectSignerUnknownAlgorithmUsesPrimarySigner(t *testing.T) {
	t.Parallel()

	legacy := &wave36ErrorSigner{did: "did:lango:legacy"}
	primary := &wave36ErrorSigner{did: "did:lango:primary"}
	h := NewHandshaker(Config{
		Signer:       primary,
		LegacySigner: legacy,
		Logger:       zap.NewNop().Sugar(),
	})

	assert.Same(t, primary, h.selectSigner("future-algorithm"))
}

func TestWave36HandleIncomingEarlyFailureBranches(t *testing.T) {
	t.Parallel()

	nonce := []byte("wave36-early-failure-nonce")
	tests := []struct {
		giveName  string
		config    Config
		challenge Challenge
		wantErr   string
	}{
		{
			giveName: "stale timestamp rejected before approval",
			challenge: Challenge{
				Nonce:     nonce,
				Timestamp: time.Now().Add(-challengeTimestampWindow - time.Minute).Unix(),
				SenderDID: "did:lango:wave36-stale",
			},
			wantErr: "challenge timestamp: timestamp too old",
		},
		{
			giveName: "far future timestamp rejected before approval",
			challenge: Challenge{
				Nonce:     nonce,
				Timestamp: time.Now().Add(challengeFutureGrace + time.Minute).Unix(),
				SenderDID: "did:lango:wave36-future",
			},
			wantErr: "challenge timestamp: timestamp too far in future",
		},
		{
			giveName: "unsigned challenge rejected when required",
			config: Config{
				RequireSignedChallenge: true,
			},
			challenge: wave36UnsignedChallenge(nonce),
			wantErr:   "unsigned challenge rejected",
		},
		{
			giveName: "v2 DID without bundle rejected",
			challenge: Challenge{
				Nonce:     nonce,
				Timestamp: time.Now().Unix(),
				SenderDID: "did:lango:v2:wave36-missing-bundle",
			},
			wantErr: "v2 DID requires identity bundle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.giveName, func(t *testing.T) {
			approvalCalled := false
			cfg := tt.config
			cfg.Signer = &wave36ErrorSigner{}
			cfg.Sessions = wave36Sessions(t)
			cfg.Timeout = time.Second
			cfg.ApprovalFn = func(context.Context, *PendingHandshake) (bool, error) {
				approvalCalled = true
				return true, nil
			}
			cfg.Logger = zap.NewNop().Sugar()

			_, err := NewHandshaker(cfg).HandleIncoming(
				context.Background(),
				wave36StreamWithChallenge(t, tt.challenge),
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.False(t, approvalCalled)
		})
	}
}

func TestWave36HandleIncomingRejectsNonceReplayBeforeApproval(t *testing.T) {
	t.Parallel()

	nonce := []byte("wave36-replay-nonce-000000000000")
	nonceCache := NewNonceCache(time.Minute)
	require.True(t, nonceCache.CheckAndRecord(nonce))
	approvalCalled := false

	h := NewHandshaker(Config{
		Signer:     &wave36ErrorSigner{},
		Sessions:   wave36Sessions(t),
		NonceCache: nonceCache,
		ApprovalFn: func(context.Context, *PendingHandshake) (bool, error) {
			approvalCalled = true
			return true, nil
		},
		Timeout: time.Second,
		Logger:  zap.NewNop().Sugar(),
	})

	_, err := h.HandleIncoming(
		context.Background(),
		wave36StreamWithChallenge(t, wave36UnsignedChallenge(nonce)),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonce replay detected")
	assert.False(t, approvalCalled)
}

func TestWave36HandleIncomingSignerFailureBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		giveName string
		signer   *wave36ErrorSigner
		wantErr  string
	}{
		{
			giveName: "public key failure",
			signer: &wave36ErrorSigner{
				publicKeyErr: errors.New("public key store unavailable"),
			},
			wantErr: "get public key: public key store unavailable",
		},
		{
			giveName: "DID failure",
			signer: &wave36ErrorSigner{
				didErr: errors.New("identity wallet locked"),
			},
			wantErr: "get signer DID: identity wallet locked",
		},
		{
			giveName: "signature failure",
			signer: &wave36ErrorSigner{
				signErr: errors.New("signer refused challenge"),
			},
			wantErr: "sign challenge: signer refused challenge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.giveName, func(t *testing.T) {
			h := NewHandshaker(Config{
				Signer:   tt.signer,
				Sessions: wave36Sessions(t),
				Timeout:  time.Second,
				Logger:   zap.NewNop().Sugar(),
			})

			_, err := h.HandleIncoming(
				context.Background(),
				wave36StreamWithChallenge(t, wave36UnsignedChallenge([]byte(tt.giveName))),
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			if tt.signer.signErr != nil {
				assert.Equal(t, 1, tt.signer.signCalls)
			}
		})
	}
}
