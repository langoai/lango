package handshake

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInitiateReturnsSignerDIDErrorBeforeWritingChallenge(t *testing.T) {
	t.Parallel()

	sessions, err := NewSessionStore(time.Hour)
	require.NoError(t, err)
	h := NewHandshaker(Config{
		Signer: &streamHandlersCloseStreamsOnEarlyFailureErrorSigner{
			didErr: errors.New("wallet locked"),
		},
		Sessions: sessions,
		Timeout:  time.Second,
		Logger:   zap.NewNop().Sugar(),
	})

	var written bytes.Buffer
	_, err = h.Initiate(context.Background(), streamHandlersCloseStreamsOnEarlyFailureStream(&written), "ignored")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get signer DID: wallet locked")
	assert.Empty(t, written.Bytes())
}

func TestInitiateContinuesWhenChallengeSigningPublicKeyFails(t *testing.T) {
	t.Parallel()

	localSigner := &streamHandlersCloseStreamsOnEarlyFailureErrorSigner{
		did:          "did:lango:unsigned-challenge-local",
		publicKeyErr: errors.New("public key unavailable"),
	}
	remoteKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	remoteSigner := &mockSigner{privKeyBytes: ethcrypto.FromECDSA(remoteKey)}
	remoteDID, err := remoteSigner.DID(context.Background())
	require.NoError(t, err)
	remotePub, err := remoteSigner.PublicKey(context.Background())
	require.NoError(t, err)

	sessions, err := NewSessionStore(time.Hour)
	require.NoError(t, err)
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
		if len(challenge.PublicKey) != 0 || len(challenge.Signature) != 0 {
			remoteErr <- errors.New("challenge should be sent unsigned when public key lookup fails")
			return
		}
		if challenge.SenderDID != "did:lango:unsigned-challenge-local" {
			remoteErr <- errors.New("challenge used unexpected sender DID")
			return
		}
		sig, err := remoteSigner.SignMessage(context.Background(), responseCanonicalPayload(challenge.Nonce, nil))
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
	assert.Equal(t, remoteDID, sess.PeerDID)
	assert.True(t, sessions.Validate(remoteDID, sess.Token))
}

func TestInitiateRejectsAuthenticatedV2ResponseWithoutBundle(t *testing.T) {
	t.Parallel()

	localKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	localSigner := &mockSigner{privKeyBytes: ethcrypto.FromECDSA(localKey)}
	remoteKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	remoteSigner := &mockSigner{privKeyBytes: ethcrypto.FromECDSA(remoteKey)}
	remotePub, err := remoteSigner.PublicKey(context.Background())
	require.NoError(t, err)

	sessions, err := NewSessionStore(time.Hour)
	require.NoError(t, err)
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
		sig, err := remoteSigner.SignMessage(context.Background(), responseCanonicalPayload(challenge.Nonce, nil))
		if err != nil {
			remoteErr <- err
			return
		}
		remoteErr <- json.NewEncoder(server).Encode(ChallengeResponse{
			Nonce:              challenge.Nonce,
			Signature:          sig,
			DID:                "did:lango:v2:missing-bundle",
			PublicKey:          remotePub,
			SignatureAlgorithm: remoteSigner.Algorithm(),
		})
	}()

	sess, err := h.Initiate(context.Background(), client, "ignored")
	require.Error(t, err)
	assert.Nil(t, sess)
	assert.Contains(t, err.Error(), "v2 response DID requires identity bundle")
	require.NoError(t, <-remoteErr)
}
