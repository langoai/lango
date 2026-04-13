package security

import (
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
	"golang.org/x/crypto/hkdf"
)

const domainPQSigningKey = "lango-pq-signing-mldsa65"

// PQSeedSize is the seed size for ML-DSA-65 key derivation (32 bytes).
const PQSeedSize = mldsa65.SeedSize

// DerivePQSigningSeed derives the 32-byte ML-DSA-65 seed from the Master Key
// using HKDF-SHA256. The caller should pass this to DerivePQKeyFromSeed to
// get the full keypair. Separated for bootstrap: store the seed, derive lazily.
func DerivePQSigningSeed(mk []byte, generation uint32) []byte {
	info := []byte(domainPQSigningKey)
	if generation > 0 {
		info = append(info, []byte(fmt.Sprintf(":%d", generation))...)
	}
	h := hkdf.New(sha256.New, mk, nil, info)
	seed := make([]byte, PQSeedSize)
	_, _ = io.ReadFull(h, seed)
	return seed
}

// DerivePQSigningKey derives an ML-DSA-65 keypair from the Master Key using
// HKDF-SHA256. The generation parameter allows key rotation from the same MK.
//
// Domain separation: the info label "lango-pq-signing-mldsa65[:generation]"
// is independent from the Ed25519 identity key domain "lango-identity-ed25519".
func DerivePQSigningKey(mk []byte, generation uint32) (*mldsa65.PublicKey, *mldsa65.PrivateKey) {
	seed := DerivePQSigningSeed(mk, generation)
	pk, sk := DerivePQKeyFromSeed(seed)
	ZeroBytes(seed)
	return pk, sk
}

// DerivePQKeyFromSeed creates an ML-DSA-65 keypair from a 32-byte seed.
func DerivePQKeyFromSeed(seed []byte) (*mldsa65.PublicKey, *mldsa65.PrivateKey) {
	var s [PQSeedSize]byte
	copy(s[:], seed)
	pk, sk := mldsa65.NewKeyFromSeed(&s)
	ZeroBytes(s[:])
	return pk, sk
}
