## Why

The ontology subsystem (`internal/ontology/`) was added in 10 commits across the `feature/lango-ontology` branch. Code review identified bug-class issues (unconditional DB queries on hot path, silent error ignoring), N+1 query patterns, and stringly-typed code that should be fixed before merging to `dev`.

## What Changes

- **Fix**: `PredicateValidator()` unconditionally queries DB on every call — switch to event-driven cache refresh
- **Fix**: `ActionExecutor` silently ignores audit log write failures — add observability logging
- **Fix**: `ComputeDigest` ignores `json.Marshal` error — add `mustMarshal` panic helper
- **Perf**: `QueryEntities` makes 2N DB queries for N entities — add batch property lookup
- **Perf**: `PropertyStore.Query` makes one DB query per filter — narrow scan with entity ID IN clause
- **Perf**: `Merge` silently ignores retraction errors — add warning logs with triple context
- **Perf**: Missing slice/map capacity hints in exchange and property store
- **Quality**: `wiring_graph.go` type-asserts to `*BoltStore` — use local `predicateValidatable` interface
- **Quality**: Raw string literals for protocol actions and import modes — use typed constants
- **Quality**: Mutable global maps lack immutability documentation

## Capabilities

### New Capabilities

_(none — this is an internal quality improvement with no new capabilities)_

### Modified Capabilities

_(no spec-level behavior changes — all fixes preserve existing contracts)_

## Impact

- **Code**: `internal/ontology/` (service, action, exchange, property_store, resolution), `internal/app/` (wiring_graph, wiring_ontology), `internal/p2p/protocol/` (ontology_messages)
- **Tests**: `ontology_test.go`, `property_test.go`, `tools_test.go` — regression tests for cache semantics and batch behavior
- **APIs**: No public API changes. `MergeResult` struct unchanged. `OntologyService` interface unchanged.
- **Dependencies**: None
