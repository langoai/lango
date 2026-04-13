package security

import (
	"bytes"
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Secp256k1Keccak256Scheme describes the secp256k1+keccak256 signature algorithm.
var Secp256k1Keccak256Scheme = &SignatureScheme{
	ID:            AlgorithmSecp256k1Keccak256,
	Verify:        VerifySecp256k1Keccak256,
	SignatureSize: 65,
	PublicKeySize: 33,
}

// VerifySecp256k1Keccak256 verifies a secp256k1+keccak256 ECDSA signature.
// It hashes the message with Keccak256, recovers the public key from the
// 65-byte signature (R+S+V), and compares with the claimed compressed key.
func VerifySecp256k1Keccak256(publicKey, message, signature []byte) error {
	if len(signature) != 65 {
		return fmt.Errorf("secp256k1: invalid signature length %d (expected 65)", len(signature))
	}
	if len(publicKey) != 33 {
		return fmt.Errorf("secp256k1: invalid public key length %d (expected 33)", len(publicKey))
	}

	hash := ethcrypto.Keccak256(message)
	recoveredPub, err := ethcrypto.SigToPub(hash, signature)
	if err != nil {
		return fmt.Errorf("secp256k1: recover public key: %w", err)
	}
	recoveredCompressed := ethcrypto.CompressPubkey(recoveredPub)
	if !bytes.Equal(recoveredCompressed, publicKey) {
		return fmt.Errorf("secp256k1: public key mismatch")
	}
	return nil
}
