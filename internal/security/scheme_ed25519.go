package security

import (
	"crypto/ed25519"
	"fmt"
)

// Ed25519Scheme describes the Ed25519 signature algorithm.
// Ed25519 is registered as a framework verification algorithm in Phase 2;
// it is not wired into production identity flows until Phase 3 (DID v2).
var Ed25519Scheme = &SignatureScheme{
	ID:            AlgorithmEd25519,
	Verify:        VerifyEd25519,
	SignatureSize: ed25519.SignatureSize, // 64
	PublicKeySize: ed25519.PublicKeySize, // 32
}

// VerifyEd25519 verifies an Ed25519 signature against a public key and message.
// Ed25519 handles hashing internally (SHA-512), so the message is passed as-is.
func VerifyEd25519(publicKey, message, signature []byte) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("ed25519: invalid public key length %d (expected %d)", len(publicKey), ed25519.PublicKeySize)
	}
	if len(signature) != ed25519.SignatureSize {
		return fmt.Errorf("ed25519: invalid signature length %d (expected %d)", len(signature), ed25519.SignatureSize)
	}
	if !ed25519.Verify(publicKey, message, signature) {
		return fmt.Errorf("ed25519: signature verification failed")
	}
	return nil
}
