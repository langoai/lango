package security

import (
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

const domainIdentityKey = "lango-identity-ed25519"

// DeriveIdentityKey derives an Ed25519 identity key from the Master Key using
// HKDF-SHA256 with a domain-separated info label. The generation parameter
// allows key rotation: generation 0 is the default, higher values produce
// different keys from the same MK.
//
// Same MK + same generation always produces the same Ed25519 key (deterministic).
// MK recovery (via mnemonic) recovers the identity key.
func DeriveIdentityKey(mk []byte, generation uint32) ed25519.PrivateKey {
	info := []byte(domainIdentityKey)
	if generation > 0 {
		info = append(info, []byte(fmt.Sprintf(":%d", generation))...)
	}
	h := hkdf.New(sha256.New, mk, nil, info)
	seed := make([]byte, ed25519.SeedSize) // 32 bytes
	_, _ = io.ReadFull(h, seed)
	return ed25519.NewKeyFromSeed(seed)
}
