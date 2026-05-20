package handshake

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/security"
)

type handshakeBehaviorReadWriter struct {
	r io.Reader
	w io.Writer
}

func (rw *handshakeBehaviorReadWriter) Read(p []byte) (int, error) {
	return rw.r.Read(p)
}

func (rw *handshakeBehaviorReadWriter) Write(p []byte) (int, error) {
	return rw.w.Write(p)
}

type handshakeBehaviorFailWriter struct {
	err error
}

func (w handshakeBehaviorFailWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func handshakeBehaviorStream(input []byte, output io.Writer) *signerBundleAndAliasRegistrationHandshakeStream {
	stream := streamHandlersCloseStreamsOnEarlyFailureStream(bytes.NewBuffer(nil))
	stream.rw = &handshakeBehaviorReadWriter{
		r: bytes.NewReader(input),
		w: output,
	}
	return stream
}

func handshakeBehaviorScriptedInput(t *testing.T, challenge Challenge, ack *SessionAck) []byte {
	t.Helper()

	var input bytes.Buffer
	require.NoError(t, json.NewEncoder(&input).Encode(challenge))
	if ack != nil {
		require.NoError(t, json.NewEncoder(&input).Encode(*ack))
	}
	return input.Bytes()
}

func handshakeBehaviorSignedChallenge(
	t *testing.T,
	signer *mockSigner,
	senderDID string,
	bundle *identity.IdentityBundle,
) Challenge {
	t.Helper()

	pub, err := signer.PublicKey(context.Background())
	require.NoError(t, err)
	nonce := []byte("handshake-behavior-signed-nonce-000000000000")
	ts := time.Now().Unix()
	payload := challengeCanonicalPayload(nonce, ts, senderDID, "", nil)
	sig, err := signer.SignMessage(context.Background(), payload)
	require.NoError(t, err)
	return Challenge{
		Nonce:              nonce,
		Timestamp:          ts,
		SenderDID:          senderDID,
		PublicKey:          pub,
		Signature:          sig,
		SignatureAlgorithm: signer.Algorithm(),
		Bundle:             bundle,
	}
}

func TestInitiateContinuesWhenChallengeSigningFails(t *testing.T) {
	t.Parallel()

	localSigner := &streamHandlersCloseStreamsOnEarlyFailureErrorSigner{
		did:       "did:lango:challenge-signing-fallback-local",
		publicKey: []byte("challenge-signing-fallback-public-key"),
		signErr:   errors.New("signing service offline"),
		signCalls: 0,
	}
	remoteSigner := signerBundleAndAliasRegistrationSigner(
		t,
		"8111111111111111111111111111111111111111111111111111111111111111",
	)
	remoteDID, err := remoteSigner.DID(context.Background())
	require.NoError(t, err)
	remotePub, err := remoteSigner.PublicKey(context.Background())
	require.NoError(t, err)
	sessions := streamHandlersCloseStreamsOnEarlyFailureSessions(t)
	h := NewHandshaker(Config{
		Signer:   localSigner,
		Sessions: sessions,
		Timeout:  time.Second,
		Logger:   zap.NewNop().Sugar(),
	})
	client, server := signerBundleAndAliasRegistrationHandshakePipe()
	defer client.Close()
	defer server.Close()

	remoteErr := make(chan error, 1)
	go func() {
		var challenge Challenge
		if err := json.NewDecoder(server).Decode(&challenge); err != nil {
			remoteErr <- err
			return
		}
		if !bytes.Equal(challenge.PublicKey, localSigner.publicKey) {
			remoteErr <- errors.New("challenge public key was not sent")
			return
		}
		if len(challenge.Signature) != 0 {
			remoteErr <- errors.New("challenge signature should be omitted after signing failure")
			return
		}
		sig, err := remoteSigner.SignMessage(
			context.Background(),
			responseCanonicalPayload(challenge.Nonce, nil),
		)
		if err != nil {
			remoteErr <- err
			return
		}
		if err := json.NewEncoder(server).Encode(ChallengeResponse{
			Nonce:              challenge.Nonce,
			Signature:          sig,
			DID:                remoteDID,
			PublicKey:          remotePub,
			SignatureAlgorithm: remoteSigner.Algorithm(),
		}); err != nil {
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

	sess, err := h.Initiate(context.Background(), client, "ignored")
	require.NoError(t, err)
	require.NoError(t, <-remoteErr)
	assert.Equal(t, 1, localSigner.signCalls)
	assert.Equal(t, remoteDID, sess.PeerDID)
	assert.True(t, sessions.Validate(remoteDID, sess.Token))
}

func TestInitiateStreamFailureBranches(t *testing.T) {
	t.Parallel()

	localSigner := signerBundleAndAliasRegistrationSigner(
		t,
		"8222222222222222222222222222222222222222222222222222222222222222",
	)

	t.Run("send challenge write failure", func(t *testing.T) {
		h := NewHandshaker(Config{
			Signer:   localSigner,
			Sessions: streamHandlersCloseStreamsOnEarlyFailureSessions(t),
			Timeout:  time.Second,
			Logger:   zap.NewNop().Sugar(),
		})
		stream := handshakeBehaviorStream(
			nil,
			handshakeBehaviorFailWriter{err: errors.New("stream write closed")},
		)

		sess, err := h.Initiate(context.Background(), stream, "ignored")
		require.Error(t, err)
		assert.Nil(t, sess)
		assert.Contains(t, err.Error(), "send challenge: stream write closed")
	})

	t.Run("receive response EOF", func(t *testing.T) {
		var written bytes.Buffer
		h := NewHandshaker(Config{
			Signer:   localSigner,
			Sessions: streamHandlersCloseStreamsOnEarlyFailureSessions(t),
			Timeout:  time.Second,
			Logger:   zap.NewNop().Sugar(),
		})
		stream := handshakeBehaviorStream(nil, &written)

		sess, err := h.Initiate(context.Background(), stream, "ignored")
		require.Error(t, err)
		assert.Nil(t, sess)
		assert.Contains(t, err.Error(), "receive challenge response: EOF")
		var challenge Challenge
		require.NoError(t, json.NewDecoder(&written).Decode(&challenge))
		assert.NotEmpty(t, challenge.Nonce)
	})
}

func TestInitiateRejectsInvalidResponsesBeforeSessionOrAck(t *testing.T) {
	t.Parallel()

	localSigner := signerBundleAndAliasRegistrationSigner(
		t,
		"8333333333333333333333333333333333333333333333333333333333333333",
	)

	t.Run("response without proof or signature", func(t *testing.T) {
		h := NewHandshaker(Config{
			Signer:   localSigner,
			Sessions: streamHandlersCloseStreamsOnEarlyFailureSessions(t),
			Timeout:  time.Second,
			Logger:   zap.NewNop().Sugar(),
		})
		client, server := signerBundleAndAliasRegistrationHandshakePipe()
		defer client.Close()
		defer server.Close()
		remoteErr := make(chan error, 1)
		go func() {
			var challenge Challenge
			if err := json.NewDecoder(server).Decode(&challenge); err != nil {
				remoteErr <- err
				return
			}
			remoteErr <- json.NewEncoder(server).Encode(ChallengeResponse{
				Nonce:     challenge.Nonce,
				DID:       "did:lango:unauthenticated-response",
				PublicKey: []byte("public"),
			})
		}()

		sess, err := h.Initiate(context.Background(), client, "ignored")
		require.Error(t, err)
		assert.Nil(t, sess)
		assert.Contains(t, err.Error(), "verify response: no proof or signature")
		require.NoError(t, <-remoteErr)
		assert.Empty(t, h.sessions.ActiveSessions())
	})

	t.Run("ack write failure after authenticated response", func(t *testing.T) {
		remoteSigner := signerBundleAndAliasRegistrationSigner(
			t,
			"8444444444444444444444444444444444444444444444444444444444444444",
		)
		remoteDID, err := remoteSigner.DID(context.Background())
		require.NoError(t, err)
		remotePub, err := remoteSigner.PublicKey(context.Background())
		require.NoError(t, err)
		h := NewHandshaker(Config{
			Signer:   localSigner,
			Sessions: streamHandlersCloseStreamsOnEarlyFailureSessions(t),
			Timeout:  time.Second,
			Logger:   zap.NewNop().Sugar(),
		})
		client, server := signerBundleAndAliasRegistrationHandshakePipe()
		defer client.Close()
		remoteErr := make(chan error, 1)
		go func() {
			defer server.Close()
			var challenge Challenge
			if err := json.NewDecoder(server).Decode(&challenge); err != nil {
				remoteErr <- err
				return
			}
			sig, err := remoteSigner.SignMessage(
				context.Background(),
				responseCanonicalPayload(challenge.Nonce, nil),
			)
			if err != nil {
				remoteErr <- err
				return
			}
			remoteErr <- json.NewEncoder(server).Encode(ChallengeResponse{
				Nonce:              challenge.Nonce,
				Signature:          sig,
				DID:                remoteDID,
				PublicKey:          remotePub,
				SignatureAlgorithm: remoteSigner.Algorithm(),
			})
		}()

		sess, err := h.Initiate(context.Background(), client, "ignored")
		require.Error(t, err)
		assert.Nil(t, sess)
		assert.Contains(t, err.Error(), "send session ack")
		require.NoError(t, <-remoteErr)
		assert.NotEmpty(t, h.sessions.Get(remoteDID))
	})
}

func TestInitiateRejectsV2ResponseBundleMismatches(t *testing.T) {
	t.Parallel()

	localSigner := signerBundleAndAliasRegistrationSigner(
		t,
		"8555555555555555555555555555555555555555555555555555555555555555",
	)

	tests := []struct {
		name      string
		response  func(t *testing.T, challenge Challenge) ChallengeResponse
		wantError string
	}{
		{
			name: "DID hash mismatch",
			response: func(t *testing.T, challenge Challenge) ChallengeResponse {
				remoteSigner := signerBundleAndAliasRegistrationSigner(
					t,
					"8666666666666666666666666666666666666666666666666666666666666666",
				)
				remoteBundle := signerBundleAndAliasRegistrationBundle(
					t,
					remoteSigner,
					"did:lango:v1-hash-mismatch",
				)
				remotePub, err := remoteSigner.PublicKey(context.Background())
				require.NoError(t, err)
				sig, err := remoteSigner.SignMessage(
					context.Background(),
					responseCanonicalPayload(challenge.Nonce, nil),
				)
				require.NoError(t, err)
				return ChallengeResponse{
					Nonce:              challenge.Nonce,
					Signature:          sig,
					DID:                "did:lango:v2:not-the-bundle-hash",
					PublicKey:          remotePub,
					SignatureAlgorithm: remoteSigner.Algorithm(),
					Bundle:             remoteBundle,
				}
			},
			wantError: "v2 response DID does not match bundle hash",
		},
		{
			name: "public key does not match bundle signing key",
			response: func(t *testing.T, challenge Challenge) ChallengeResponse {
				bundleSigner := signerBundleAndAliasRegistrationSigner(
					t,
					"8777777777777777777777777777777777777777777777777777777777777777",
				)
				responseSigner := signerBundleAndAliasRegistrationSigner(
					t,
					"8888888888888888888888888888888888888888888888888888888888888888",
				)
				bundle := signerBundleAndAliasRegistrationBundle(
					t,
					bundleSigner,
					"did:lango:v1-key-mismatch",
				)
				did, err := identity.ComputeDIDv2(bundle)
				require.NoError(t, err)
				pub, err := responseSigner.PublicKey(context.Background())
				require.NoError(t, err)
				sig, err := responseSigner.SignMessage(
					context.Background(),
					responseCanonicalPayload(challenge.Nonce, nil),
				)
				require.NoError(t, err)
				return ChallengeResponse{
					Nonce:              challenge.Nonce,
					Signature:          sig,
					DID:                did,
					PublicKey:          pub,
					SignatureAlgorithm: responseSigner.Algorithm(),
					Bundle:             bundle,
				}
			},
			wantError: "response public key does not match bundle signing key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandshaker(Config{
				Signer:   localSigner,
				Sessions: streamHandlersCloseStreamsOnEarlyFailureSessions(t),
				Timeout:  time.Second,
				Logger:   zap.NewNop().Sugar(),
			})
			client, server := signerBundleAndAliasRegistrationHandshakePipe()
			defer client.Close()
			defer server.Close()
			remoteErr := make(chan error, 1)
			go func() {
				var challenge Challenge
				if err := json.NewDecoder(server).Decode(&challenge); err != nil {
					remoteErr <- err
					return
				}
				remoteErr <- json.NewEncoder(server).Encode(tt.response(t, challenge))
			}()

			sess, err := h.Initiate(context.Background(), client, "ignored")
			require.Error(t, err)
			assert.Nil(t, sess)
			assert.Contains(t, err.Error(), tt.wantError)
			require.NoError(t, <-remoteErr)
			assert.Empty(t, h.sessions.ActiveSessions())
		})
	}
}

func TestInitiateRejectsInvalidKEMCiphertext(t *testing.T) {
	t.Parallel()

	localSigner := signerBundleAndAliasRegistrationSigner(
		t,
		"8999999999999999999999999999999999999999999999999999999999999999",
	)
	remoteSigner := signerBundleAndAliasRegistrationSigner(
		t,
		"8aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	)
	remoteDID, err := remoteSigner.DID(context.Background())
	require.NoError(t, err)
	remotePub, err := remoteSigner.PublicKey(context.Background())
	require.NoError(t, err)
	h := NewHandshaker(Config{
		Signer:      localSigner,
		Sessions:    streamHandlersCloseStreamsOnEarlyFailureSessions(t),
		Timeout:     time.Second,
		EnablePQKEM: true,
		Logger:      zap.NewNop().Sugar(),
	})
	client, server := signerBundleAndAliasRegistrationHandshakePipe()
	defer client.Close()
	defer server.Close()
	remoteErr := make(chan error, 1)
	go func() {
		var challenge Challenge
		if err := json.NewDecoder(server).Decode(&challenge); err != nil {
			remoteErr <- err
			return
		}
		if len(challenge.KEMPublicKey) == 0 ||
			challenge.KEMAlgorithm != security.AlgorithmX25519MLKEM768 {
			remoteErr <- errors.New("initiator did not advertise KEM")
			return
		}
		badCiphertext := []byte("not-a-valid-kem-ciphertext")
		sig, err := remoteSigner.SignMessage(
			context.Background(),
			responseCanonicalPayload(challenge.Nonce, badCiphertext),
		)
		if err != nil {
			remoteErr <- err
			return
		}
		remoteErr <- json.NewEncoder(server).Encode(ChallengeResponse{
			Nonce:              challenge.Nonce,
			Signature:          sig,
			DID:                remoteDID,
			PublicKey:          remotePub,
			SignatureAlgorithm: remoteSigner.Algorithm(),
			KEMCiphertext:      badCiphertext,
		})
	}()

	sess, err := h.Initiate(context.Background(), client, "ignored")
	require.Error(t, err)
	assert.Nil(t, sess)
	assert.Contains(t, err.Error(), "KEM decapsulate")
	require.NoError(t, <-remoteErr)
}

func TestHandleIncomingRejectsSignedChallengeIdentityMismatches(t *testing.T) {
	t.Parallel()

	responder := signerBundleAndAliasRegistrationSigner(
		t,
		"8bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	)
	initiator := signerBundleAndAliasRegistrationSigner(
		t,
		"8ccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
	)

	tests := []struct {
		name      string
		challenge func(t *testing.T) Challenge
		wantError string
	}{
		{
			name: "v1 SenderDID does not match public key",
			challenge: func(t *testing.T) Challenge {
				other := signerBundleAndAliasRegistrationSigner(
					t,
					"8ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
				)
				otherDID, err := other.DID(context.Background())
				require.NoError(t, err)
				return handshakeBehaviorSignedChallenge(t, initiator, otherDID, nil)
			},
			wantError: "challenge SenderDID does not match public key",
		},
		{
			name: "v2 DID hash mismatch",
			challenge: func(t *testing.T) Challenge {
				bundle := signerBundleAndAliasRegistrationBundle(
					t,
					initiator,
					"did:lango:v1-incoming-hash-mismatch",
				)
				return handshakeBehaviorSignedChallenge(
					t,
					initiator,
					"did:lango:v2:not-the-incoming-bundle-hash",
					bundle,
				)
			},
			wantError: "v2 DID does not match bundle hash",
		},
		{
			name: "v2 public key does not match bundle signing key",
			challenge: func(t *testing.T) Challenge {
				bundleSigner := signerBundleAndAliasRegistrationSigner(
					t,
					"8eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
				)
				bundle := signerBundleAndAliasRegistrationBundle(
					t,
					bundleSigner,
					"did:lango:v1-incoming-key-mismatch",
				)
				did, err := identity.ComputeDIDv2(bundle)
				require.NoError(t, err)
				return handshakeBehaviorSignedChallenge(t, initiator, did, bundle)
			},
			wantError: "handshake public key does not match bundle signing key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			approvalCalled := false
			h := NewHandshaker(Config{
				Signer:   responder,
				Sessions: streamHandlersCloseStreamsOnEarlyFailureSessions(t),
				ApprovalFn: func(context.Context, *PendingHandshake) (bool, error) {
					approvalCalled = true
					return true, nil
				},
				Timeout: time.Second,
				Logger:  zap.NewNop().Sugar(),
			})

			sess, err := h.HandleIncoming(
				context.Background(),
				streamHandlersCloseStreamsOnEarlyFailureStreamWithChallenge(t, tt.challenge(t)),
			)

			require.Error(t, err)
			assert.Nil(t, sess)
			assert.Contains(t, err.Error(), tt.wantError)
			assert.False(t, approvalCalled)
		})
	}
}

func TestHandleIncomingCachesAuthenticatedBundleBeforeApprovalButDoesNotAlias(t *testing.T) {
	t.Parallel()

	responder := signerBundleAndAliasRegistrationSigner(
		t,
		"8fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	)
	initiator := signerBundleAndAliasRegistrationSigner(
		t,
		"9111111111111111111111111111111111111111111111111111111111111111",
	)
	bundle := signerBundleAndAliasRegistrationBundle(
		t,
		initiator,
		"did:lango:v1-cache-before-approval",
	)
	did, err := identity.ComputeDIDv2(bundle)
	require.NoError(t, err)
	cache := identity.NewMemoryBundleCache()
	alias := identity.NewDIDAlias()
	h := NewHandshaker(Config{
		Signer:      responder,
		Sessions:    streamHandlersCloseStreamsOnEarlyFailureSessions(t),
		BundleCache: cache,
		DIDAlias:    alias,
		ApprovalFn: func(context.Context, *PendingHandshake) (bool, error) {
			return false, nil
		},
		Timeout: time.Second,
		Logger:  zap.NewNop().Sugar(),
	})

	sess, err := h.HandleIncoming(
		context.Background(),
		streamHandlersCloseStreamsOnEarlyFailureStreamWithChallenge(
			t,
			handshakeBehaviorSignedChallenge(t, initiator, did, bundle),
		),
	)

	require.Error(t, err)
	assert.Nil(t, sess)
	assert.Contains(t, err.Error(), "handshake denied by user")
	cached, err := cache.ResolveBundle(did)
	require.NoError(t, err)
	assert.Equal(t, bundle, cached)
	assert.Equal(t, did, alias.CanonicalDID(did))
}

func TestHandleIncomingZKProofAndFallbackResponses(t *testing.T) {
	t.Parallel()

	initiator := signerBundleAndAliasRegistrationSigner(
		t,
		"9222222222222222222222222222222222222222222222222222222222222222",
	)
	initiatorDID, err := initiator.DID(context.Background())
	require.NoError(t, err)
	responder := signerBundleAndAliasRegistrationSigner(
		t,
		"9333333333333333333333333333333333333333333333333333333333333333",
	)

	t.Run("ZK proof response marks stored session verified", func(t *testing.T) {
		challenge := handshakeBehaviorSignedChallenge(t, initiator, initiatorDID, nil)
		ack := &SessionAck{Token: "zk-session-token", ExpiresAt: time.Now().Add(time.Hour).Unix()}
		var output bytes.Buffer
		sessions := streamHandlersCloseStreamsOnEarlyFailureSessions(t)
		h := NewHandshaker(Config{
			Signer:    responder,
			Sessions:  sessions,
			ZKEnabled: true,
			ZKProver: func(_ context.Context, challengeNonce []byte) ([]byte, error) {
				assert.Equal(t, challenge.Nonce, challengeNonce)
				return []byte("zk-proof"), nil
			},
			Timeout: time.Second,
			Logger:  zap.NewNop().Sugar(),
		})

		sess, err := h.HandleIncoming(
			context.Background(),
			handshakeBehaviorStream(handshakeBehaviorScriptedInput(t, challenge, ack), &output),
		)

		require.NoError(t, err)
		assert.Equal(t, initiatorDID, sess.PeerDID)
		assert.Equal(t, "zk-session-token", sess.Token)
		assert.True(t, sess.ZKVerified)
		stored := sessions.Get(initiatorDID)
		require.NotNil(t, stored)
		assert.True(t, stored.ZKVerified)
		var resp ChallengeResponse
		require.NoError(t, json.NewDecoder(&output).Decode(&resp))
		assert.Equal(t, []byte("zk-proof"), resp.ZKProof)
		assert.Empty(t, resp.Signature)
	})

	t.Run("ZK prover error falls back to signed response", func(t *testing.T) {
		challenge := handshakeBehaviorSignedChallenge(t, initiator, initiatorDID, nil)
		ack := &SessionAck{Token: "fallback-session-token", ExpiresAt: time.Now().Add(time.Hour).Unix()}
		var output bytes.Buffer
		h := NewHandshaker(Config{
			Signer:    responder,
			Sessions:  streamHandlersCloseStreamsOnEarlyFailureSessions(t),
			ZKEnabled: true,
			ZKProver: func(context.Context, []byte) ([]byte, error) {
				return nil, errors.New("prover unavailable")
			},
			Timeout: time.Second,
			Logger:  zap.NewNop().Sugar(),
		})

		sess, err := h.HandleIncoming(
			context.Background(),
			handshakeBehaviorStream(handshakeBehaviorScriptedInput(t, challenge, ack), &output),
		)

		require.NoError(t, err)
		assert.False(t, sess.ZKVerified)
		var resp ChallengeResponse
		require.NoError(t, json.NewDecoder(&output).Decode(&resp))
		assert.Empty(t, resp.ZKProof)
		assert.NotEmpty(t, resp.Signature)
		assert.NoError(t, h.verifyResponse(context.Background(), &resp, challenge.Nonce))
	})
}

func TestHandleIncomingResponseWriteAndAckReadFailures(t *testing.T) {
	t.Parallel()

	initiator := signerBundleAndAliasRegistrationSigner(
		t,
		"9444444444444444444444444444444444444444444444444444444444444444",
	)
	initiatorDID, err := initiator.DID(context.Background())
	require.NoError(t, err)
	responder := signerBundleAndAliasRegistrationSigner(
		t,
		"9555555555555555555555555555555555555555555555555555555555555555",
	)

	t.Run("send response write failure", func(t *testing.T) {
		challenge := handshakeBehaviorSignedChallenge(t, initiator, initiatorDID, nil)
		h := NewHandshaker(Config{
			Signer:   responder,
			Sessions: streamHandlersCloseStreamsOnEarlyFailureSessions(t),
			Timeout:  time.Second,
			Logger:   zap.NewNop().Sugar(),
		})

		sess, err := h.HandleIncoming(
			context.Background(),
			handshakeBehaviorStream(
				handshakeBehaviorScriptedInput(t, challenge, &SessionAck{}),
				handshakeBehaviorFailWriter{err: errors.New("response writer closed")},
			),
		)

		require.Error(t, err)
		assert.Nil(t, sess)
		assert.Contains(t, err.Error(), "send response: response writer closed")
	})

	t.Run("receive session ack EOF after response", func(t *testing.T) {
		challenge := handshakeBehaviorSignedChallenge(t, initiator, initiatorDID, nil)
		var output bytes.Buffer
		h := NewHandshaker(Config{
			Signer:   responder,
			Sessions: streamHandlersCloseStreamsOnEarlyFailureSessions(t),
			Timeout:  time.Second,
			Logger:   zap.NewNop().Sugar(),
		})

		sess, err := h.HandleIncoming(
			context.Background(),
			handshakeBehaviorStream(handshakeBehaviorScriptedInput(t, challenge, nil), &output),
		)

		require.Error(t, err)
		assert.Nil(t, sess)
		assert.Contains(t, err.Error(), "receive session ack: EOF")
		var resp ChallengeResponse
		require.NoError(t, json.NewDecoder(&output).Decode(&resp))
		assert.Equal(t, challenge.Nonce, resp.Nonce)
		assert.NotEmpty(t, resp.Signature)
	})
}
