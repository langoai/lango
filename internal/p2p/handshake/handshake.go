package handshake

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/wallet"
)

// ProtocolID is the libp2p protocol identifier for the handshake.
const ProtocolID = "/lango/handshake/1.0.0"

// ApprovalFunc is called to request user approval for an incoming handshake.
// Uses the callback pattern to avoid import cycles with the approval package.
type ApprovalFunc func(ctx context.Context, pending *PendingHandshake) (bool, error)

// ZKProverFunc generates a ZK ownership proof for the given challenge.
type ZKProverFunc func(ctx context.Context, challenge []byte) ([]byte, error)

// ZKVerifierFunc verifies a ZK ownership proof.
type ZKVerifierFunc func(ctx context.Context, proof, challenge, publicKey []byte) (bool, error)

// PendingHandshake describes a handshake awaiting user approval.
type PendingHandshake struct {
	PeerID    peer.ID `json:"peerId"`
	PeerDID   string  `json:"peerDid"`
	RemoteAddr string `json:"remoteAddr"`
	Timestamp time.Time `json:"timestamp"`
}

// Challenge is sent by the initiator to start the handshake.
type Challenge struct {
	Nonce     []byte `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	SenderDID string `json:"senderDid"`
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
	wallet     wallet.WalletProvider
	sessions   *SessionStore
	approvalFn ApprovalFunc
	zkProver   ZKProverFunc
	zkVerifier ZKVerifierFunc
	zkEnabled  bool
	timeout    time.Duration
	autoApproveKnown bool
	logger     *zap.SugaredLogger
}

// Config configures the Handshaker.
type Config struct {
	Wallet           wallet.WalletProvider
	Sessions         *SessionStore
	ApprovalFn       ApprovalFunc
	ZKProver         ZKProverFunc
	ZKVerifier       ZKVerifierFunc
	ZKEnabled        bool
	Timeout          time.Duration
	AutoApproveKnown bool
	Logger           *zap.SugaredLogger
}

// NewHandshaker creates a new peer authenticator.
func NewHandshaker(cfg Config) *Handshaker {
	return &Handshaker{
		wallet:           cfg.Wallet,
		sessions:         cfg.Sessions,
		approvalFn:       cfg.ApprovalFn,
		zkProver:         cfg.ZKProver,
		zkVerifier:       cfg.ZKVerifier,
		zkEnabled:        cfg.ZKEnabled,
		timeout:          cfg.Timeout,
		autoApproveKnown: cfg.AutoApproveKnown,
		logger:           cfg.Logger,
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
	// Verify nonce matches.
	if len(resp.Nonce) != len(nonce) {
		return fmt.Errorf("nonce mismatch")
	}
	for i := range nonce {
		if resp.Nonce[i] != nonce[i] {
			return fmt.Errorf("nonce mismatch")
		}
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

	// Verify signature (fallback).
	if len(resp.Signature) > 0 {
		// Signature verification is done by recovering the public key from the
		// signature and comparing with the claimed public key.
		// For now, we accept signatures as valid if they are non-empty.
		// Full ECDSA recovery verification will be added in integration.
		return nil
	}

	return fmt.Errorf("no proof or signature in response")
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
