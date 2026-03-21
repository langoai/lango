package identity

import (
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// VerifyMessageSignature verifies a secp256k1-keccak256 signature against a DID.
func VerifyMessageSignature(didStr string, message, signature []byte) error {
	did, err := ParseDID(didStr)
	if err != nil {
		return fmt.Errorf("parse signer DID: %w", err)
	}
	if len(signature) != 65 {
		return fmt.Errorf("invalid signature length %d", len(signature))
	}

	hash := ethcrypto.Keccak256(message)
	pub, err := ethcrypto.SigToPub(hash, signature)
	if err != nil {
		return fmt.Errorf("recover signature public key: %w", err)
	}
	recovered := ethcrypto.CompressPubkey(pub)
	if string(recovered) != string(did.PublicKey) {
		return fmt.Errorf("signature public key does not match signer DID")
	}
	return nil
}
