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

type streamHandlersCloseStreamsOnEarlyFailureErrorSigner struct {
	publicKeyErr error
	didErr       error
	signErr      error
	publicKey    []byte
	did          string
	algorithm    string
	signCalls    int
}

func (s *streamHandlersCloseStreamsOnEarlyFailureErrorSigner) SignMessage(context.Context, []byte) ([]byte, error) {
	s.signCalls++
	if s.signErr != nil {
		return nil, s.signErr
	}
	return []byte("runCockpitBootstrapErrorStopsBeforeStartup6-signature"), nil
}

func (s *streamHandlersCloseStreamsOnEarlyFailureErrorSigner) PublicKey(context.Context) ([]byte, error) {
	if s.publicKeyErr != nil {
		return nil, s.publicKeyErr
	}
	if s.publicKey != nil {
		return s.publicKey, nil
	}
	return []byte("runCockpitBootstrapErrorStopsBeforeStartup6-public-key"), nil
}

func (s *streamHandlersCloseStreamsOnEarlyFailureErrorSigner) Algorithm() string {
	if s.algorithm != "" {
		return s.algorithm
	}
	return security.AlgorithmSecp256k1Keccak256
}

func (s *streamHandlersCloseStreamsOnEarlyFailureErrorSigner) DID(context.Context) (string, error) {
	if s.didErr != nil {
		return "", s.didErr
	}
	if s.did != "" {
		return s.did, nil
	}
	return "did:lango:runCockpitBootstrapErrorStopsBeforeStartup6-local", nil
}

func streamHandlersCloseStreamsOnEarlyFailureSessions(t *testing.T) *SessionStore {
	t.Helper()
	sessions, err := NewSessionStore(time.Hour)
	require.NoError(t, err)
	return sessions
}

func streamHandlersCloseStreamsOnEarlyFailureStreamWithChallenge(t *testing.T, challenge Challenge) *signerBundleAndAliasRegistrationHandshakeStream {
	t.Helper()

	var input bytes.Buffer
	require.NoError(t, json.NewEncoder(&input).Encode(challenge))
	return streamHandlersCloseStreamsOnEarlyFailureStream(&input)
}

func streamHandlersCloseStreamsOnEarlyFailureStream(rw *bytes.Buffer) *signerBundleAndAliasRegistrationHandshakeStream {
	return &signerBundleAndAliasRegistrationHandshakeStream{
		rw: rw,
		conn: &signerBundleAndAliasRegistrationHandshakeConn{
			localPeer:  peer.ID("runCockpitBootstrapErrorStopsBeforeStartup6-local"),
			remotePeer: peer.ID("runCockpitBootstrapErrorStopsBeforeStartup6-remote"),
			localAddr:  ma.StringCast("/ip4/127.0.0.1/tcp/36001"),
			remoteAddr: ma.StringCast("/ip4/127.0.0.1/tcp/36002"),
		},
	}
}

func streamHandlersCloseStreamsOnEarlyFailureUnsignedChallenge(nonce []byte) Challenge {
	return Challenge{
		Nonce:     nonce,
		Timestamp: time.Now().Unix(),
		SenderDID: "did:lango:runCockpitBootstrapErrorStopsBeforeStartup6-remote",
	}
}

func TestStreamHandlersCloseStreamsOnEarlyFailure(t *testing.T) {
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
				Signer:   &streamHandlersCloseStreamsOnEarlyFailureErrorSigner{},
				Sessions: streamHandlersCloseStreamsOnEarlyFailureSessions(t),
				Timeout:  time.Second,
				Logger:   zap.NewNop().Sugar(),
			})
			stream := streamHandlersCloseStreamsOnEarlyFailureStream(bytes.NewBufferString("{not-json"))

			tt.handler(h)(stream)

			assert.True(t, stream.closed)
		})
	}
}

func TestPreferredProtocolsExactOrder(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{ProtocolIDv12, ProtocolIDv11, ProtocolID}, PreferredProtocols(true))
	assert.Equal(t, []string{ProtocolIDv11, ProtocolID}, PreferredProtocols(false))
}

func TestSelectSignerUnknownAlgorithmUsesPrimarySigner(t *testing.T) {
	t.Parallel()

	legacy := &streamHandlersCloseStreamsOnEarlyFailureErrorSigner{did: "did:lango:legacy"}
	primary := &streamHandlersCloseStreamsOnEarlyFailureErrorSigner{did: "did:lango:primary"}
	h := NewHandshaker(Config{
		Signer:       primary,
		LegacySigner: legacy,
		Logger:       zap.NewNop().Sugar(),
	})

	assert.Same(t, primary, h.selectSigner("future-algorithm"))
}

func TestHandleIncomingEarlyFailureBranches(t *testing.T) {
	t.Parallel()

	nonce := []byte("runCockpitBootstrapErrorStopsBeforeStartup6-early-failure-nonce")
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
				SenderDID: "did:lango:runCockpitBootstrapErrorStopsBeforeStartup6-stale",
			},
			wantErr: "challenge timestamp: timestamp too old",
		},
		{
			giveName: "far future timestamp rejected before approval",
			challenge: Challenge{
				Nonce:     nonce,
				Timestamp: time.Now().Add(challengeFutureGrace + time.Minute).Unix(),
				SenderDID: "did:lango:runCockpitBootstrapErrorStopsBeforeStartup6-future",
			},
			wantErr: "challenge timestamp: timestamp too far in future",
		},
		{
			giveName: "unsigned challenge rejected when required",
			config: Config{
				RequireSignedChallenge: true,
			},
			challenge: streamHandlersCloseStreamsOnEarlyFailureUnsignedChallenge(nonce),
			wantErr:   "unsigned challenge rejected",
		},
		{
			giveName: "v2 DID without bundle rejected",
			challenge: Challenge{
				Nonce:     nonce,
				Timestamp: time.Now().Unix(),
				SenderDID: "did:lango:v2:runCockpitBootstrapErrorStopsBeforeStartup6-missing-bundle",
			},
			wantErr: "v2 DID requires identity bundle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.giveName, func(t *testing.T) {
			approvalCalled := false
			cfg := tt.config
			cfg.Signer = &streamHandlersCloseStreamsOnEarlyFailureErrorSigner{}
			cfg.Sessions = streamHandlersCloseStreamsOnEarlyFailureSessions(t)
			cfg.Timeout = time.Second
			cfg.ApprovalFn = func(context.Context, *PendingHandshake) (bool, error) {
				approvalCalled = true
				return true, nil
			}
			cfg.Logger = zap.NewNop().Sugar()

			_, err := NewHandshaker(cfg).HandleIncoming(
				context.Background(),
				streamHandlersCloseStreamsOnEarlyFailureStreamWithChallenge(t, tt.challenge),
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.False(t, approvalCalled)
		})
	}
}

func TestHandleIncomingRejectsNonceReplayBeforeApproval(t *testing.T) {
	t.Parallel()

	nonce := []byte("replay-nonce-0000000000000000000")
	nonceCache := NewNonceCache(time.Minute)
	require.True(t, nonceCache.CheckAndRecord(nonce))
	approvalCalled := false

	h := NewHandshaker(Config{
		Signer:     &streamHandlersCloseStreamsOnEarlyFailureErrorSigner{},
		Sessions:   streamHandlersCloseStreamsOnEarlyFailureSessions(t),
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
		streamHandlersCloseStreamsOnEarlyFailureStreamWithChallenge(t, streamHandlersCloseStreamsOnEarlyFailureUnsignedChallenge(nonce)),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonce replay detected")
	assert.False(t, approvalCalled)
}

func TestHandleIncomingSignerFailureBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		giveName string
		signer   *streamHandlersCloseStreamsOnEarlyFailureErrorSigner
		wantErr  string
	}{
		{
			giveName: "public key failure",
			signer: &streamHandlersCloseStreamsOnEarlyFailureErrorSigner{
				publicKeyErr: errors.New("public key store unavailable"),
			},
			wantErr: "get public key: public key store unavailable",
		},
		{
			giveName: "DID failure",
			signer: &streamHandlersCloseStreamsOnEarlyFailureErrorSigner{
				didErr: errors.New("identity wallet locked"),
			},
			wantErr: "get signer DID: identity wallet locked",
		},
		{
			giveName: "signature failure",
			signer: &streamHandlersCloseStreamsOnEarlyFailureErrorSigner{
				signErr: errors.New("signer refused challenge"),
			},
			wantErr: "sign challenge: signer refused challenge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.giveName, func(t *testing.T) {
			h := NewHandshaker(Config{
				Signer:   tt.signer,
				Sessions: streamHandlersCloseStreamsOnEarlyFailureSessions(t),
				Timeout:  time.Second,
				Logger:   zap.NewNop().Sugar(),
			})

			_, err := h.HandleIncoming(
				context.Background(),
				streamHandlersCloseStreamsOnEarlyFailureStreamWithChallenge(t, streamHandlersCloseStreamsOnEarlyFailureUnsignedChallenge([]byte(tt.giveName))),
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			if tt.signer.signErr != nil {
				assert.Equal(t, 1, tt.signer.signCalls)
			}
		})
	}
}
