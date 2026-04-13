## Why

Off-chain artifacts (provenance bundles, gossip cards) are currently signed only with classical algorithms (secp256k1-keccak256 or Ed25519). Gossip cards are unsigned entirely. A future quantum computer could forge classical signatures, breaking artifact authenticity. Adding ML-DSA-65 (FIPS 204, NIST Level 3) post-quantum signatures as dual signatures alongside classical ones provides quantum-resistant authenticity while maintaining backward compatibility.

## What Changes

- Add ML-DSA-65 signature scheme (`VerifyMLDSA65`, `SignMLDSA65`) and PQ key derivation from Master Key
- Extend IdentityBundle with `PQSigningKey` and `PQGeneration` (excluded from canonical bytes — DID v2 hash unchanged)
- Add `PQBundleSigner` optional interface for dual-signing provenance bundles
- Add self-contained PQ verification: embed `PQSignerPublicKey` in artifacts for rotation-safe verification
- Add GossipCard signing (classical + PQ dual signatures, unsigned legacy cards accepted)
- Bootstrap: new `phaseDerivePQKey` phase (11→12 phases)
- No config flag — PQ signing auto-enabled when MK is available

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `p2p-identity` — PQSigningKey + PQGeneration in IdentityBundle, MLDSA65 proof in BundleProofs
- `session-provenance` — PQ dual signatures on provenance bundles (PQSignerPublicKey + PQSignature)
- `p2p-discovery` — GossipCard signing with canonical payload, classical + PQ signatures
- `cli-security-status` — PQ signing key status display

## Impact

- **Code**: `internal/security/` (ML-DSA scheme + PQ key derivation), `internal/p2p/identity/` (bundle PQ extension), `internal/provenance/` (dual signatures), `internal/p2p/discovery/` (card signing), `internal/bootstrap/` (PQ key phase), `internal/app/` (wiring), `internal/cli/` (status)
- **Dependencies**: `cloudflare/circl` already direct (from Phase 4) — `sign/mldsa/mldsa65` subpackage now used
- **Wire format**: Backward compatible — new `omitempty` fields, unsigned legacy cards accepted
- **Size overhead**: +1952B (PQ pubkey) + 3309B (PQ sig) per artifact — acceptable for bundles and gossip
