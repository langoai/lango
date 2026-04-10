## Why

`internal/p2p/handshake/`와 `internal/p2p/identity/`가 `internal/wallet`을 직접 import하여 settlement용 `WalletProvider`(5 methods)에 종속되어 있다. 실제로는 `PublicKey()`와 `SignMessage()` 2개만 사용. `internal/provenance/`도 `internal/p2p/identity`를 직접 import하여 `VerifyMessageSignature`에 하드코딩 의존. 이 결합은 Phase 2(Algorithm Agility)와 Phase 3(Hybrid DID v2) 진행을 막는 구조적 장벽이다.

## What Changes

- `internal/p2p/identity/`: `wallet.WalletProvider` 의존을 1-method `KeyProvider` interface로 교체. wallet import 제거.
- `internal/p2p/handshake/`: `wallet.WalletProvider` 의존을 2-method `Signer` interface로 교체. inline DID 조립을 `identity.DIDFromPublicKey`로 교체. wallet import 제거.
- `internal/p2p/handshake/`: response verification 로직을 injectable `ResponseVerifyFunc`로 추출하여 Phase 2 algorithm agility 준비.
- `internal/provenance/`: `BundleSignFunc` callback을 `BundleSigner` interface로 교체. signature algorithm을 signer에서 제공. verifier map을 wiring에서 주입. `internal/p2p/identity` import 제거.
- `internal/p2p/identity/signature.go`: `string()` 비교를 `bytes.Equal`로 수정 (bug fix).
- `internal/archtest/`: `forbiddenForP2P` trailing slash 버그 수정. `p2p/identity`를 networking prefixes에 추가. provenance → p2p/identity 금지 규칙 추가.

## Capabilities

### New Capabilities

(없음 — 이 change는 기존 capability의 내부 boundary 정리)

### Modified Capabilities

- `p2p-identity`: `WalletDIDProvider`가 `wallet.WalletProvider` 대신 consumer-local `KeyProvider` interface를 받도록 변경. `VerifyMessageSignature`의 byte 비교 bug fix.
- `p2p-handshake`: `Handshaker`가 `wallet.WalletProvider` 대신 consumer-local `Signer` interface를 받도록 변경. inline DID 조립을 `identity.DIDFromPublicKey`로 교체. response verification을 injectable `ResponseVerifyFunc`로 추출.
- `session-provenance`: `BundleService.Export`가 `BundleSignFunc` 대신 `BundleSigner` interface를 받도록 변경. `BundleService.Verify`가 주입된 verifier map을 사용. provenance 패키지에서 `p2p/identity` import 제거 (의존 역전).
- `architecture-boundary-enforcement`: archtest에 p2p/identity 포함, provenance → p2p/identity 금지 규칙 추가, trailing slash 매칭 버그 수정.

## Impact

- **코드:** `internal/p2p/identity/`, `internal/p2p/handshake/`, `internal/provenance/`, `internal/app/`, `internal/cli/provenance/`, `internal/archtest/`
- **API:** `NewBundleService`, `Export`, `provenanceSigner` 시그니처 변경 (internal only, 외부 API 무영향)
- **프로토콜:** P2P handshake Challenge/ChallengeResponse 구조 불변. DID 포맷 불변.
- **의존성:** 새 외부 의존성 없음.
- **하위 호환성:** `wallet.WalletProvider`는 `KeyProvider`와 `Signer`를 implicit satisfy하므로 wiring 코드는 타입 변환 없이 동작.
