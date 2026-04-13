package security

// Algorithm identifier constants. These are the canonical source of truth
// for algorithm identifiers across the codebase.
const (
	AlgorithmSecp256k1Keccak256 = "secp256k1-keccak256"
	AlgorithmEd25519            = "ed25519"
	AlgorithmMLDSA65            = "ml-dsa-65"
)

// SignatureScheme is a canonical algorithm descriptor with verification
// function and metadata. This type does NOT act as a registry — actual
// dispatch is handled by injected verifier maps in each consumer
// (handshake, provenance). SignatureScheme provides a single source of
// truth for algorithm metadata (key/signature sizes) and verification logic.
type SignatureScheme struct {
	// ID is the algorithm identifier string.
	ID string

	// Verify checks a signature against a public key and message.
	// The implementation handles any algorithm-specific hashing internally.
	Verify func(publicKey, message, signature []byte) error

	// SignatureSize is the expected byte length of signatures (e.g., 65 for secp256k1, 64 for Ed25519).
	SignatureSize int

	// PublicKeySize is the expected byte length of public keys (e.g., 33 for compressed secp256k1, 32 for Ed25519).
	PublicKeySize int
}
