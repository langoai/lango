## Why

The current `did:lango:<secp256k1-hex>` format ties the entire agent identity to a single wallet secp256k1 key. Phase 0 removed wallet dependencies and Phase 2 introduced algorithm agility, but the DID format itself directly encodes the secp256k1 public key, preventing new algorithms (Ed25519, ML-DSA) from being used for identity. Additionally, losing the passphrase causes simultaneous loss of wallet key and identity, and key rotation changes the DID.

## What Changes

- **DID v2 format**: `did:lango:v2:<40-hex>` (SHA-256(canonical bundle)[:20] bytes, content-addressed)
- **IdentityBundle**: Ed25519 signing key + secp256k1 settlement key + legacy DID + dual proofs (public information)
- **Ed25519 identity key**: `HKDF(MK, "lango-identity-ed25519", generation)` — MK recovery = identity recovery
- **ParseDID v1/v2 dispatcher**: continues to support v1 DIDs, v2 DIDs resolved via BundleResolver
- **BundleProvider**: local identity management (LocalIdentityProvider). Remote DID resolution via BundleResolver
- **Handshake v2**: Signer.DID() method, LegacySigner fallback, Bundle transport in Challenge/ChallengeResponse
- **Economy/Escrow**: AddressResolver interface, v2 DID -> bundle -> settlement key -> Ethereum address
- **DID alias**: v1/v2 DID mapping for session/reputation continuity
- **PeerID separation**: DID v2's PeerID is a transport routing identifier (node key based), not derived from identity key

**`did:lango` v1 is NOT removed.** New agents issue v2 + legacy v1 simultaneously. Existing agents keep v1.

## Capabilities

### New Capabilities

(None — IdentityBundle + DID v2 covered within the p2p-identity spec)

### Modified Capabilities

- `p2p-identity`: DID v2 format, IdentityBundle type, BundleProvider (LocalIdentityProvider), BundleResolver, ParseDID v1/v2 dispatcher, peerIDFromPublicKey multi-algo, DIDAlias, ComputeDIDv2
- `p2p-handshake`: Signer.DID() method, LegacySigner in Config, Bundle field in Challenge/ChallengeResponse, Ed25519 default verifier
- `escrow-settlement`: AddressResolver interface, v2 DID -> bundle -> settlement key -> address

## Impact

- **Code:** `internal/p2p/identity/`, `internal/p2p/handshake/`, `internal/economy/escrow/`, `internal/app/`, `internal/bootstrap/`, `internal/security/`, `internal/cli/security/`, `internal/p2p/discovery/`, `internal/a2a/`
- **Protocol:** `Bundle` omitempty field added to Challenge/ChallengeResponse (backward compatible). DID v2 strings appear in protocol messages.
- **Filesystem:** `~/.lango/identity-bundle.json` (0600), `~/.lango/known-bundles/` directory
- **Bootstrap:** 10 -> 11 phases (phaseDeriveIdentityKey added)
- **Dependencies:** No new external dependencies (`crypto/ed25519` is Go stdlib)
