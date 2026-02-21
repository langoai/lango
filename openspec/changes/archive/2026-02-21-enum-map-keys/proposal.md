## Why

The previous refactoring changed `KnowledgeEntry.Category` and `LearningEntry.Category` fields to Ent-generated enum types, but map keys/values still use `string` with explicit `string(enum)` casts. Using the enum type directly as map keys eliminates unnecessary casts and strengthens type safety.

## What Changes

- Change `LearningStats.ByCategory` from `map[string]int` to `map[entlearning.Category]int`
- Remove `string(e.Category)` cast when populating the map
- Update `make()` call to use the typed map

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `knowledge-store`: `LearningStats.ByCategory` field type changes from `map[string]int` to `map[entlearning.Category]int` for stronger type safety

## Impact

- **Code**: `internal/knowledge/store.go` — `LearningStats` struct and `GetLearningStats` method
- **API compatibility**: JSON serialization remains identical since `entlearning.Category` is `type Category string`
- **Dependencies**: No new dependencies; uses existing `entlearning` import
