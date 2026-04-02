package escrow

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/langoai/lango/internal/types"
)

// ErrInvalidDID indicates a malformed or unresolvable DID string.
var ErrInvalidDID = errors.New("invalid DID")

// ResolveAddress converts a DID string "did:lango:<hex-compressed-pubkey>" to
// an Ethereum common.Address. It decodes the hex suffix as a compressed
// secp256k1 public key, decompresses it, and derives the Ethereum address.
func ResolveAddress(did string) (common.Address, error) {
	if !strings.HasPrefix(did, types.DIDPrefix) {
		return common.Address{}, fmt.Errorf("missing prefix %q: %w", types.DIDPrefix, ErrInvalidDID)
	}

	hexKey := strings.TrimPrefix(did, types.DIDPrefix)
	if hexKey == "" {
		return common.Address{}, fmt.Errorf("empty public key in DID: %w", ErrInvalidDID)
	}

	compressed, err := hex.DecodeString(hexKey)
	if err != nil {
		return common.Address{}, fmt.Errorf("decode hex %q: %w", hexKey, ErrInvalidDID)
	}

	pub, err := crypto.DecompressPubkey(compressed)
	if err != nil {
		return common.Address{}, fmt.Errorf("decompress pubkey: %w", ErrInvalidDID)
	}

	return crypto.PubkeyToAddress(*pub), nil
}
