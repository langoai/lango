package handshake

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/wallet"
)

// Protocol version constants for handshake negotiation.
const (
	// ProtocolID is the legacy protocol identifier (unsigned challenges).
	ProtocolID = "/lango/handshake/1.0.0"

	// ProtocolIDv11 is the signed-challenge protocol (v1.1).
	ProtocolIDv11 = "/lango/handshake/1.1.0"
)

// challengeTimestampWindow is the maximum age of a challenge timestamp (5 min).
const challengeTimestampWindow = 5 * time.Minute

// challengeFutureGrace is the maximum future drift allowed for challenge timestamps.
const challengeFutureGrace = 30 * time.Second

// ApprovalFunc is called to request user approval for an incoming handshake.
// Uses the callback pattern to avoid import cycles with the approval package.
type ApprovalFunc func(ctx context.Context, pending *PendingHandshake) (bool, error)

// ZKProverFunc generates a ZK ownership proof for the given challenge.
type ZKProverFunc func(ctx context.Context, challenge []byte) ([]byte, error)

// ZKVerifierFunc verifies a ZK ownership proof.
type ZKVerifierFunc func(ctx context.Context, proof, challenge, publicKey []byte) (bool, error)

// PendingHandshake describes a handshake awaiting user approval.
type PendingHandshake struct {
	PeerID     peer.ID   `json:"peerId"`
	PeerDID    string    `json:"peerDid"`
	RemoteAddr string    `json:"remoteAddr"`
	Timestamp  time.Time `json:"timestamp"`
}

// Challenge is sent by the initiator to start the handshake.
type Challenge struct {
	Nonce     []byte `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	SenderDID string `json:"senderDid"`
	PublicKey []byte `json:"publicKey,omitempty"`  // v1.1: initiator's public key
	Signature []byte `json:"signature,omitempty"` // v1.1: ECDSA signature over canonical payload
}

// ChallengeResponse is the target's reply with proof of identity.
type ChallengeResponse struct {
	Nonce     []byte `json:"nonce"`
	Signature []byte `json:"signature,omitempty"`
	ZKProof   []byte `json:"zkProof,omitempty"`
	DID       string `json:"did"`
	PublicKey []byte `json:"publicKey"`
}

// SessionAck is sent by the initiator after verifying the response.
type SessionAck struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
}

// Handshaker manages peer authentication using wallet signatures or ZK proofs.
type Handshaker struct {
	wallet                wallet.WalletProvider
	sessions              *SessionStore
	approvalFn            ApprovalFunc
	zkProver              ZKProverFunc
	zkVerifier            ZKVerifierFunc
	zkEnabled             bool
	timeout               time.Duration
	autoApproveKnown      bool
	nonceCache            *NonceCache
	requireSignedChallenge bool
	logger                *zap.SugaredLogger
}

// Config configures the Handshaker.
type Config struct {
	Wallet                wallet.WalletProvider
	Sessions              *SessionStore
	ApprovalFn            ApprovalFunc
	ZKProver              ZKProverFunc
	ZKVerifier            ZKVerifierFunc
	ZKEnabled             bool
	Timeout               time.Duration
	AutoApproveKnown      bool
	NonceCache            *NonceCache
	RequireSignedChallenge bool
	Logger                *zap.SugaredLogger
}

// NewHandshaker creates a new peer authenticator.
func NewHandshaker(cfg Config) *Handshaker {
	return &Handshaker{
		wallet:                 cfg.Wallet,
		sessions:               cfg.Sessions,
		approvalFn:             cfg.ApprovalFn,
		zkProver:               cfg.ZKProver,
		zkVerifier:             cfg.ZKVerifier,
		zkEnabled:              cfg.ZKEnabled,
		timeout:                cfg.Timeout,
		autoApproveKnown:       cfg.AutoApproveKnown,
		nonceCache:             cfg.NonceCache,
		requireSignedChallenge: cfg.RequireSignedChallenge,
		logger:                 cfg.Logger,
	}
}

// Initiate starts a handshake with a remote peer over the given stream.
func (h *Handshaker) Initiate(ctx context.Context, s network.Stream, localDID string) (*Session, error) {
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	// Generate challenge nonce.
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	challenge := Challenge{
		Nonce:     nonce,
		Timestamp: time.Now().Unix(),
		SenderDID: localDID,
	}

	// Sign the challenge (v1.1 protocol).
	pubkey, err := h.wallet.PublicKey(ctx)
	if err != nil {
		h.logger.Warnw("challenge signing skipped: get public key", "error", err)
	} else {
		challenge.PublicKey = pubkey
		payload := challengeSignPayload(nonce, challenge.Timestamp, localDID)
		sig, err := h.wallet.SignMessage(ctx, payload)
		if err != nil {
			h.logger.Warnw("challenge signing skipped: sign", "error", err)
		} else {
			challenge.Signature = sig
		}
	}

	// Send challenge.
	enc := json.NewEncoder(s)
	if err := enc.Encode(challenge); err != nil {
		return nil, fmt.Errorf("send challenge: %w", err)
	}

	// Receive response.
	var resp ChallengeResponse
	dec := json.NewDecoder(s)
	if err := dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("receive challenge response: %w", err)
	}

	// Verify response.
	if err := h.verifyResponse(ctx, &resp, nonce); err != nil {
		return nil, fmt.Errorf("verify response: %w", err)
	}

	// Determine ZK verification status.
	zkVerified := len(resp.ZKProof) > 0

	// Create session.
	sess, err := h.sessions.Create(resp.DID, zkVerified)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	// Send session acknowledgment.
	ack := SessionAck{
		Token:     sess.Token,
		ExpiresAt: sess.ExpiresAt.Unix(),
	}
	if err := enc.Encode(ack); err != nil {
		return nil, fmt.Errorf("send session ack: %w", err)
	}

	h.logger.Infow("handshake initiated",
		"remoteDID", resp.DID,
		"zkVerified", zkVerified,
	)

	return sess, nil
}

// HandleIncoming processes an incoming handshake request.
func (h *Handshaker) HandleIncoming(ctx context.Context, s network.Stream) (*Session, error) {
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	// Receive challenge.
	var challenge Challenge
	dec := json.NewDecoder(s)
	if err := dec.Decode(&challenge); err != nil {
		return nil, fmt.Errorf("receive challenge: %w", err)
	}

	// Validate challenge timestamp (reject stale or far-future challenges).
	if err := validateChallengeTimestamp(challenge.Timestamp); err != nil {
		return nil, fmt.Errorf("challenge timestamp: %w", err)
	}

	// Check nonce replay.
	if h.nonceCache != nil {
		if !h.nonceCache.CheckAndRecord(challenge.Nonce) {
			return nil, fmt.Errorf("nonce replay detected")
		}
	}

	// Verify challenge signature (v1.1 protocol).
	if len(challenge.Signature) > 0 && len(challenge.PublicKey) > 0 {
		if err := verifyChallengeSignature(&challenge); err != nil {
			return nil, fmt.Errorf("challenge signature: %w", err)
		}
		h.logger.Debugw("challenge signature verified", "senderDID", challenge.SenderDID)
	} else if h.requireSignedChallenge {
		return nil, fmt.Errorf("unsigned challenge rejected (requireSignedChallenge=true)")
	}

	// Request user approval (HITL).
	remotePeer := s.Conn().RemotePeer()
	if h.approvalFn != nil {
		// Check if auto-approve is enabled for known peers.
		existing := h.sessions.Get(challenge.SenderDID)
		needsApproval := existing == nil || !h.autoApproveKnown

		if needsApproval {
			pending := &PendingHandshake{
				PeerID:     remotePeer,
				PeerDID:    challenge.SenderDID,
				RemoteAddr: s.Conn().RemoteMultiaddr().String(),
				Timestamp:  time.Now(),
			}
			approved, err := h.approvalFn(ctx, pending)
			if err != nil {
				return nil, fmt.Errorf("approval request: %w", err)
			}
			if !approved {
				return nil, fmt.Errorf("handshake denied by user")
			}
		}
	}

	// Get local public key.
	pubkey, err := h.wallet.PublicKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("get public key: %w", err)
	}

	// Build response.
	resp := ChallengeResponse{
		Nonce:     challenge.Nonce,
		PublicKey: pubkey,
	}

	// Generate DID from pubkey.
	resp.DID = "did:lango:" + fmt.Sprintf("%x", pubkey)

	// Sign or generate ZK proof.
	if h.zkEnabled && h.zkProver != nil {
		proof, err := h.zkProver(ctx, challenge.Nonce)
		if err != nil {
			h.logger.Warnw("ZK proof generation failed, falling back to signature", "error", err)
			// Fall back to signature mode.
			sig, err := h.wallet.SignMessage(ctx, challenge.Nonce)
			if err != nil {
				return nil, fmt.Errorf("sign challenge: %w", err)
			}
			resp.Signature = sig
		} else {
			resp.ZKProof = proof
		}
	} else {
		sig, err := h.wallet.SignMessage(ctx, challenge.Nonce)
		if err != nil {
			return nil, fmt.Errorf("sign challenge: %w", err)
		}
		resp.Signature = sig
	}

	// Send response.
	enc := json.NewEncoder(s)
	if err := enc.Encode(resp); err != nil {
		return nil, fmt.Errorf("send response: %w", err)
	}

	// Receive session acknowledgment.
	var ack SessionAck
	if err := dec.Decode(&ack); err != nil {
		return nil, fmt.Errorf("receive session ack: %w", err)
	}

	zkVerified := len(resp.ZKProof) > 0
	sess := &Session{
		PeerDID:    challenge.SenderDID,
		Token:      ack.Token,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Unix(ack.ExpiresAt, 0),
		ZKVerified: zkVerified,
	}

	// Store the session locally as well.
	h.sessions.mu.Lock()
	h.sessions.sessions[challenge.SenderDID] = sess
	h.sessions.mu.Unlock()

	h.logger.Infow("handshake accepted",
		"remoteDID", challenge.SenderDID,
		"zkVerified", zkVerified,
	)

	return sess, nil
}

// verifyResponse checks the challenge response authenticity.
func (h *Handshaker) verifyResponse(ctx context.Context, resp *ChallengeResponse, nonce []byte) error {
	// Verify nonce matches using constant-time comparison to prevent timing attacks.
	if !hmac.Equal(resp.Nonce, nonce) {
		return fmt.Errorf("nonce mismatch")
	}

	// Verify ZK proof if provided.
	if len(resp.ZKProof) > 0 && h.zkVerifier != nil {
		valid, err := h.zkVerifier(ctx, resp.ZKProof, nonce, resp.PublicKey)
		if err != nil {
			return fmt.Errorf("ZK proof verification: %w", err)
		}
		if !valid {
			return fmt.Errorf("ZK proof invalid")
		}
		return nil
	}

	// Verify ECDSA signature by recovering the public key and comparing with the
	// claimed key (secp256k1 recovery, matching wallet.SignMessage pattern).
	if len(resp.Signature) > 0 {
		// Signature must be exactly 65 bytes: R(32) + S(32) + V(1).
		if len(resp.Signature) != 65 {
			return fmt.Errorf("invalid signature length: %d (expected 65)", len(resp.Signature))
		}

		// Hash the nonce using Keccak256 (consistent with wallet.SignMessage).
		hash := ethcrypto.Keccak256(nonce)

		// Recover the public key from the signature.
		recoveredPub, err := ethcrypto.SigToPub(hash, resp.Signature)
		if err != nil {
			return fmt.Errorf("recover public key from signature: %w", err)
		}

		// Compare the recovered compressed public key with the claimed key.
		recoveredCompressed := ethcrypto.CompressPubkey(recoveredPub)
		if !bytes.Equal(recoveredCompressed, resp.PublicKey) {
			return fmt.Errorf("signature public key mismatch")
		}

		return nil
	}

	return fmt.Errorf("no proof or signature in response")
}

// StreamHandlerV11 returns a libp2p stream handler for v1.1 (signed challenge) handshakes.
// Uses the same HandleIncoming logic since it handles both signed and unsigned challenges.
func (h *Handshaker) StreamHandlerV11() network.StreamHandler {
	return func(s network.Stream) {
		defer s.Close()

		ctx := context.Background()
		_, err := h.HandleIncoming(ctx, s)
		if err != nil {
			h.logger.Warnw("incoming v1.1 handshake failed", "peer", s.Conn().RemotePeer(), "error", err)
		}
	}
}

// challengeSignPayload constructs the canonical bytes for challenge signing:
// nonce || bigEndian(timestamp, 8) || utf8(senderDID)
func challengeSignPayload(nonce []byte, timestamp int64, senderDID string) []byte {
	buf := make([]byte, 0, len(nonce)+8+len(senderDID))
	buf = append(buf, nonce...)
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(timestamp))
	buf = append(buf, ts...)
	buf = append(buf, []byte(senderDID)...)
	return ethcrypto.Keccak256(buf)
}

// verifyChallengeSignature verifies the ECDSA signature on a v1.1 challenge.
func verifyChallengeSignature(c *Challenge) error {
	if len(c.Signature) != 65 {
		return fmt.Errorf("invalid signature length: %d (expected 65)", len(c.Signature))
	}

	payload := challengeSignPayload(c.Nonce, c.Timestamp, c.SenderDID)
	recovered, err := ethcrypto.SigToPub(payload, c.Signature)
	if err != nil {
		return fmt.Errorf("recover public key: %w", err)
	}

	recoveredCompressed := ethcrypto.CompressPubkey(recovered)
	if !bytes.Equal(recoveredCompressed, c.PublicKey) {
		return fmt.Errorf("public key mismatch")
	}

	return nil
}

// validateChallengeTimestamp ensures the challenge timestamp is within the
// acceptable window: not older than challengeTimestampWindow and not more
// than challengeFutureGrace in the future.
func validateChallengeTimestamp(ts int64) error {
	if ts <= 0 || ts > math.MaxInt64/2 {
		return fmt.Errorf("invalid timestamp value: %d", ts)
	}

	now := time.Now()
	challengeTime := time.Unix(ts, 0)

	if now.Sub(challengeTime) > challengeTimestampWindow {
		return fmt.Errorf("timestamp too old: %v ago (max %v)", now.Sub(challengeTime), challengeTimestampWindow)
	}

	if challengeTime.Sub(now) > challengeFutureGrace {
		return fmt.Errorf("timestamp too far in future: %v ahead (max %v)", challengeTime.Sub(now), challengeFutureGrace)
	}

	return nil
}

// StreamHandler returns a libp2p stream handler for incoming handshakes.
func (h *Handshaker) StreamHandler() network.StreamHandler {
	return func(s network.Stream) {
		defer s.Close()

		ctx := context.Background()
		_, err := h.HandleIncoming(ctx, s)
		if err != nil {
			h.logger.Warnw("incoming handshake failed", "peer", s.Conn().RemotePeer(), "error", err)
		}
	}
}
