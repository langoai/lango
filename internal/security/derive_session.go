package security

import (
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

const domainSessionKey = "lango-p2p-session-v1"

// HybridSharedSecretSize is the expected shared secret size from X25519-MLKEM768
// (32 bytes ML-KEM-768 + 32 bytes X25519).
const HybridSharedSecretSize = 64

// DeriveSessionKey derives a 32-byte AES-256 session key from the hybrid KEM
// shared secret using HKDF-SHA256.
//
// Parameters:
//   - sharedSecret: 64 bytes from hybrid KEM (ML-KEM-768 SS || X25519 SS)
//   - initiatorDID: DID of the handshake initiator (from selected signer)
//   - responderDID: DID of the handshake responder (from selected signer)
//
// The info parameter provides domain separation and binds the key to the
// specific peer pair. Both sides MUST use the same DID ordering (initiator
// first, responder second) to derive identical keys.
func DeriveSessionKey(sharedSecret []byte, initiatorDID, responderDID string) ([]byte, error) {
	if len(sharedSecret) != HybridSharedSecretSize {
		return nil, fmt.Errorf("unexpected shared secret size %d (want %d)", len(sharedSecret), HybridSharedSecretSize)
	}

	info := []byte(domainSessionKey + ":" + initiatorDID + ":" + responderDID)

	h := hkdf.New(sha256.New, sharedSecret, nil, info)
	key := make([]byte, KeySize) // 32 bytes
	if _, err := io.ReadFull(h, key); err != nil {
		return nil, fmt.Errorf("derive session key: %w", err)
	}
	return key, nil
}
