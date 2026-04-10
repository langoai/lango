## Context

Phase 0 (boundary cleanup), Phase 1 (MK/KEK), Phase 2 (algorithm agility) completed and archived. DID v2 + IdentityBundle introduction separates agent identity from the wallet secp256k1 key.

## Goals / Non-Goals

**Goals:**
- Content-addressed DID v2 format (`did:lango:v2:<hash>`)
- IdentityBundle (Ed25519 signing + secp256k1 settlement + dual proofs)
- v1/v2 coexistence + DID alias for session/reputation continuity
- Economy/escrow v2 DID -> settlement address resolution
- Handshake v2 DID + Bundle transport + outbound identity selection

**Non-Goals:**
- ML-DSA/PQC keys (Phase 5)
- ML-KEM/KEM provider (Phase 4)
- DID v1 removal (long-term deprecation)
- W3C DID Document standard compliance

## Decisions

### D1: DID v2 = content-addressed, PeerID separated
- DID v2 ID = SHA-256(canonical bundle)[:20] hex
- Canonical bytes: Version + SigningKey + SettlementKey + LegacyDID (CreatedAt, Proofs excluded)
- DID v2 struct has empty PeerID — transport PeerID (node key) != identity key
- PeerID mapping resolved via BundleResolver + GossipCard

### D2: Ed25519 identity key = HKDF(MK, domain, generation)
- `DeriveIdentityKey(mk, generation)` — HKDF(SHA256, MK, nil, "lango-identity-ed25519[:N]")
- Generation defaults to 0, stored in identity-bundle.json
- MK recovery = identity recovery. Same MK always produces same identity (intentional)

### D3: Interface 3-way split
- `LocalIdentityProvider`: DID, PublicKey, SignMessage, Algorithm, DIDString, Bundle, LegacyDID
- `BundleResolver`: ResolveBundle(did) -> *IdentityBundle (for remote peers)
- `AddressResolver`: ResolveAddress(did) -> common.Address (for settlement)

### D4: Handshake outbound identity selection
- Config has `Signer` (v2 Ed25519) + `LegacySigner` (v1 secp256k1)
- selectSigner(peerAlgo): empty/secp256k1 -> LegacySigner, ed25519 -> Signer
- Unknown peer -> LegacySigner (safe default)

### D5: Bundle transport in handshake
- Challenge/ChallengeResponse include `Bundle *IdentityBundle` omitempty field
- Bundles automatically exchanged during handshake -> stored in MemoryBundleCache
- GossipCard also has optional Bundle field

### D6: DID alias for session/reputation continuity
- DIDAlias maps v2 DID <-> v1 DID (via bundle.LegacyDID)
- SessionStore, reputation, firewall use CanonicalDID(did)
- v1 DID is the canonical key (preserves existing data)

### D7: Bundle persistence
- Local bundle: `~/.lango/identity-bundle.json` (atomic write, 0600)
- Remote bundles: `~/.lango/known-bundles/<did-hash>.json` + in-memory cache
- Deal/authorization preserves settlement address (resolve once)

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| ParseDID v2 hollow parse (empty PeerID) | v2 DIDs only created by new agents -> no runtime impact on existing code |
| v2->v1 handshake (Ed25519->secp256k1 peer) | LegacySigner fallback, unknown peer -> secp256k1 |
| ResolveAddress v2 bundle unavailable | known-bundles persistent store + address preserved in deal record |
| Bootstrap 10->11 phases | phase count test updated |
| identity_generation rotation complexity | Initial generation=0, rotation deferred to Phase 6+ |
