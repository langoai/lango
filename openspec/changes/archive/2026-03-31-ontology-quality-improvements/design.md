## Context

The ontology subsystem was built across 10 commits. A systematic code review (reuse, quality, efficiency) identified 10 actionable items spanning bug-class issues, performance regressions, and code hygiene. All fixes preserve existing API contracts and test behavior.

## Goals / Non-Goals

### Goals
- Fix hot-path DB query in `PredicateValidator()`
- Add observability for silent error paths (ActionExecutor, Merge retraction)
- Reduce N+1 query patterns in QueryEntities and PropertyStore.Query
- Replace concrete type assertions and raw strings with proper abstractions

### Non-Goals
- Change `OntologyService` interface signature (reason parameter stays per governance spec)
- Persist governance daily counts (v1 in-memory is acceptable)
- Refactor permission check boilerplate (explicit pattern is readable)

## Approach

### Priority 1: Bug-class fixes

**1-1. Event-driven predicate cache** (`service.go:284-311`)
- Remove `refreshPredicateCache()` call from `PredicateValidator()` — return cached closure only
- Add `ctx context.Context` parameter to `refreshPredicateCache(ctx)`
- Existing callers (`RegisterPredicate:220`, `DeprecatePredicate:243`, `PromotePredicate:630`, `ImportSchema:697`) already call refresh — just pass `ctx` through
- `NewService` initializes cache once at construction via `context.Background()` (acceptable for init)

**1-2. Observable audit log errors** (`action.go:132,138,148`)
- Replace `_ = e.logStore.XXX(...)` with `if err := ...; err != nil { slog.Warn("action log write", "op", "complete/fail/compensate", "logID", logID, "error", err) }`
- Use `log/slog` (standard library, already available in the project)

**1-3. mustMarshal helper** (`exchange.go:125`)
- Add package-private `mustMarshal(v any) []byte` that panics on error
- Replace `data, _ := json.Marshal(sorted)` with `data := mustMarshal(sorted)`
- Signature of `ComputeDigest` unchanged — no caller impact

### Priority 2: Performance fixes

**2-1. Batch property lookup** (`service.go:511-528`)
- Add `GetPropertiesBatch(ctx, entityIDs) (map[string]map[string]string, error)` to `PropertyStore`
- Single Ent query: `WHERE entity_id IN (...)` → group results by entity_id
- `QueryEntities` loop: replace per-entity `GetProperties()` with batch result map lookup
- Graph store queries remain sequential (no batch API in Store interface)

**2-2. Filter narrowing** (`property_store.go:122-166`)
- After first filter produces entity IDs, inject `WHERE entity_id IN (prevResult)` into subsequent filters
- Ent: `query.Where(entityproperty.EntityIDIn(ids...))` — built-in predicate
- Early return on empty result already exists (:163)

**2-3. Merge retraction logging** (`resolution.go:112-117`)
- Collect retraction errors in a slice during merge loop
- After loop: if any errors, `slog.Warn("merge retraction partial failure", "count", len(errs), "first_subject", errs[0].Subject, "first_predicate", errs[0].Predicate)`
- `MergeResult` struct unchanged

**2-4. Capacity hints**
- `exchange.go:143`: `slimTypes := make([]SchemaTypeSlim, 0, len(types))`
- `exchange.go:149`: `slimPreds := make([]SchemaPredicateSlim, 0, len(preds))`
- `property_store.go:186`: `var result []string` → `result := make([]string, 0, len(entries))`

### Priority 3: Code quality

**3-1. Local interface for predicate validator injection** (`wiring_graph.go:71-74`)
- Define in `wiring_graph.go`:
  ```go
  type predicateValidatable interface {
      SetPredicateValidator(graph.PredicateValidatorFunc)
  }
  ```
- Replace `gc.store.(*graph.BoltStore)` with `gc.store.(predicateValidatable)`

**3-2. Protocol action constants** (`ontology_messages.go`)
- Add `const OntologyActionAccepted/Partial/Rejected` and use in message constructors

**3-3. Import mode constant** (`wiring_ontology.go:115`)
- Replace `"shadow"` with `string(ontology.ImportShadow)`

**3-4. Immutability documentation** (`governance.go:12`, `types.go:36`)
- Add `// immutable after init — do not modify at runtime` comment to `validTransitions` and `SourcePrecedence`

## Risks

| Risk | Mitigation |
|------|------------|
| Cache staleness after removing eager refresh | All mutation paths already call refresh; regression test verifies |
| Batch query returns different order | QueryEntities already sorts by propStore.Query result order |
| `entityproperty.EntityIDIn` not available in Ent codegen | Verify generated predicates include `In` variant; fall back to manual `Where` |
