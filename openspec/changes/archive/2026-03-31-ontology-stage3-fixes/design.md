## Context

Stage 3 코드 리뷰 5건 수정. 통합 경로 버그 + 보안 가정 위반 해결.

## Goals / Non-Goals

**Goals:** 5건 모두 수정, 회귀 테스트 추가, 전체 빌드/테스트 통과.

**Non-Goals:** 새 기능 추가 없음.

## Decisions

### D1: TrustScorer interface
`*reputation.Store` 직접 참조 대신 `TrustScorer` interface (GetScore method만) 도입. 테스트 용이성 + 의존성 분리.

### D2: Post-build wiring
Bridge를 `intelligenceValues`에 실어 post-build 단계에서 P2P handler와 연결. Module dependency 추가 없음.

### D3: Registry UpdateStatus methods
PromoteType/PromotePredicate가 create-only RegisterType 대신 update-only UpdateTypeStatus 사용. DeprecateType 패턴 재사용.
