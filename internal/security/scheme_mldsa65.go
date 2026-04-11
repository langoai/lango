package security

import (
	"fmt"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
)

// MLDSA65Scheme describes the ML-DSA-65 (FIPS 204) post-quantum signature algorithm.
var MLDSA65Scheme = &SignatureScheme{
	ID:            AlgorithmMLDSA65,
	Verify:        VerifyMLDSA65,
	SignatureSize: mldsa65.SignatureSize, // 3309
	PublicKeySize: mldsa65.PublicKeySize, // 1952
}

// VerifyMLDSA65 verifies an ML-DSA-65 signature against a public key and message.
func VerifyMLDSA65(publicKey, message, signature []byte) error {
	if len(publicKey) != mldsa65.PublicKeySize {
		return fmt.Errorf("ml-dsa-65: invalid public key length %d (expected %d)", len(publicKey), mldsa65.PublicKeySize)
	}
	if len(signature) != mldsa65.SignatureSize {
		return fmt.Errorf("ml-dsa-65: invalid signature length %d (expected %d)", len(signature), mldsa65.SignatureSize)
	}

	var pk mldsa65.PublicKey
	if err := pk.UnmarshalBinary(publicKey); err != nil {
		return fmt.Errorf("ml-dsa-65: unmarshal public key: %w", err)
	}

	if !mldsa65.Verify(&pk, message, nil, signature) {
		return fmt.Errorf("ml-dsa-65: signature verification failed")
	}
	return nil
}

// SignMLDSA65 signs a message with an ML-DSA-65 private key.
// Uses deterministic signing (randomized=false) for reproducibility.
func SignMLDSA65(sk *mldsa65.PrivateKey, message []byte) ([]byte, error) {
	sig := make([]byte, mldsa65.SignatureSize)
	if err := mldsa65.SignTo(sk, message, nil, false, sig); err != nil {
		return nil, fmt.Errorf("ml-dsa-65: sign: %w", err)
	}
	return sig, nil
}
