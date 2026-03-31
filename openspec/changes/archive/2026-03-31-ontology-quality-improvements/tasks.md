## Tasks

### Priority 1: Must Fix

- [x] **1-1** Event-driven predicate cache (`internal/ontology/service.go`)
  - Remove `s.refreshPredicateCache()` from `PredicateValidator()` (:287)
  - Add `ctx context.Context` parameter to `refreshPredicateCache`
  - Update all callers: `RegisterPredicate(:220)`, `DeprecatePredicate(:243)`, `PromotePredicate(:630)`, `ImportSchema(:697)` to pass ctx
  - Add regression test in `ontology_test.go`: verify cache reflects mutations without calling `PredicateValidator()` again

- [x] **1-2** Observable audit log errors (`internal/ontology/action.go`)
  - Replace `_ = e.logStore.Compensated(ctx, logID)` (:132) with slog.Warn on error
  - Replace `_ = e.logStore.Fail(ctx, logID, ...)` (:138) with slog.Warn on error
  - Replace `_ = e.logStore.Complete(ctx, logID, effects)` (:148) with slog.Warn on error

- [x] **1-3** mustMarshal helper (`internal/ontology/exchange.go`)
  - Add `func mustMarshal(v any) []byte` that panics with descriptive message
  - Replace `data, _ := json.Marshal(sorted)` (:125) with `data := mustMarshal(sorted)`

- [x] **P1 gate** Run `go build ./...` and `go test ./...` — all pass

### Priority 2: Should Fix (Performance)

- [x] **2-1** Batch property lookup (`internal/ontology/property_store.go`, `service.go`)
  - Add `GetPropertiesBatch(ctx, entityIDs []string) (map[string]map[string]string, error)` to PropertyStore
  - Single query: `WHERE entity_id IN (...)`
  - Update `QueryEntities` (:511-528) to use batch lookup
  - Add regression test in `property_test.go`

- [x] **2-2** Filter narrowing (`internal/ontology/property_store.go`)
  - After first filter, inject `entityproperty.EntityIDIn(ids...)` into subsequent filter queries (:125-145)
  - Verify `EntityIDIn` exists in generated Ent predicates
  - Add regression test in `property_test.go`

- [x] **2-3** Merge retraction logging (`internal/ontology/resolution.go`)
  - Collect retraction errors in loop (:112-117) instead of discarding
  - After loop: `slog.Warn` with count + first triple's subject/predicate
  - `MergeResult` struct unchanged

- [x] **2-4** Capacity hints
  - `exchange.go:143` — `make([]SchemaTypeSlim, 0, len(types))`
  - `exchange.go:149` — `make([]SchemaPredicateSlim, 0, len(preds))`
  - `property_store.go:186` — `make([]string, 0, len(entries))`

- [x] **P2 gate** Run `go build ./...` and `go test ./...` — all pass

### Priority 3: Nice to Fix (Code Quality)

- [x] **3-1** Local interface for validator injection (`internal/app/wiring_graph.go`)
  - Define `type predicateValidatable interface { SetPredicateValidator(graph.PredicateValidatorFunc) }`
  - Replace `gc.store.(*graph.BoltStore)` (:71) with `gc.store.(predicateValidatable)`

- [x] **3-2** Protocol action constants (`internal/p2p/protocol/ontology_messages.go`)
  - Add `const OntologyActionAccepted/Partial/Rejected`
  - Replace raw strings in message constructors

- [x] **3-3** Import mode constant (`internal/app/wiring_ontology.go`)
  - Replace `"shadow"` (:116) with `string(ontology.ImportShadow)`

- [x] **3-4** Immutability documentation
  - `governance.go:12` — add `// immutable after init — do not modify at runtime`
  - `types.go:36` — add `// immutable after init — do not modify at runtime`

- [x] **P3 gate** Run `go build ./...` and `go test ./...` — all pass

### Downstream Artifact Audit

- [x] Check README.md for ontology API/usage sections — update if affected
- [x] Check prompts/agents/ontologist/IDENTITY.md — update if internal behavior described
- [x] Check docs/ for ontology references — update if affected
- [x] Confirm CLI/TUI unaffected
- [x] Confirm skills unaffected
