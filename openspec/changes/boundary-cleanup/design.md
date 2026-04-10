## Context

Phase 1 (master-key-envelope)에서 저장 암호화 계층을 MK/KEK로 분리 완료. 다음 단계인 Phase 2 (Algorithm Agility)와 Phase 3 (Hybrid DID v2)를 진행하려면, wallet secp256k1 키가 settlement 외 영역에 직접 침투한 결합도를 먼저 해소해야 한다.

현재 결합 구조:
- `p2p/handshake` → `wallet` (imports WalletProvider for signing)
- `p2p/identity` → `wallet` (imports WalletProvider for public key)
- `provenance` → `p2p/identity` (imports VerifyMessageSignature)
- `archtest`의 boundary rule이 trailing slash 버그로 wallet import를 놓치고 있음

## Goals / Non-Goals

**Goals:**
- `internal/p2p/handshake/`에서 `internal/wallet` import 제거
- `internal/p2p/identity/`에서 `internal/wallet` import 제거
- `internal/provenance/`에서 `internal/p2p/identity` import 제거 (의존 역전)
- handshake response verification을 injectable function으로 추출
- archtest boundary enforcement 강화

**Non-Goals:**
- `wallet.WalletProvider` 인터페이스 변경
- P2P 프로토콜 포맷 변경 (Challenge/ChallengeResponse)
- DID 포맷 변경 (did:lango:v2는 Phase 3)
- 새로운 signature algorithm 추가 (Phase 2)
- `internal/p2p/settlement/` 의존 변경 (정당한 wallet 의존)

## Decisions

### D1: Consumer-local interface (not shared package)

Go 관용: interface는 consumer에서 정의. `identity`는 `KeyProvider` (1 method), `handshake`는 `Signer` (2 methods)를 각각 정의. 공유 interface 패키지를 만들지 않음.

**대안:** `internal/security/signer/` 공유 패키지 → 불필요한 추상화 레이어 (no-dead-abstraction-layer 규칙 위반)

### D2: Handshake에서 identity.DIDFromPublicKey 사용

Unit 1 완료 후 `identity` 패키지는 wallet-free. `handshake → identity` import는 안전하며, inline DID 조립 (`"did:lango:" + fmt.Sprintf("%x", pubkey)`) 중복을 제거.

**대안:** `types.DIDPrefix + hex.EncodeToString` → DID 조립 로직 중복, 향후 DID v2에서 동기화 실패 위험

### D3: Provenance verifier는 wiring에서만 주입

`BundleService`는 `verifiers map[string]SignatureVerifyFunc`를 생성자에서 받음. provenance 패키지 내부에 default verifier를 넣지 않음. 빈 map이면 Verify는 "unsupported algorithm" 에러.

**핵심 원칙:** 검증 구현의 소유권은 `app/cli` integration layer. provenance 패키지는 type 정의만.

**대안:** provenance 내부에 default verifier → `p2p/identity` import가 남아서 분리 실패

### D4: ResponseVerifyFunc는 Config에서 optional

`Config.ResponseVerifier`가 nil이면 `VerifySecp256k1Signature` (default). 이 패턴은 이미 `ZKProverFunc`/`ZKVerifierFunc`에서 사용 중.

### D5: BundleSigner interface (callback 대체)

`BundleSignFunc func(ctx, payload) ([]byte, error)`를 `BundleSigner interface { Sign(); Algorithm() }`으로 교체. algorithm을 signer가 자신의 속성으로 제공하므로 하드코딩 제거.

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| NewBundleService 시그니처 변경 (11곳) | 누락 시 컴파일 에러 | 전체 호출부 목록 계획서에 명시 + go build 검증 |
| implicit interface satisfaction 착각 | 런타임 실패 | compile-time check `var _ Signer = (*wallet.LocalWallet)(nil)` 불필요 — wiring에서 바로 컴파일 에러 |
| archtest trailing slash 수정이 기존 테스트를 깨뜨림 | CI 실패 | 모든 p2p 패키지에서 wallet import가 이미 제거된 후 (Wave 4)에 archtest 강화 |
| provenance default verifier 회귀 | 분리 경계 붕괴 | archtest rule로 `provenance → p2p/identity` 금지 강제 |
