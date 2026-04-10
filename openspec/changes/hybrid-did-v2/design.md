## Context

Phase 0 (boundary cleanup), Phase 1 (MK/KEK), Phase 2 (algorithm agility) 완료 및 archived. DID v2 + IdentityBundle 도입으로 agent identity를 wallet secp256k1 키에서 분리.

## Goals / Non-Goals

**Goals:**
- Content-addressed DID v2 format (`did:lango:v2:<hash>`)
- IdentityBundle (Ed25519 signing + secp256k1 settlement + dual proofs)
- v1/v2 공존 + DID alias로 session/reputation 연속성
- Economy/escrow에서 v2 DID → settlement address 해석
- Handshake에서 v2 DID + Bundle transport + outbound identity selection

**Non-Goals:**
- ML-DSA/PQC 키 (Phase 5)
- ML-KEM/KEM provider (Phase 4)
- DID v1 제거 (장기 deprecation)
- W3C DID Document 표준 준수

## Decisions

### D1: DID v2 = content-addressed, PeerID 분리
- DID v2 ID = SHA-256(canonical bundle)[:20] hex
- Canonical bytes: Version + SigningKey + SettlementKey + LegacyDID (CreatedAt, Proofs 제외)
- DID v2 struct에서 PeerID는 비어있음 — transport PeerID(node key) ≠ identity key
- PeerID 매핑은 BundleResolver + GossipCard에서 해결

### D2: Ed25519 identity key = HKDF(MK, domain, generation)
- `DeriveIdentityKey(mk, generation)` — HKDF(SHA256, MK, nil, "lango-identity-ed25519[:N]")
- generation 초기값 0, identity-bundle.json에 저장
- MK recovery = identity recovery. MK 같으면 identity 같음 (의도적)

### D3: Interface 3분할
- `LocalIdentityProvider`: DID, PublicKey, SignMessage, Algorithm, DIDString, Bundle, LegacyDID
- `BundleResolver`: ResolveBundle(did) → *IdentityBundle (remote peer용)
- `AddressResolver`: ResolveAddress(did) → common.Address (settlement용)

### D4: Handshake outbound identity selection
- Config에 `Signer` (v2 Ed25519) + `LegacySigner` (v1 secp256k1)
- selectSigner(peerAlgo): empty/secp256k1 → LegacySigner, ed25519 → Signer
- Unknown peer → LegacySigner (safe default)

### D5: Bundle transport in handshake
- Challenge/ChallengeResponse에 `Bundle *IdentityBundle` omitempty 필드
- Handshake 시 자동 bundle 교환 → MemoryBundleCache에 저장
- GossipCard에도 optional Bundle 필드

### D6: DID alias for session/reputation continuity
- DIDAlias maps v2 DID ↔ v1 DID (via bundle.LegacyDID)
- SessionStore, reputation, firewall에서 CanonicalDID(did) 사용
- v1 DID가 canonical key (기존 데이터 보존)

### D7: Bundle 영속성
- 로컬 bundle: `~/.lango/identity-bundle.json` (atomic write, 0600)
- Remote bundles: `~/.lango/known-bundles/<did-hash>.json` + in-memory cache
- Deal/authorization에 settlement address 보존 (resolve 1회)

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| ParseDID v2 hollow parse (empty PeerID) | v2 DID는 new agent에서만 생성 → 기존 코드 runtime 무영향 |
| v2→v1 handshake (Ed25519→secp256k1 peer) | LegacySigner fallback, unknown peer → secp256k1 |
| ResolveAddress v2 bundle 미보유 | known-bundles persistent store + deal에 address 보존 |
| Bootstrap 11→12 phases | phase count 테스트 업데이트 |
| identity_generation rotation complexity | 초기 generation=0, rotation은 Phase 6+ |
