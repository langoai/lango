package security

import (
	"fmt"

	"github.com/cloudflare/circl/kem"
	"github.com/cloudflare/circl/kem/hybrid"
)

// AlgorithmX25519MLKEM768 is the algorithm identifier for the hybrid
// X25519 + ML-KEM-768 key encapsulation mechanism (FIPS 203, NIST Level 3).
const AlgorithmX25519MLKEM768 = "X25519-MLKEM768"

// KEMDecapsulator recovers a shared secret from a KEM ciphertext.
// Returned by GenerateEphemeralKEM as a closure capturing the private key.
// The caller MUST NOT persist this function — it holds an ephemeral
// private key that should be discarded after a single handshake.
type KEMDecapsulator func(ciphertext []byte) (sharedSecret []byte, err error)

// HybridKEMScheme returns the X25519-MLKEM768 hybrid KEM scheme from circl.
func HybridKEMScheme() kem.Scheme {
	return hybrid.X25519MLKEM768()
}

// GenerateEphemeralKEM generates an ephemeral KEM keypair and returns the
// serialized public key and a decapsulator closure. The private key is
// captured inside the closure and never leaves the security package.
//
// The caller MUST NOT persist the decapsulator — it is ephemeral per-handshake.
// After the handshake completes, the closure (and its captured private key)
// becomes unreferenced and is garbage collected.
func GenerateEphemeralKEM() (pubKeyBytes []byte, decap KEMDecapsulator, err error) {
	scheme := HybridKEMScheme()
	pk, sk, err := scheme.GenerateKeyPair()
	if err != nil {
		return nil, nil, fmt.Errorf("generate KEM keypair: %w", err)
	}
	pubBytes, err := pk.MarshalBinary()
	if err != nil {
		return nil, nil, fmt.Errorf("marshal KEM public key: %w", err)
	}

	decap = func(ciphertext []byte) ([]byte, error) {
		return scheme.Decapsulate(sk, ciphertext)
	}
	return pubBytes, decap, nil
}

// KEMEncapsulate takes a peer's serialized KEM public key and returns
// (ciphertext, sharedSecret). The shared secret is 64 bytes for X25519-MLKEM768
// (32 bytes ML-KEM-768 || 32 bytes X25519).
func KEMEncapsulate(peerPubKeyBytes []byte) (ct, ss []byte, err error) {
	scheme := HybridKEMScheme()
	pk, err := scheme.UnmarshalBinaryPublicKey(peerPubKeyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("unmarshal KEM public key: %w", err)
	}
	ct, ss, err = scheme.Encapsulate(pk)
	if err != nil {
		return nil, nil, fmt.Errorf("KEM encapsulate: %w", err)
	}
	return ct, ss, nil
}
