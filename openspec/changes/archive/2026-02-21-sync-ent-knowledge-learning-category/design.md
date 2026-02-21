## Context

The previous enum type strengthening refactoring introduced `KnowledgeCategory` with 6 constants but only updated the domain type — the Ent schema still had 4 enum values. Additionally, `LearningEntry.Category` remained `string` while the parallel `KnowledgeEntry.Category` was changed to a typed enum. This created an inconsistency where Knowledge and Learning entries followed different type safety patterns.

## Goals / Non-Goals

**Goals:**
- Align Ent Knowledge schema enum with `KnowledgeCategory` domain type (add `pattern`, `correction`).
- Introduce `LearningCategory` typed enum and change `LearningEntry.Category` from `string` to `LearningCategory`.
- Confine `string()` casts to system boundaries: Ent DB writes, metadata maps, tool parameter parsing.
- Maintain backward compatibility — no behavioral changes, no DB migration needed.

**Non-Goals:**
- Changing `SearchLearnings` category parameter to typed enum (it's a query filter at DB boundary, empty string means "no filter").
- Adding `LearningCategory` to config/CLI (Learning categories are internal, not user-configurable).
- Changing `ContextItem.Category` to typed enum (cross-domain container, remains `string`).

## Decisions

**1. Add enum values to Ent schema, not remove from domain type**
`pattern` and `correction` are valid analysis types that could be stored as Knowledge entries. Adding them to Ent makes the schema forward-compatible. Removing from the domain type would lose expressiveness.

**2. LearningCategory constants mirror Ent Learning enum exactly**
The 6 constants (`LearningToolError`, `LearningProviderError`, `LearningUserCorrection`, `LearningTimeout`, `LearningPermission`, `LearningGeneral`) match Ent's Learning category enum 1:1. This prevents the same schema/domain mismatch that prompted this change.

**3. Boundary casts follow established pattern**
Same pattern as `KnowledgeEntry.Category`: `string()` at Ent write + metadata map, `LearningCategory()` at Ent read. `categorizeError()` and `mapLearningCategory()` return typed enum directly since they produce domain values.

## Risks / Trade-offs

- **Risk**: Ent schema change requires `go generate` — **Mitigation**: Adding enum values is backward-compatible, no migration needed. SQLite stores strings directly.
- **Trade-off**: `SearchLearnings` category parameter stays `string` while `LearningEntry.Category` is typed — acceptable because it's a query filter at the DB interface boundary, and empty string idiom is cleaner than a zero-value typed enum.
