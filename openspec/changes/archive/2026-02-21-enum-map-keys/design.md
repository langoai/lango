## Context

The `LearningStats.ByCategory` field uses `map[string]int` despite `Learning.Category` being an Ent-generated enum type (`entlearning.Category`, which is `type Category string`). This requires explicit `string()` casts when populating the map.

## Goals / Non-Goals

**Goals:**
- Replace `map[string]int` with `map[entlearning.Category]int` for `LearningStats.ByCategory`
- Remove `string(e.Category)` cast in `GetLearningStats`
- Maintain JSON serialization compatibility

**Non-Goals:**
- Changing callback metadata maps (`map[string]string`) — these are generic interfaces at system boundaries
- Creating new enum types for `automationAvailable` — only 3 values, would be over-engineering
- Modifying any other map types in the codebase

## Decisions

**Use `entlearning.Category` as map key directly**
- Since `entlearning.Category` is `type Category string`, it is a valid map key type
- `encoding/json` marshals string-based types identically to plain strings, so JSON output is unchanged
- No new types or abstractions needed — reuses existing Ent-generated type

## Risks / Trade-offs

- [Minimal risk] Callers accessing `ByCategory` with string literals will need to use `entlearning.Category("value")` instead — but no external callers exist currently, only internal usage
