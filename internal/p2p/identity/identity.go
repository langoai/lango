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

	"github.com/langowarny/lango/internal/wallet"
)

const (
	// DIDPrefix is the method-specific prefix for Lango DIDs.
	DIDPrefix = "did:lango:"
)

// DID represents a decentralized identifier derived from a wallet public key.
type DID struct {
	ID        string  `json:"id"`        // "did:lango:<hex-compressed-pubkey>"
	PublicKey []byte  `json:"publicKey"` // compressed secp256k1 public key
	PeerID    peer.ID `json:"peerId"`    // libp2p peer ID derived from pubkey
}

// Provider creates and verifies DIDs.
type Provider interface {
	// DID returns the DID for the current wallet.
	DID(ctx context.Context) (*DID, error)
	// VerifyDID checks that a DID matches the claimed peer ID.
	VerifyDID(did *DID, peerID peer.ID) error
}

// WalletDIDProvider derives DIDs from a wallet's public key.
type WalletDIDProvider struct {
	wallet wallet.WalletProvider
	logger *zap.SugaredLogger
	mu     sync.RWMutex
	cached *DID
}

// Compile-time interface check.
var _ Provider = (*WalletDIDProvider)(nil)

// NewProvider creates a new WalletDIDProvider.
func NewProvider(w wallet.WalletProvider, logger *zap.SugaredLogger) *WalletDIDProvider {
	return &WalletDIDProvider{
		wallet: w,
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

	pubkey, err := p.wallet.PublicKey(ctx)
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

// ParseDID parses a "did:lango:<hexkey>" string into a DID.
func ParseDID(didStr string) (*DID, error) {
	if !strings.HasPrefix(didStr, DIDPrefix) {
		return nil, fmt.Errorf("invalid DID scheme: expected prefix %q, got %q", DIDPrefix, didStr)
	}

	hexKey := strings.TrimPrefix(didStr, DIDPrefix)
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
	}, nil
}

// DIDFromPublicKey creates a DID from a compressed secp256k1 public key.
func DIDFromPublicKey(pubkey []byte) (*DID, error) {
	if len(pubkey) == 0 {
		return nil, fmt.Errorf("empty public key")
	}

	peerID, err := peerIDFromPublicKey(pubkey)
	if err != nil {
		return nil, fmt.Errorf("derive peer ID: %w", err)
	}

	return &DID{
		ID:        DIDPrefix + hex.EncodeToString(pubkey),
		PublicKey: pubkey,
		PeerID:    peerID,
	}, nil
}

// peerIDFromPublicKey derives a libp2p peer ID from a compressed secp256k1 public key.
func peerIDFromPublicKey(pubkey []byte) (peer.ID, error) {
	libp2pKey, err := crypto.UnmarshalSecp256k1PublicKey(pubkey)
	if err != nil {
		return "", fmt.Errorf("unmarshal secp256k1 public key: %w", err)
	}

	peerID, err := peer.IDFromPublicKey(libp2pKey)
	if err != nil {
		return "", fmt.Errorf("derive peer ID from public key: %w", err)
	}

	return peerID, nil
}
