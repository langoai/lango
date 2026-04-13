## Context

Phase 0-4 complete. Provenance bundles have single classical signature. GossipCards are unsigned. ML-DSA-65 available via `circl v1.6.1`.

## Goals / Non-Goals

**Goals:**
- ML-DSA-65 dual signatures on provenance bundles and gossip cards
- Self-contained PQ verification (embedded PQ pubkey in artifacts, rotation-safe)
- PQ key derivation from Master Key via HKDF (separate generation from Ed25519)
- IdentityBundle PQ extension without changing DID v2 hash

**Non-Goals:**
- AgentCard signing (HTTP-based, different trust model — Phase 6+)
- Handshake ML-DSA verifier (transport/auth scope — Phase 6+)
- PQ key rotation mechanism (generation infrastructure only, actual rotation Phase 7+)
- Message encryption with PQ (Phase 4b scope)

## Decisions

### D1: ML-DSA-65 via circl
FIPS 204, NIST Level 3. Seed=32B, PubKey=1952B, Sig=3309B. `NewKeyFromSeed`, `SignTo`, `Verify`.

### D2: Self-contained artifact verification
Embed `PQSignerPublicKey` in artifacts. Classical canonical payload includes PQ pubkey (excludes only Signature + PQSignature). Classical sig authenticates the PQ pubkey → trust chain.

### D3: Separate generation counters
`IdentityBundle.Generation` (Ed25519) + `IdentityBundle.PQGeneration` (ML-DSA). Independent rotation.

### D4: PQBundleSigner optional interface
Type assertion pattern (`BundleAttacher` precedent). No breaking change to `BundleSigner`.

### D5: CanonicalBundleBytes unchanged
PQSigningKey + PQGeneration excluded → DID v2 hash stable.

### D6: GossipCard canonical includes Bundle
Prevents bundle substitution. Excludes only Signature + PQSignature.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| DID v2 hash instability | Unit test: CanonicalBundleBytes unchanged with/without PQ key |
| PQ key rotation breaks past artifacts | Self-contained verification (embedded pubkey) |
| GossipSub message size +5KB | Within 1MB limit |
| Old peers ignore PQ fields | Classical sig always present; PQ optional |
