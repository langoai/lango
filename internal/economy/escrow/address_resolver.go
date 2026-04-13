package escrow

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/langoai/lango/internal/types"
)

// ErrInvalidDID indicates a malformed or unresolvable DID string.
var ErrInvalidDID = errors.New("invalid DID")

// ErrBundleNotFound indicates a v2 DID's identity bundle is not available.
var ErrBundleNotFound = errors.New("identity bundle not found")

// SettlementKeyLookup resolves a v2 DID to a compressed secp256k1 settlement
// public key. This avoids importing p2p/identity directly — the lookup is
// injected at the app wiring layer.
type SettlementKeyLookup func(did string) ([]byte, error)

// AddressResolver converts a DID string to an Ethereum address.
type AddressResolver interface {
	ResolveAddress(did string) (common.Address, error)
}

// DefaultAddressResolver dispatches v1 DIDs directly (secp256k1 decompress)
// and v2 DIDs via settlement key lookup.
type DefaultAddressResolver struct {
	settlementLookup SettlementKeyLookup // nil = v1-only mode
}

// NewDefaultAddressResolver creates an address resolver.
// settlementLookup may be nil for v1-only environments.
func NewDefaultAddressResolver(settlementLookup SettlementKeyLookup) *DefaultAddressResolver {
	return &DefaultAddressResolver{settlementLookup: settlementLookup}
}

// ResolveAddress converts a DID to an Ethereum address.
func (r *DefaultAddressResolver) ResolveAddress(did string) (common.Address, error) {
	if strings.HasPrefix(did, types.DIDv2Prefix) {
		return r.resolveV2(did)
	}
	return resolveV1(did)
}

// resolveV2 looks up the settlement key (secp256k1) for a v2 DID and derives
// the Ethereum address.
func (r *DefaultAddressResolver) resolveV2(did string) (common.Address, error) {
	if r.settlementLookup == nil {
		return common.Address{}, fmt.Errorf("v2 DID %q: %w (no settlement key resolver)", did, ErrBundleNotFound)
	}
	settlementKey, err := r.settlementLookup(did)
	if err != nil {
		return common.Address{}, fmt.Errorf("v2 DID %q: %w", did, ErrBundleNotFound)
	}
	pub, err := ethcrypto.DecompressPubkey(settlementKey)
	if err != nil {
		return common.Address{}, fmt.Errorf("v2 DID %q decompress settlement key: %w", did, ErrInvalidDID)
	}
	return ethcrypto.PubkeyToAddress(*pub), nil
}

// resolveV1 converts a v1 DID "did:lango:<hex-compressed-pubkey>" to an Ethereum address.
func resolveV1(did string) (common.Address, error) {
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
	pub, err := ethcrypto.DecompressPubkey(compressed)
	if err != nil {
		return common.Address{}, fmt.Errorf("decompress pubkey: %w", ErrInvalidDID)
	}
	return ethcrypto.PubkeyToAddress(*pub), nil
}

// ResolveAddress is a backward-compatible package-level function that resolves
// v1 DIDs directly. For v2 DID support, use DefaultAddressResolver.
func ResolveAddress(did string) (common.Address, error) {
	return resolveV1(did)
}
