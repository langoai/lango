package handshake

import (
	"bytes"
	"encoding/binary"
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/langoai/lango/internal/security"
)

// SignatureVerifyFunc verifies a signature against a claimed public key.
// Used for both challenge and response signature verification.
// pubkey is the claimed public key, message is the raw message (verifier
// handles algorithm-specific hashing internally), and signature is the
// raw signature bytes.
type SignatureVerifyFunc func(pubkey, message, signature []byte) error

// VerifySecp256k1Signature is the default verifier using secp256k1+keccak256.
// It hashes the message with Keccak256, recovers the public key from the
// 65-byte ECDSA signature (R+S+V), and compares with the claimed compressed key.
func VerifySecp256k1Signature(pubkey, message, signature []byte) error {
	if len(signature) != 65 {
		return fmt.Errorf("invalid signature length: %d (expected 65)", len(signature))
	}
	hash := ethcrypto.Keccak256(message)
	recoveredPub, err := ethcrypto.SigToPub(hash, signature)
	if err != nil {
		return fmt.Errorf("recover public key from signature: %w", err)
	}
	recoveredCompressed := ethcrypto.CompressPubkey(recoveredPub)
	if !bytes.Equal(recoveredCompressed, pubkey) {
		return fmt.Errorf("signature public key mismatch")
	}
	return nil
}

// verifyChallengeSignature verifies the signature on a v1.1/v1.2 challenge.
// Dispatches by challenge.SignatureAlgorithm, defaulting to secp256k1-keccak256.
// For v1.2, the canonical payload includes KEM fields (transcript binding).
func (h *Handshaker) verifyChallengeSignature(c *Challenge) error {
	algo := c.SignatureAlgorithm
	if algo == "" {
		algo = security.AlgorithmSecp256k1Keccak256
	}
	verifier, ok := h.verifiers[algo]
	if !ok {
		return fmt.Errorf("unsupported challenge signature algorithm %q", algo)
	}
	canonical := challengeCanonicalPayload(c.Nonce, c.Timestamp, c.SenderDID, c.KEMAlgorithm, c.KEMPublicKey)
	return verifier(c.PublicKey, canonical, c.Signature)
}

// challengeCanonicalPayload constructs the canonical bytes for challenge signing:
// nonce || bigEndian(timestamp, 8) || utf8(senderDID) [|| utf8(kemAlgorithm) || kemPublicKey]
//
// KEM fields are appended only when kemPublicKey is non-empty (v1.2).
// When empty (v1.1), the payload is identical to the previous format.
func challengeCanonicalPayload(nonce []byte, timestamp int64, senderDID string, kemAlgorithm string, kemPublicKey []byte) []byte {
	buf := make([]byte, 0, len(nonce)+8+len(senderDID)+len(kemAlgorithm)+len(kemPublicKey))
	buf = append(buf, nonce...)
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(timestamp))
	buf = append(buf, ts...)
	buf = append(buf, []byte(senderDID)...)
	if len(kemPublicKey) > 0 {
		buf = append(buf, []byte(kemAlgorithm)...)
		buf = append(buf, kemPublicKey...)
	}
	return buf
}

// responseCanonicalPayload constructs the canonical bytes for response signing:
// nonce [|| kemCiphertext]
//
// KEM ciphertext is appended only when non-empty (v1.2 transcript binding).
// When empty (v1.1), the payload is the nonce only — matching current behavior.
func responseCanonicalPayload(nonce []byte, kemCiphertext []byte) []byte {
	buf := make([]byte, 0, len(nonce)+len(kemCiphertext))
	buf = append(buf, nonce...)
	if len(kemCiphertext) > 0 {
		buf = append(buf, kemCiphertext...)
	}
	return buf
}
