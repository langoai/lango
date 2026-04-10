## Why

현재 `did:lango:<secp256k1-hex>` 포맷은 wallet의 secp256k1 키 하나에 agent 전체 identity가 묶여 있다. Phase 0에서 wallet 의존을 분리하고 Phase 2에서 algorithm agility를 도입했지만, DID 포맷 자체가 secp256k1 공개키를 직접 인코딩하므로 새 알고리즘(Ed25519, ML-DSA)을 identity에 적용할 수 없다. 또한 passphrase 분실 시 wallet key와 identity가 동시에 소실되고, key rotation 시 DID가 변경된다.

## What Changes

- **DID v2 format**: `did:lango:v2:<40-hex>` (SHA-256(canonical bundle)[:20] bytes, content-addressed)
- **IdentityBundle**: Ed25519 signing key + secp256k1 settlement key + legacy DID + dual proofs (공개 정보)
- **Ed25519 identity key**: `HKDF(MK, "lango-identity-ed25519", generation)` — MK recovery = identity recovery
- **ParseDID v1/v2 dispatcher**: 기존 v1 DID 계속 지원, v2 DID는 BundleResolver로 해석
- **BundleProvider**: 로컬 identity 관리 (LocalIdentityProvider). Remote DID 해석은 BundleResolver
- **Handshake v2**: Signer.DID() method, LegacySigner fallback, Bundle transport in Challenge/ChallengeResponse
- **Economy/Escrow**: AddressResolver interface, v2 DID → bundle → settlement key → Ethereum address
- **DID alias**: v1/v2 DID 매핑으로 session/reputation 연속성 보장
- **PeerID 분리**: DID v2의 PeerID는 transport routing identifier(node key 기반), identity key에서 유도하지 않음

**`did:lango` v1은 제거하지 않음.** 새 agent는 v2 + legacy v1 동시 발급. 기존 agent는 v1 유지.

## Capabilities

### New Capabilities

(없음 — p2p-identity 스펙 내에서 IdentityBundle + DID v2 커버)

### Modified Capabilities

- `p2p-identity`: DID v2 format, IdentityBundle type, BundleProvider (LocalIdentityProvider), BundleResolver, ParseDID v1/v2 dispatcher, peerIDFromPublicKey multi-algo, DIDAlias, ComputeDIDv2
- `p2p-handshake`: Signer.DID() method, LegacySigner in Config, Bundle field in Challenge/ChallengeResponse, Ed25519 default verifier
- `escrow-settlement`: AddressResolver interface, v2 DID → bundle → settlement key → address

## Impact

- **코드:** `internal/p2p/identity/`, `internal/p2p/handshake/`, `internal/economy/escrow/`, `internal/app/`, `internal/bootstrap/`, `internal/security/`, `internal/cli/security/`, `internal/p2p/discovery/`, `internal/a2a/`
- **프로토콜:** Challenge/ChallengeResponse에 `Bundle` omitempty 필드 추가 (backward compat). DID v2 문자열이 프로토콜 메시지에 등장.
- **파일시스템:** `~/.lango/identity-bundle.json` (0600), `~/.lango/known-bundles/` directory
- **Bootstrap:** 11→12 phases (phaseDeriveIdentityKey 추가)
- **의존성:** 새 외부 의존성 없음 (`crypto/ed25519`는 Go stdlib)
