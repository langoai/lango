## Context

Phase 0에서 wallet 의존 제거, provenance verifier injection 도입 완료. 현재 handshake/identity의 서명 알고리즘이 secp256k1+keccak256 하드코딩. challenge signature에 double-hash 버그 존재.

## Goals / Non-Goals

**Goals:**
- 서명 알고리즘을 교체 가능하게 만드는 framework
- Ed25519를 두 번째 알고리즘으로 등록하여 framework 검증
- Challenge double-hash 버그 수정

**Non-Goals:**
- `did:lango` 포맷 확장 (Phase 3)
- ML-DSA/PQC 알고리즘 (Phase 5)
- KEM provider (Phase 4)
- SchemeRegistry singleton (상수 + injection으로 충분)
- Ed25519 production wiring (framework 검증용 테스트만)

## Decisions

### D1: SignatureScheme은 canonical descriptor, registry 아님

`SignatureScheme` struct는 알고리즘 메타데이터(ID, Verify, sizes)를 담는 descriptor. 실제 dispatch는 각 consumer의 injected verifier map이 담당. 공용 registry singleton 없음.

**대안:** SchemeRegistry → consumer 간 등록 상태 drift 위험은 있지만, 현재 consumer가 2개(handshake, provenance)이고 각각 다른 verifier 시그니처를 사용하므로 통합 registry가 오히려 adapter 복잡도를 증가.

### D2: Ed25519 verifier는 wiring closure

`identity.ParseDIDPublicKey(didStr) → security.VerifyEd25519(pubkey, msg, sig)` closure를 app/cli wiring에서 조립. identity 패키지에 Ed25519 verifier를 넣지 않음 — 넣으면 `did:lango:<ed25519-pubkey>` 부분 허용과 같음.

### D3: Signer interface에 Algorithm() 추가

handshake Signer: `SignMessage + PublicKey + Algorithm`. WalletProvider는 implicit satisfaction 깨짐 → `walletHandshakeSigner` wrapper (walletBundleSigner 패턴 동일).

### D4: Challenge/ChallengeResponse에 SignatureAlgorithm 필드

`json:"signatureAlgorithm,omitempty"` — empty = legacy secp256k1 (backward compat). 이전 peer는 이 필드를 보내지 않으므로 empty로 deserialize → secp256k1 기본값.

### D5: ResponseVerifyFunc → SignatureVerifyFunc 이름 변경

challenge와 response 양쪽 검증에 동일 함수 타입 사용. "Response" 접두사는 misleading.

### D6: challengeSignPayload → challengeCanonicalPayload (bug fix)

기존: `Keccak256(canonical)` 반환 → wallet.SignMessage가 다시 hash → double-hash 서명 / single-hash 검증 불일치.
수정: raw canonical bytes 반환, verifier가 hash 담당. signing/verification 동일 깊이.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Signer.Algorithm() 추가 → WalletProvider implicit satisfaction 깨짐 | walletHandshakeSigner wrapper (기존 패턴) |
| Ed25519 verifier가 production에 연결될 위험 | verifier 등록은 test에서만, wiring은 secp256k1 only |
| 상수 이전 (provenance → security) → 호출부 변경 | re-export로 backward compat |
| challenge double-hash fix → old peer 호환 | 기존 challenge 검증이 이미 broken이므로 regression 없음 |
