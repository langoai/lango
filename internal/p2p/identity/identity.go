// Package identity provides decentralized identity (DID) derivation from wallet public keys.
// DIDs are deterministically derived from compressed secp256k1 public keys and mapped to
// libp2p peer IDs for P2P networking. Private keys never leave the wallet layer.
package identity

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/types"
)

// DID represents a decentralized identifier.
// For v1 (did:lango:<hex>): PublicKey and PeerID are populated from the DID string.
// For v2 (did:lango:v2:<hash>): PublicKey and PeerID are empty (requires BundleResolver).
type DID struct {
	ID        string  `json:"id"`                  // "did:lango:<hex>" or "did:lango:v2:<hash>"
	PublicKey []byte  `json:"publicKey,omitempty"` // signing key (v1: secp256k1, v2: empty until resolved)
	PeerID    peer.ID `json:"peerId,omitempty"`    // libp2p peer ID (v1: derived, v2: empty until resolved)
	Version   int     `json:"version"`             // 1 or 2
}

// KeyProvider is the minimal interface for public key retrieval.
// wallet.WalletProvider satisfies this via Go structural typing.
type KeyProvider interface {
	PublicKey(ctx context.Context) ([]byte, error)
}

// Provider creates and verifies DIDs.
type Provider interface {
	// DID returns the DID for the current identity key.
	DID(ctx context.Context) (*DID, error)
	// VerifyDID checks that a DID matches the claimed peer ID.
	VerifyDID(did *DID, peerID peer.ID) error
}

// WalletDIDProvider derives DIDs from a public key provider.
type WalletDIDProvider struct {
	keys   KeyProvider
	logger *zap.SugaredLogger
	mu     sync.RWMutex
	cached *DID
}

// Compile-time interface check.
var _ Provider = (*WalletDIDProvider)(nil)

// NewProvider creates a new WalletDIDProvider.
func NewProvider(keys KeyProvider, logger *zap.SugaredLogger) *WalletDIDProvider {
	return &WalletDIDProvider{
		keys:   keys,
		logger: logger,
	}
}

// DID returns the DID for the current wallet, caching the result since the
// wallet key does not change.
func (p *WalletDIDProvider) DID(ctx context.Context) (*DID, error) {
	p.mu.RLock()
	if p.cached != nil {
		defer p.mu.RUnlock()
		return p.cached, nil
	}
	p.mu.RUnlock()

	pubkey, err := p.keys.PublicKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("get wallet public key: %w", err)
	}

	did, err := DIDFromPublicKey(pubkey)
	if err != nil {
		return nil, fmt.Errorf("derive DID from public key: %w", err)
	}

	p.mu.Lock()
	p.cached = did
	p.mu.Unlock()

	p.logger.Infow("derived DID from wallet", "did", did.ID, "peerID", did.PeerID)
	return did, nil
}

// VerifyDID checks that a DID's public key produces the claimed peer ID.
func (p *WalletDIDProvider) VerifyDID(did *DID, peerID peer.ID) error {
	if did == nil {
		return fmt.Errorf("nil DID")
	}

	derivedPeerID, err := peerIDFromPublicKey(did.PublicKey)
	if err != nil {
		return fmt.Errorf("derive peer ID from DID public key: %w", err)
	}

	if derivedPeerID != peerID {
		return fmt.Errorf("peer ID mismatch: DID derives %s, claimed %s", derivedPeerID, peerID)
	}

	return nil
}

// ParseDIDPublicKey extracts the raw public key bytes from a v1 DID string
// without deriving a peer ID. Returns an error for v2 DIDs (content-addressed,
// no embedded public key — use BundleResolver instead).
func ParseDIDPublicKey(didStr string) ([]byte, error) {
	if strings.HasPrefix(didStr, types.DIDv2Prefix) {
		return nil, fmt.Errorf("DID v2 does not embed a public key; use BundleResolver")
	}
	if !strings.HasPrefix(didStr, types.DIDPrefix) {
		return nil, fmt.Errorf("invalid DID scheme: expected prefix %q, got %q", types.DIDPrefix, didStr)
	}
	hexKey := strings.TrimPrefix(didStr, types.DIDPrefix)
	if hexKey == "" {
		return nil, fmt.Errorf("empty public key in DID %q", didStr)
	}
	return hex.DecodeString(hexKey)
}

// ParseDID parses a DID string into a DID struct. Supports both v1 and v2 formats.
// V1 (did:lango:<hex>): PublicKey and PeerID are populated from the embedded key.
// V2 (did:lango:v2:<hash>): Version=2, PublicKey=nil, PeerID="" (requires BundleResolver).
func ParseDID(didStr string) (*DID, error) {
	if strings.HasPrefix(didStr, types.DIDv2Prefix) {
		return parseDIDv2(didStr)
	}
	return parseDIDv1(didStr)
}

// parseDIDv1 parses a v1 DID with an embedded secp256k1 public key.
func parseDIDv1(didStr string) (*DID, error) {
	if !strings.HasPrefix(didStr, types.DIDPrefix) {
		return nil, fmt.Errorf("invalid DID scheme: expected prefix %q, got %q", types.DIDPrefix, didStr)
	}

	hexKey := strings.TrimPrefix(didStr, types.DIDPrefix)
	if hexKey == "" {
		return nil, fmt.Errorf("empty public key in DID %q", didStr)
	}

	pubkey, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("decode hex public key: %w", err)
	}

	peerID, err := peerIDFromPublicKey(pubkey)
	if err != nil {
		return nil, fmt.Errorf("derive peer ID: %w", err)
	}

	return &DID{
		ID:        didStr,
		PublicKey: pubkey,
		PeerID:    peerID,
		Version:   1,
	}, nil
}

// parseDIDv2 parses a v2 content-addressed DID. The DID string contains a hash,
// not a public key. PublicKey and PeerID are left empty — the caller must resolve
// them via BundleResolver.
func parseDIDv2(didStr string) (*DID, error) {
	hashHex := strings.TrimPrefix(didStr, types.DIDv2Prefix)
	if hashHex == "" {
		return nil, fmt.Errorf("empty hash in DID v2 %q", didStr)
	}
	if len(hashHex) != 40 {
		return nil, fmt.Errorf("invalid DID v2 hash length: expected 40 hex chars, got %d", len(hashHex))
	}
	if _, err := hex.DecodeString(hashHex); err != nil {
		return nil, fmt.Errorf("invalid DID v2 hash hex: %w", err)
	}
	return &DID{
		ID:      didStr,
		Version: 2,
		// PublicKey and PeerID intentionally empty — resolve via BundleResolver
	}, nil
}

// DIDFromPublicKey creates a v1 DID from a compressed secp256k1 public key.
func DIDFromPublicKey(pubkey []byte) (*DID, error) {
	if len(pubkey) == 0 {
		return nil, fmt.Errorf("empty public key")
	}

	peerID, err := peerIDFromPublicKey(pubkey)
	if err != nil {
		return nil, fmt.Errorf("derive peer ID: %w", err)
	}

	return &DID{
		ID:        types.DIDPrefix + hex.EncodeToString(pubkey),
		PublicKey: pubkey,
		PeerID:    peerID,
		Version:   1,
	}, nil
}

// peerIDFromPublicKey derives a libp2p peer ID from a public key.
// Supports secp256k1 (33 bytes compressed) and Ed25519 (32 bytes).
func peerIDFromPublicKey(pubkey []byte) (peer.ID, error) {
	var libp2pKey crypto.PubKey
	var err error
	switch len(pubkey) {
	case 33: // compressed secp256k1
		libp2pKey, err = crypto.UnmarshalSecp256k1PublicKey(pubkey)
	case 32: // Ed25519
		libp2pKey, err = crypto.UnmarshalEd25519PublicKey(pubkey)
	default:
		return "", fmt.Errorf("unsupported public key size: %d bytes", len(pubkey))
	}
	if err != nil {
		return "", fmt.Errorf("unmarshal public key: %w", err)
	}

	peerID, err := peer.IDFromPublicKey(libp2pKey)
	if err != nil {
		return "", fmt.Errorf("derive peer ID from public key: %w", err)
	}

	return peerID, nil
}
