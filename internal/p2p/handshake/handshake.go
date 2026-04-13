package handshake

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/types"
)

// Protocol version constants for handshake negotiation.
const (
	// ProtocolID is the legacy protocol identifier (unsigned challenges).
	ProtocolID = "/lango/handshake/1.0.0"

	// ProtocolIDv11 is the signed-challenge protocol (v1.1).
	ProtocolIDv11 = "/lango/handshake/1.1.0"

	// ProtocolIDv12 is the PQ KEM handshake protocol (v1.2).
	ProtocolIDv12 = "/lango/handshake/1.2.0"
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
	Nonce              []byte                   `json:"nonce"`
	Timestamp          int64                    `json:"timestamp"`
	SenderDID          string                   `json:"senderDid"`
	PublicKey          []byte                   `json:"publicKey,omitempty"`          // v1.1: initiator's public key
	Signature          []byte                   `json:"signature,omitempty"`          // v1.1: signature over canonical payload
	SignatureAlgorithm string                   `json:"signatureAlgorithm,omitempty"` // algorithm (empty = secp256k1-keccak256)
	Bundle             *identity.IdentityBundle `json:"bundle,omitempty"`             // v2: initiator's identity bundle
	KEMPublicKey       []byte                   `json:"kemPublicKey,omitempty"`       // v1.2: ephemeral hybrid KEM public key
	KEMAlgorithm       string                   `json:"kemAlgorithm,omitempty"`       // v1.2: "X25519-MLKEM768"
}

// ChallengeResponse is the target's reply with proof of identity.
type ChallengeResponse struct {
	Nonce              []byte                   `json:"nonce"`
	Signature          []byte                   `json:"signature,omitempty"`
	ZKProof            []byte                   `json:"zkProof,omitempty"`
	DID                string                   `json:"did"`
	PublicKey          []byte                   `json:"publicKey"`
	SignatureAlgorithm string                   `json:"signatureAlgorithm,omitempty"` // algorithm (empty = secp256k1-keccak256)
	Bundle             *identity.IdentityBundle `json:"bundle,omitempty"`             // v2: responder's identity bundle
	KEMCiphertext      []byte                   `json:"kemCiphertext,omitempty"`      // v1.2: KEM ciphertext
}

// SessionAck is sent by the initiator after verifying the response.
type SessionAck struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
}

// Signer is the minimal interface for identity signing operations.
// Implementations must declare their algorithm and DID via Algorithm() and DID().
type Signer interface {
	SignMessage(ctx context.Context, message []byte) ([]byte, error)
	PublicKey(ctx context.Context) ([]byte, error)
	Algorithm() string
	DID(ctx context.Context) (string, error)
}

// Handshaker manages peer authentication using wallet signatures or ZK proofs.
type Handshaker struct {
	signer                 Signer
	legacySigner           Signer // v1 fallback for unknown/v1 peers
	sessions               *SessionStore
	approvalFn             ApprovalFunc
	zkProver               ZKProverFunc
	zkVerifier             ZKVerifierFunc
	zkEnabled              bool
	timeout                time.Duration
	autoApproveKnown       bool
	nonceCache             *NonceCache
	requireSignedChallenge bool
	verifiers              map[string]SignatureVerifyFunc
	bundleCache            identity.BundleResolver // optional: cache received bundles
	didAlias               *identity.DIDAlias      // optional: v1/v2 DID alias for session continuity
	kemEnabled             bool                    // PQ KEM key exchange enabled
	logger                 *zap.SugaredLogger
}

// Config configures the Handshaker.
type Config struct {
	Signer                 Signer
	LegacySigner           Signer                        // v1 secp256k1 fallback (optional)
	Sessions               *SessionStore
	ApprovalFn             ApprovalFunc
	ZKProver               ZKProverFunc
	ZKVerifier             ZKVerifierFunc
	ZKEnabled              bool
	Timeout                time.Duration
	AutoApproveKnown       bool
	NonceCache             *NonceCache
	RequireSignedChallenge bool
	Verifiers              map[string]SignatureVerifyFunc // nil → default with secp256k1 + ed25519
	BundleCache            identity.BundleResolver       // optional: for caching received bundles
	DIDAlias               *identity.DIDAlias            // optional: v1/v2 DID alias for session continuity
	EnablePQKEM            bool                          // enable PQ KEM key exchange (default false)
	Logger                 *zap.SugaredLogger
}

// NewHandshaker creates a new peer authenticator.
func NewHandshaker(cfg Config) *Handshaker {
	verifiers := cfg.Verifiers
	if verifiers == nil {
		verifiers = map[string]SignatureVerifyFunc{
			security.AlgorithmSecp256k1Keccak256: VerifySecp256k1Signature,
			security.AlgorithmEd25519:            security.VerifyEd25519,
		}
	}
	return &Handshaker{
		signer:                 cfg.Signer,
		legacySigner:           cfg.LegacySigner,
		sessions:               cfg.Sessions,
		approvalFn:             cfg.ApprovalFn,
		zkProver:               cfg.ZKProver,
		zkVerifier:             cfg.ZKVerifier,
		zkEnabled:              cfg.ZKEnabled,
		timeout:                cfg.Timeout,
		autoApproveKnown:       cfg.AutoApproveKnown,
		nonceCache:             cfg.NonceCache,
		requireSignedChallenge: cfg.RequireSignedChallenge,
		verifiers:              verifiers,
		bundleCache:            cfg.BundleCache,
		didAlias:               cfg.DIDAlias,
		kemEnabled:             cfg.EnablePQKEM,
		logger:                 cfg.Logger,
	}
}

// BundleAttacher is an optional interface that Signers can implement to provide
// their IdentityBundle for inclusion in handshake messages.
type BundleAttacher interface {
	Bundle() *identity.IdentityBundle
}

// signerBundle extracts the IdentityBundle from a signer, if available.
func signerBundle(s Signer) *identity.IdentityBundle {
	if ba, ok := s.(BundleAttacher); ok {
		return ba.Bundle()
	}
	return nil
}

// canonicalDID resolves a DID through the alias registry for session/reputation continuity.
func (h *Handshaker) canonicalDID(did string) string {
	if h.didAlias != nil {
		return h.didAlias.CanonicalDID(did)
	}
	return did
}

// registerAlias registers a v2↔v1 DID alias from a received bundle.
func (h *Handshaker) registerAlias(bundle *identity.IdentityBundle, didV2 string) {
	if h.didAlias != nil && bundle != nil {
		h.didAlias.RegisterFromBundle(bundle, didV2)
	}
}

// selectSigner picks the appropriate signer based on the peer's algorithm.
// Unknown or v1 peers get the legacy signer (secp256k1). v2 peers get the primary signer.
func (h *Handshaker) selectSigner(peerAlgo string) Signer {
	if peerAlgo == "" || peerAlgo == security.AlgorithmSecp256k1Keccak256 {
		if h.legacySigner != nil {
			return h.legacySigner
		}
	}
	return h.signer
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

	// Use legacy signer for unknown peers (safe default for mixed-version).
	// Legacy peers may not have Ed25519 verifiers, so sending Ed25519
	// challenges would fail. The responder's algorithm is learned after
	// the first handshake; subsequent connections upgrade via selectSigner.
	initSigner := h.selectSigner("")

	// Anchor challenge.SenderDID to the signing identity.
	initDID, err := initSigner.DID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get signer DID: %w", err)
	}

	challenge := Challenge{
		Nonce:     nonce,
		Timestamp: time.Now().Unix(),
		SenderDID: initDID,
		Bundle:    signerBundle(initSigner),
	}

	// Generate ephemeral KEM keypair for PQ key exchange (v1.2).
	var kemDecap security.KEMDecapsulator
	if h.kemEnabled {
		kemPub, decap, kemErr := security.GenerateEphemeralKEM()
		if kemErr != nil {
			h.logger.Warnw("KEM keypair generation failed, proceeding without KEM", "error", kemErr)
		} else {
			kemDecap = decap
			challenge.KEMPublicKey = kemPub
			challenge.KEMAlgorithm = security.AlgorithmX25519MLKEM768
		}
	}

	// Sign the challenge (v1.1/v1.2 protocol — transcript includes KEM fields).
	pubkey, err := initSigner.PublicKey(ctx)
	if err != nil {
		h.logger.Warnw("challenge signing skipped: get public key", "error", err)
	} else {
		challenge.PublicKey = pubkey
		challenge.SignatureAlgorithm = initSigner.Algorithm()
		payload := challengeCanonicalPayload(nonce, challenge.Timestamp, initDID, challenge.KEMAlgorithm, challenge.KEMPublicKey)
		sig, err := initSigner.SignMessage(ctx, payload)
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

	// Verify response BEFORE caching bundle (never cache unauthenticated data).
	if err := h.verifyResponse(ctx, &resp, nonce); err != nil {
		return nil, fmt.Errorf("verify response: %w", err)
	}

	// For v2 DIDs: require bundle, verify DID hash, and bind signing key.
	if strings.HasPrefix(resp.DID, types.DIDv2Prefix) {
		if resp.Bundle == nil {
			return nil, fmt.Errorf("v2 response DID requires identity bundle")
		}
		computedDID, compErr := identity.ComputeDIDv2(resp.Bundle)
		if compErr != nil || computedDID != resp.DID {
			return nil, fmt.Errorf("v2 response DID does not match bundle hash")
		}
		if !bytes.Equal(resp.PublicKey, resp.Bundle.SigningKey.PublicKey) {
			return nil, fmt.Errorf("response public key does not match bundle signing key")
		}
	}

	// Cache received bundle AFTER authentication succeeds.
	if resp.Bundle != nil {
		if cache, ok := h.bundleCache.(*identity.MemoryBundleCache); ok {
			cache.Store(resp.DID, resp.Bundle)
		}
		h.registerAlias(resp.Bundle, resp.DID)
	}

	// Derive session key from KEM exchange (v1.2).
	var sessionKey []byte
	if kemDecap != nil && len(resp.KEMCiphertext) > 0 {
		ss, kemErr := kemDecap(resp.KEMCiphertext)
		if kemErr != nil {
			return nil, fmt.Errorf("KEM decapsulate: %w", kemErr)
		}
		sessionKey, err = security.DeriveSessionKey(ss, initDID, resp.DID)
		security.ZeroBytes(ss)
		if err != nil {
			return nil, fmt.Errorf("derive session key: %w", err)
		}
	}

	// Determine ZK verification status.
	zkVerified := len(resp.ZKProof) > 0

	// Create session.
	sess, err := h.sessions.Create(h.canonicalDID(resp.DID), zkVerified)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	if sessionKey != nil {
		sess.EncryptionKey = sessionKey
		sess.KEMUsed = true
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
		"kemUsed", sess.KEMUsed,
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
		if err := h.verifyChallengeSignature(&challenge); err != nil {
			return nil, fmt.Errorf("challenge signature: %w", err)
		}
		h.logger.Debugw("challenge signature verified", "senderDID", challenge.SenderDID)
	} else if h.requireSignedChallenge {
		return nil, fmt.Errorf("unsigned challenge rejected (requireSignedChallenge=true)")
	}

	// Verify SenderDID matches the signing public key for v1 DIDs.
	// A peer signing with their own key but claiming another DID would
	// otherwise poison the bundle cache and DID alias.
	// v2 DIDs (did:lango:v2:...) are hash-based and cannot be derived from
	// the pubkey directly; they are authenticated via bundle proofs instead.
	if len(challenge.PublicKey) > 0 && challenge.SenderDID != "" &&
		!strings.HasPrefix(challenge.SenderDID, types.DIDv2Prefix) {
		expectedDID, didErr := identity.DIDFromPublicKey(challenge.PublicKey)
		if didErr == nil && expectedDID.ID != challenge.SenderDID {
			return nil, fmt.Errorf("challenge SenderDID does not match public key")
		}
	}

	// For v2 DIDs: require bundle, verify DID hash, and bind signing key.
	if strings.HasPrefix(challenge.SenderDID, types.DIDv2Prefix) {
		if challenge.Bundle == nil {
			return nil, fmt.Errorf("v2 DID requires identity bundle")
		}
		computedDID, compErr := identity.ComputeDIDv2(challenge.Bundle)
		if compErr != nil || computedDID != challenge.SenderDID {
			return nil, fmt.Errorf("v2 DID does not match bundle hash")
		}
		// Verify the handshake public key is the bundle's signing key.
		// Without this, an attacker could replay someone else's bundle
		// while signing with a different key.
		if !bytes.Equal(challenge.PublicKey, challenge.Bundle.SigningKey.PublicKey) {
			return nil, fmt.Errorf("handshake public key does not match bundle signing key")
		}
	}

	// Cache received bundle AFTER authentication succeeds (never before).
	// Bundle is cached here but alias is NOT registered yet — alias registration
	// happens after approval to prevent forged LegacyDID from bypassing auto-approve.
	if challenge.Bundle != nil {
		if cache, ok := h.bundleCache.(*identity.MemoryBundleCache); ok {
			cache.Store(challenge.SenderDID, challenge.Bundle)
		}
	}

	// Request user approval (HITL).
	remotePeer := s.Conn().RemotePeer()
	if h.approvalFn != nil {
		// Check if auto-approve is enabled for known peers.
		// Look up session under: (1) existing alias if registered, (2) raw DID,
		// (3) bundle's LegacyDID for v1→v2 upgrade transitions.
		lookupDID := challenge.SenderDID
		if h.didAlias != nil {
			if canonical := h.didAlias.CanonicalDID(challenge.SenderDID); canonical != challenge.SenderDID {
				lookupDID = canonical
			}
		}
		existing := h.sessions.Get(lookupDID)
		needsApproval := existing == nil || !h.autoApproveKnown

		if needsApproval {
			pending := &PendingHandshake{
				PeerID:     remotePeer,
				PeerDID:    lookupDID, // canonical DID for reputation lookup
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

	// Select signer based on peer's algorithm (v1/v2 downgrade).
	signer := h.selectSigner(challenge.SignatureAlgorithm)

	// Get local public key.
	pubkey, err := signer.PublicKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("get public key: %w", err)
	}

	// Get DID from signer (v1: DIDFromPublicKey, v2: BundleProvider.DID).
	didStr, err := signer.DID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get signer DID: %w", err)
	}

	// KEM encapsulation (v1.2) — derive session key if initiator sent KEM pubkey.
	var sessionKey []byte
	var kemCiphertext []byte
	if len(challenge.KEMPublicKey) > 0 && challenge.KEMAlgorithm == security.AlgorithmX25519MLKEM768 {
		ct, ss, kemErr := security.KEMEncapsulate(challenge.KEMPublicKey)
		if kemErr != nil {
			h.logger.Warnw("KEM encapsulate failed, proceeding without KEM", "error", kemErr)
		} else {
			kemCiphertext = ct
			sessionKey, err = security.DeriveSessionKey(ss, challenge.SenderDID, didStr)
			security.ZeroBytes(ss)
			if err != nil {
				h.logger.Warnw("session key derivation failed", "error", err)
				kemCiphertext = nil
				sessionKey = nil
			}
		}
	}

	// Build response.
	resp := ChallengeResponse{
		Nonce:              challenge.Nonce,
		PublicKey:          pubkey,
		DID:                didStr,
		SignatureAlgorithm: signer.Algorithm(),
		Bundle:             signerBundle(signer),
		KEMCiphertext:      kemCiphertext,
	}

	// Sign or generate ZK proof.
	// v1.2: response signature covers nonce + kemCiphertext (transcript binding).
	signPayload := responseCanonicalPayload(challenge.Nonce, kemCiphertext)
	if h.zkEnabled && h.zkProver != nil {
		proof, err := h.zkProver(ctx, challenge.Nonce)
		if err != nil {
			h.logger.Warnw("ZK proof generation failed, falling back to signature", "error", err)
			sig, err := signer.SignMessage(ctx, signPayload)
			if err != nil {
				return nil, fmt.Errorf("sign challenge: %w", err)
			}
			resp.Signature = sig
		} else {
			resp.ZKProof = proof
		}
	} else {
		sig, err := signer.SignMessage(ctx, signPayload)
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

	// Register alias AFTER approval succeeds. Registering before approval
	// would let a forged LegacyDID bypass auto-approve by inheriting an
	// existing trusted session via canonicalDID().
	if challenge.Bundle != nil {
		h.registerAlias(challenge.Bundle, challenge.SenderDID)
	}

	zkVerified := len(resp.ZKProof) > 0
	canonicalPeer := h.canonicalDID(challenge.SenderDID)
	sess := &Session{
		PeerDID:       canonicalPeer,
		Token:         ack.Token,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Unix(ack.ExpiresAt, 0),
		ZKVerified:    zkVerified,
		EncryptionKey: sessionKey,
		KEMUsed:       sessionKey != nil,
	}

	// Store the session locally (zero existing key if overwriting).
	h.sessions.mu.Lock()
	if existing, ok := h.sessions.sessions[canonicalPeer]; ok {
		security.ZeroBytes(existing.EncryptionKey)
	}
	h.sessions.sessions[canonicalPeer] = sess
	h.sessions.mu.Unlock()

	h.logger.Infow("handshake accepted",
		"remoteDID", challenge.SenderDID,
		"zkVerified", zkVerified,
		"kemUsed", sess.KEMUsed,
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

	// Verify signature using algorithm-dispatched verifier.
	if len(resp.Signature) > 0 {
		algo := resp.SignatureAlgorithm
		if algo == "" {
			algo = security.AlgorithmSecp256k1Keccak256 // backward compat
		}
		verifier, ok := h.verifiers[algo]
		if !ok {
			return fmt.Errorf("unsupported signature algorithm %q", algo)
		}
		// v1.2 transcript binding: verify against nonce + kemCiphertext.
		verifyPayload := responseCanonicalPayload(nonce, resp.KEMCiphertext)
		return verifier(resp.PublicKey, verifyPayload, resp.Signature)
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

// StreamHandlerV12 returns a libp2p stream handler for v1.2 (PQ KEM) handshakes.
func (h *Handshaker) StreamHandlerV12() network.StreamHandler {
	return func(s network.Stream) {
		defer s.Close()

		ctx := context.Background()
		_, err := h.HandleIncoming(ctx, s)
		if err != nil {
			h.logger.Warnw("incoming v1.2 handshake failed", "peer", s.Conn().RemotePeer(), "error", err)
		}
	}
}

// PreferredProtocols returns the ordered list of protocol IDs for initiator
// stream negotiation. When kemEnabled is true, v1.2 is tried first.
// Returns []string since libp2p protocol.ID is a string alias.
func PreferredProtocols(kemEnabled bool) []string {
	if kemEnabled {
		return []string{ProtocolIDv12, ProtocolIDv11, ProtocolID}
	}
	return []string{ProtocolIDv11, ProtocolID}
}
