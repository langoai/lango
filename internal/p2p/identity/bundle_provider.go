package identity

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"sync"
	"time"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/security"
)

// LocalIdentityProvider is the interface for local agent identity operations.
// It combines DID retrieval with signing capabilities. Remote DID verification
// is handled separately by BundleResolver.
type LocalIdentityProvider interface {
	DID(ctx context.Context) (*DID, error)
	Bundle() *IdentityBundle
	LegacyDID(ctx context.Context) (*DID, error)
	SignMessage(ctx context.Context, message []byte) ([]byte, error)
	PublicKey(ctx context.Context) ([]byte, error)
	Algorithm() string
	DIDString(ctx context.Context) (string, error)
}

// BundleProviderConfig holds the configuration for creating a BundleProvider.
type BundleProviderConfig struct {
	SigningKey       ed25519.PrivateKey
	SettlementPub   []byte // compressed secp256k1 public key from wallet
	PQSigningKeySeed []byte // 32-byte HKDF seed for ML-DSA-65 (optional, nil = no PQ)
	LangoDir        string
	Legacy          *WalletDIDProvider
	Logger          *zap.SugaredLogger
}

// BundleProvider manages the local agent's v2 identity. It creates and caches
// the IdentityBundle, derives the DID v2, and provides Ed25519 signing.
//
// BundleProvider does NOT implement VerifyDID for remote peers — that
// responsibility belongs to BundleResolver.
type BundleProvider struct {
	bundle       *IdentityBundle
	signingKey   ed25519.PrivateKey
	pqSigningKey *mldsa65.PrivateKey // nil when PQ unavailable
	legacyProv   *WalletDIDProvider
	did          *DID // cached v2 DID
	langoDir     string
	logger       *zap.SugaredLogger
	mu           sync.RWMutex
}

// Compile-time interface check.
var _ LocalIdentityProvider = (*BundleProvider)(nil)

// NewBundleProvider creates a BundleProvider. If a bundle file exists, it is
// loaded and verified. Otherwise, a new bundle is created, proofs generated,
// and the bundle stored to disk.
func NewBundleProvider(cfg BundleProviderConfig) (*BundleProvider, error) {
	if cfg.SigningKey == nil || len(cfg.SettlementPub) == 0 {
		return nil, fmt.Errorf("signing key and settlement public key are required")
	}

	p := &BundleProvider{
		signingKey: cfg.SigningKey,
		legacyProv: cfg.Legacy,
		langoDir:   cfg.LangoDir,
		logger:     cfg.Logger,
	}

	// Derive ML-DSA-65 PQ signing key from seed if available.
	if len(cfg.PQSigningKeySeed) == mldsa65.SeedSize {
		var seed [mldsa65.SeedSize]byte
		copy(seed[:], cfg.PQSigningKeySeed)
		_, sk := mldsa65.NewKeyFromSeed(&seed)
		security.ZeroBytes(seed[:])
		p.pqSigningKey = sk
	}

	// Try loading existing bundle.
	if cfg.LangoDir != "" {
		existing, err := LoadBundleFile(cfg.LangoDir)
		if err != nil {
			return nil, fmt.Errorf("load identity bundle: %w", err)
		}
		if existing != nil {
			// Verify the signing key matches.
			pub := cfg.SigningKey.Public().(ed25519.PublicKey)
			if string(existing.SigningKey.PublicKey) == string(pub) {
				p.bundle = existing
				return p, nil
			}
			p.logger.Warnw("existing identity bundle has different signing key, regenerating")
		}
	}

	// Create new bundle.
	bundle, err := p.createBundle(cfg.SettlementPub)
	if err != nil {
		return nil, err
	}
	p.bundle = bundle

	// Persist to disk.
	if cfg.LangoDir != "" {
		if err := StoreBundleFile(cfg.LangoDir, bundle); err != nil {
			return nil, fmt.Errorf("store identity bundle: %w", err)
		}
	}

	return p, nil
}

// createBundle constructs a new IdentityBundle with dual proofs.
func (p *BundleProvider) createBundle(settlementPub []byte) (*IdentityBundle, error) {
	pub := p.signingKey.Public().(ed25519.PublicKey)

	// Get legacy DID for the bundle.
	var legacyDID string
	if p.legacyProv != nil {
		d, err := p.legacyProv.DID(context.Background())
		if err == nil && d != nil {
			legacyDID = d.ID
		}
	}

	bundle := &IdentityBundle{
		Version: 1,
		SigningKey: PublicKeyEntry{
			Algorithm: "ed25519",
			PublicKey: []byte(pub),
		},
		SettlementKey: PublicKeyEntry{
			Algorithm: "secp256k1-keccak256",
			PublicKey: settlementPub,
		},
		LegacyDID: legacyDID,
		CreatedAt: time.Now(),
	}

	// Add PQ signing key if available.
	if p.pqSigningKey != nil {
		pqPub := p.pqSigningKey.Public().(*mldsa65.PublicKey)
		pqPubBytes, err := pqPub.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("marshal PQ public key: %w", err)
		}
		bundle.PQSigningKey = &PublicKeyEntry{
			Algorithm: security.AlgorithmMLDSA65,
			PublicKey: pqPubBytes,
		}
	}

	// Generate canonical bytes for proof signing.
	canonical, err := CanonicalBundleBytes(bundle)
	if err != nil {
		return nil, fmt.Errorf("canonical bundle bytes: %w", err)
	}

	// Ed25519 proof.
	bundle.Proofs.Ed25519 = ed25519.Sign(p.signingKey, canonical)

	// ML-DSA-65 proof.
	if p.pqSigningKey != nil {
		pqSig, err := security.SignMLDSA65(p.pqSigningKey, canonical)
		if err != nil {
			return nil, fmt.Errorf("ML-DSA-65 proof: %w", err)
		}
		bundle.Proofs.MLDSA65 = pqSig
	}

	// Legacy proof (secp256k1 via wallet).
	if p.legacyProv != nil && p.legacyProv.keys != nil {
		p.logger.Debugw("legacy proof requires wallet signer — skipping if unavailable")
	}

	return bundle, nil
}

// DID returns the v2 DID for this agent.
func (p *BundleProvider) DID(_ context.Context) (*DID, error) {
	p.mu.RLock()
	if p.did != nil {
		defer p.mu.RUnlock()
		return p.did, nil
	}
	p.mu.RUnlock()

	didStr, err := ComputeDIDv2(p.bundle)
	if err != nil {
		return nil, err
	}

	pub := p.signingKey.Public().(ed25519.PublicKey)
	peerID, err := peerIDFromPublicKey([]byte(pub))
	if err != nil {
		return nil, fmt.Errorf("derive peer ID from Ed25519 key: %w", err)
	}

	did := &DID{
		ID:        didStr,
		PublicKey: []byte(pub),
		PeerID:    peerID,
		Version:   2,
	}

	p.mu.Lock()
	p.did = did
	p.mu.Unlock()

	return did, nil
}

// VerifyDID checks that a v1 DID matches the claimed peer ID.
// For v2 DIDs, use BundleResolver instead.
func (p *BundleProvider) VerifyDID(did *DID, peerID peer.ID) error {
	if did == nil {
		return fmt.Errorf("nil DID")
	}
	if did.Version == 2 {
		return fmt.Errorf("v2 DID verification requires BundleResolver, not BundleProvider")
	}
	if p.legacyProv != nil {
		return p.legacyProv.VerifyDID(did, peerID)
	}
	return fmt.Errorf("no legacy provider for v1 DID verification")
}

// Bundle returns the current IdentityBundle.
func (p *BundleProvider) Bundle() *IdentityBundle {
	return p.bundle
}

// LegacyDID returns the v1 DID for backward compatibility.
func (p *BundleProvider) LegacyDID(ctx context.Context) (*DID, error) {
	if p.legacyProv == nil {
		return nil, fmt.Errorf("no legacy identity provider")
	}
	return p.legacyProv.DID(ctx)
}

// SignMessage signs a message with the Ed25519 identity key.
func (p *BundleProvider) SignMessage(_ context.Context, message []byte) ([]byte, error) {
	return ed25519.Sign(p.signingKey, message), nil
}

// PublicKey returns the Ed25519 public key.
func (p *BundleProvider) PublicKey(_ context.Context) ([]byte, error) {
	return []byte(p.signingKey.Public().(ed25519.PublicKey)), nil
}

// Algorithm returns the signing algorithm identifier.
func (p *BundleProvider) Algorithm() string {
	return "ed25519"
}

// SignPQ signs a message with the ML-DSA-65 PQ signing key.
func (p *BundleProvider) SignPQ(_ context.Context, message []byte) ([]byte, error) {
	if p.pqSigningKey == nil {
		return nil, fmt.Errorf("PQ signing key not available")
	}
	return security.SignMLDSA65(p.pqSigningKey, message)
}

// PQAlgorithm returns the PQ signing algorithm identifier.
func (p *BundleProvider) PQAlgorithm() string {
	return security.AlgorithmMLDSA65
}

// PQPublicKey returns the ML-DSA-65 public key bytes, or nil if unavailable.
func (p *BundleProvider) PQPublicKey() []byte {
	if p.pqSigningKey == nil {
		return nil
	}
	pub := p.pqSigningKey.Public().(*mldsa65.PublicKey)
	b, _ := pub.MarshalBinary()
	return b
}

// HasPQKey reports whether a PQ signing key is available.
func (p *BundleProvider) HasPQKey() bool {
	return p.pqSigningKey != nil
}

// DIDString returns the DID v2 string for the Signer.DID() interface.
func (p *BundleProvider) DIDString(ctx context.Context) (string, error) {
	did, err := p.DID(ctx)
	if err != nil {
		return "", err
	}
	return did.ID, nil
}
