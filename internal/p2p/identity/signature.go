package identity

import (
	"bytes"
	"fmt"
	"strings"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/langoai/lango/internal/types"
)

// VerifyMessageSignature verifies a secp256k1-keccak256 signature against a v1 DID.
// For v2 DIDs, use the verifier map pattern (BundleResolver → pubkey → verify).
func VerifyMessageSignature(didStr string, message, signature []byte) error {
	if strings.HasPrefix(didStr, types.DIDv2Prefix) {
		return fmt.Errorf("VerifyMessageSignature does not support DID v2; use BundleResolver + verifier map")
	}
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
	if !bytes.Equal(recovered, did.PublicKey) {
		return fmt.Errorf("signature public key does not match signer DID")
	}
	return nil
}
