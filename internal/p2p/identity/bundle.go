package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/langoai/lango/internal/types"
)

// IdentityBundle is a public identity document containing signing and
// settlement keys with dual proofs. It is not secret — the bundle is
// shared with peers via handshake and gossip.
type IdentityBundle struct {
	Version       int            `json:"version"`
	Generation    uint32         `json:"generation"`               // Ed25519 key derivation generation (default 0)
	SigningKey    PublicKeyEntry `json:"signing_key"`              // Ed25519 primary signing key
	SettlementKey PublicKeyEntry `json:"settlement_key"`           // secp256k1 (from wallet)
	LegacyDID     string         `json:"legacy_did,omitempty"`     // did:lango:<secp256k1-hex> for v1 compat
	PQGeneration  uint32         `json:"pq_generation,omitempty"`  // ML-DSA key derivation generation (default 0)
	PQSigningKey  *PublicKeyEntry `json:"pq_signing_key,omitempty"` // ML-DSA-65 PQ signing key (nil if unavailable)
	Proofs        BundleProofs   `json:"proofs"`
	CreatedAt     time.Time      `json:"created_at"`
}

// PublicKeyEntry describes a public key with its algorithm.
type PublicKeyEntry struct {
	Algorithm string `json:"algorithm"`
	PublicKey []byte `json:"public_key"`
}

// BundleProofs contains ownership proofs over the canonical bundle.
type BundleProofs struct {
	Legacy  []byte `json:"legacy,omitempty"`  // secp256k1+keccak256 signature over canonical
	Ed25519 []byte `json:"ed25519,omitempty"` // Ed25519 signature over canonical
	MLDSA65 []byte `json:"mldsa65,omitempty"` // ML-DSA-65 signature over canonical
}

// canonicalBundleData is the subset of IdentityBundle fields included in the
// DID v2 hash. CreatedAt and Proofs are excluded for determinism.
type canonicalBundleData struct {
	Version       int            `json:"version"`
	SigningKey    PublicKeyEntry `json:"signing_key"`
	SettlementKey PublicKeyEntry `json:"settlement_key"`
	LegacyDID     string         `json:"legacy_did,omitempty"`
}

// CanonicalBundleBytes returns the deterministic JSON encoding of the
// canonical bundle fields (Version, SigningKey, SettlementKey, LegacyDID).
// CreatedAt and Proofs are excluded so the same key set always produces
// the same bytes.
func CanonicalBundleBytes(b *IdentityBundle) ([]byte, error) {
	if b == nil {
		return nil, fmt.Errorf("nil identity bundle")
	}
	canonical := canonicalBundleData{
		Version:       b.Version,
		SigningKey:    b.SigningKey,
		SettlementKey: b.SettlementKey,
		LegacyDID:     b.LegacyDID,
	}
	return json.Marshal(canonical)
}

// ComputeDIDv2 computes the content-addressed DID v2 string from an
// IdentityBundle. The ID is SHA-256(canonical bytes)[:20] hex-encoded.
// Same key set + same legacy DID always produces the same DID v2.
func ComputeDIDv2(b *IdentityBundle) (string, error) {
	canonical, err := CanonicalBundleBytes(b)
	if err != nil {
		return "", fmt.Errorf("compute DID v2: %w", err)
	}
	h := sha256.Sum256(canonical)
	return types.DIDv2Prefix + hex.EncodeToString(h[:20]), nil
}

// BundleResolver looks up IdentityBundles for remote peers by DID v2 string.
// Implementations are populated during handshakes and gossip.
type BundleResolver interface {
	ResolveBundle(did string) (*IdentityBundle, error)
}

// MemoryBundleCache is a simple in-memory BundleResolver populated during
// handshakes and gossip card exchanges.
type MemoryBundleCache struct {
	mu      sync.RWMutex
	bundles map[string]*IdentityBundle // key: DID v2 string
}

// NewMemoryBundleCache creates a new empty bundle cache.
func NewMemoryBundleCache() *MemoryBundleCache {
	return &MemoryBundleCache{
		bundles: make(map[string]*IdentityBundle),
	}
}

// Store caches an IdentityBundle under its DID v2.
func (c *MemoryBundleCache) Store(didV2 string, bundle *IdentityBundle) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bundles[didV2] = bundle
}

// ResolveBundle looks up a bundle by DID v2 string.
func (c *MemoryBundleCache) ResolveBundle(did string) (*IdentityBundle, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	b, ok := c.bundles[did]
	if !ok {
		return nil, fmt.Errorf("bundle not found for DID %q", did)
	}
	return b, nil
}

// DIDAlias maps v2 DID ↔ v1 DID for session/reputation continuity.
// When a peer has both a v1 and v2 DID, CanonicalDID returns the v1 DID
// to preserve existing session and reputation data.
type DIDAlias struct {
	mu     sync.RWMutex
	v2ToV1 map[string]string
	v1ToV2 map[string]string
}

// NewDIDAlias creates a new empty alias registry.
func NewDIDAlias() *DIDAlias {
	return &DIDAlias{
		v2ToV1: make(map[string]string),
		v1ToV2: make(map[string]string),
	}
}

// RegisterFromBundle registers the v2 ↔ v1 alias from an IdentityBundle.
func (a *DIDAlias) RegisterFromBundle(bundle *IdentityBundle, didV2 string) {
	if bundle == nil || bundle.LegacyDID == "" || didV2 == "" {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.v2ToV1[didV2] = bundle.LegacyDID
	a.v1ToV2[bundle.LegacyDID] = didV2
}

// CanonicalDID returns the canonical DID for session/reputation lookups.
// If the input is a v2 DID with a known v1 alias, returns the v1 DID.
// Otherwise returns the input unchanged.
func (a *DIDAlias) CanonicalDID(did string) string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if v1, ok := a.v2ToV1[did]; ok {
		return v1
	}
	return did
}
