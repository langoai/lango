## Why

Stage 3 코드 리뷰에서 5건 발견: P2P 배선 누락(Critical), 거버넌스 승격 실패(Major), ImportSchema 캐시 미갱신(Major), trust threshold 무효(Major), event publish 미구현(Minor). 빌드/테스트 통과하지만 통합 경로에서 기능 미작동.

## What Changes

- Bridge에 `SetReputation(TrustScorer)`, `SetEventBus(*eventbus.Bus)` setter 추가, event publish 구현
- `TrustScorer` interface 도입 (reputation.Store 직접 의존 제거)
- `intelligenceValues`에 bridge 포함, post-build wiring에서 `handler.SetOntologyHandler(bridge)` 연결
- Registry에 `UpdateTypeStatus`/`UpdatePredicateStatus` 추가, PromoteType/PromotePredicate 수정
- ImportSchema 후 `refreshPredicateCache()` + `version.Add()` 추가
- 6개 회귀 테스트 추가

## Capabilities

### Modified Capabilities
- `ontology-exchange-bridge`: TrustScorer interface, SetReputation/SetEventBus setter, event publish, post-build wiring
- `ontology-governance`: UpdateTypeStatus/UpdatePredicateStatus registry methods, PromoteType/PromotePredicate fix
- `ontology-schema-codec`: ImportSchema cache refresh + version bump

## Impact

- `internal/p2p/ontologybridge/bridge.go` — TrustScorer interface, setters, event publish
- `internal/ontology/registry.go` — +2 interface methods
- `internal/ontology/registry_ent.go` — +2 implementations
- `internal/ontology/service.go` — PromoteType/PromotePredicate fix, ImportSchema cache fix
- `internal/app/modules.go` — intelligenceValues에 bridge 필드
- `internal/app/app.go` — post-build wiring
