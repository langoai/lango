## Why

Handshake challenge/response verification과 identity signature verification이 secp256k1+keccak256에 하드코딩되어 있다. Phase 0에서 wallet 의존은 제거했지만, 알고리즘 자체는 여전히 고정. Phase 3 (DID v2)와 Phase 5 (PQ signatures)에서 새 알고리즘을 추가하려면, 서명 알고리즘을 교체 가능하게 만드는 framework이 먼저 필요하다.

추가 발견: challenge signature에 double-hash 버그가 있어 signed challenge 검증이 항상 실패. 이번에 같이 수정.

## What Changes

- `internal/security/`에 `SignatureScheme` 타입 + 알고리즘 상수 (`AlgorithmSecp256k1Keccak256`, `AlgorithmEd25519`) + Verify 함수 구현 추가
- `internal/p2p/handshake/`: `Signer` interface에 `Algorithm()` 추가, `Challenge`/`ChallengeResponse`에 `SignatureAlgorithm` 필드 추가 (backward compat), verifier map dispatch 도입, challenge double-hash 버그 수정
- `internal/p2p/identity/`: `ParseDIDPublicKey` 함수 추가 (peerID 없이 pubkey 추출)
- `internal/provenance/` + wiring: Ed25519 verifier를 wiring closure로 등록 (framework 검증용, production 기능 아님)
- `.golangci.yml`: handshake/identity를 p2p-infra-no-economy 규칙에 추가
- `provenance.AlgorithmSecp256k1Keccak256` → `security.AlgorithmSecp256k1Keccak256`로 canonical source 이전

**`did:lango` 포맷은 secp256k1 전용으로 유지. Ed25519는 framework 검증용 integration test까지만.**

## Capabilities

### New Capabilities

- `signature-algorithms`: SignatureScheme 타입, 알고리즘 상수, secp256k1-keccak256 및 Ed25519 Verify 함수

### Modified Capabilities

- `p2p-handshake`: Signer interface에 Algorithm() 추가, Challenge/ChallengeResponse에 SignatureAlgorithm 필드 추가, verifier map dispatch, challenge double-hash 버그 수정
- `p2p-identity`: ParseDIDPublicKey 함수 추가 (DID 포맷/peerID 무변경)
- `session-provenance`: Ed25519 verifier 등록 (wiring closure, framework 검증용)
- `architecture-boundary-enforcement`: depguard에 handshake/identity 추가, 구식 NOTE 제거

## Impact

- **코드:** `internal/security/`, `internal/p2p/handshake/`, `internal/p2p/identity/`, `internal/provenance/`, `internal/app/`, `internal/cli/provenance/`, `.golangci.yml`
- **프로토콜:** Challenge/ChallengeResponse에 omitempty 필드 추가 (backward compat). 기존 peer와 호환.
- **API:** Signer interface에 Algorithm() 추가 — 기존 WalletProvider는 implicit satisfaction 깨짐 → wiring wrapper 필요
- **의존성:** 새 외부 의존성 없음 (`crypto/ed25519`는 Go stdlib)
