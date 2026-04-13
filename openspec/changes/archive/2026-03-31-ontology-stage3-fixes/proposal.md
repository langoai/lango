## Why

5 issues found in Stage 3 code review: P2P wiring missing (Critical), governance promotion failure (Major), ImportSchema cache not updated (Major), trust threshold ineffective (Major), event publish not implemented (Minor). Build/test passes but features do not work in integration paths.

## What Changes

- Add `SetReputation(TrustScorer)`, `SetEventBus(*eventbus.Bus)` setters to Bridge, implement event publish
- Introduce `TrustScorer` interface (removes direct reputation.Store dependency)
- Include bridge in `intelligenceValues`, connect `handler.SetOntologyHandler(bridge)` in post-build wiring
- Add `UpdateTypeStatus`/`UpdatePredicateStatus` to Registry, fix PromoteType/PromotePredicate
- Add `refreshPredicateCache()` + `version.Add()` after ImportSchema
- Add 6 regression tests

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
- `internal/app/modules.go` — bridge field in intelligenceValues
- `internal/app/app.go` — post-build wiring
