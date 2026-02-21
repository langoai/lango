## Why

The Ent Knowledge schema's `category` enum only has 4 values (`rule`, `definition`, `preference`, `fact`) while the domain type `KnowledgeCategory` has 6 (`pattern`, `correction` added). This mismatch means saving a knowledge entry with `CategoryPattern` or `CategoryCorrection` would fail at runtime with an Ent validation error. Additionally, `LearningEntry.Category` is a plain `string`, inconsistent with the typed enum pattern already applied to `KnowledgeEntry.Category`, `Message.Role`, and `Confidence`.

## What Changes

- Add `pattern` and `correction` to the Ent Knowledge schema enum to match `knowledge.KnowledgeCategory` constants.
- Introduce `knowledge.LearningCategory` typed enum with constants matching the Ent Learning schema (`LearningToolError`, `LearningProviderError`, `LearningUserCorrection`, `LearningTimeout`, `LearningPermission`, `LearningGeneral`).
- Change `LearningEntry.Category` from `string` to `LearningCategory`, applying the same boundary-cast pattern (string cast at DB/metadata boundaries, typed enum internally).
- Update `categorizeError()` and `mapLearningCategory()` return types from `string` to `LearningCategory`.

## Capabilities

### New Capabilities

### Modified Capabilities
- `knowledge-store`: `KnowledgeEntry` Ent schema gains `pattern`/`correction` enum values; `LearningEntry.Category` changes from `string` to `LearningCategory` typed enum.

## Impact

- **ent/schema/knowledge.go**: Add 2 enum values, regenerate Ent code
- **knowledge package**: `types.go` (new `LearningCategory` type), `store.go` (DB boundary casts), `retriever.go` (ContextItem boundary)
- **learning package**: `parse.go` (`mapLearningCategory` return type), `analyzer.go` (`categorizeError` return type), `engine.go`/`session_learner.go`/`conversation_analyzer.go` (typed constants)
- **app package**: `tools.go` (boundary cast for `save_learning` tool handler)
- **test files**: String literal categories replaced with typed constants
