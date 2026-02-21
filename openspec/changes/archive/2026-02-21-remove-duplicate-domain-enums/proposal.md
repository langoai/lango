# Remove Duplicate Domain Enums: Use Ent-Generated Types Directly

## Problem

`knowledge.KnowledgeCategory` and `knowledge.LearningCategory` were 100% duplicates of `ent/knowledge.Category` and `ent/learning.Category`. This caused:
- Double maintenance: every enum value change required updating two locations
- Unnecessary boundary casts between domain and ORM layers
- No additional type safety (the values were identical)

## Solution

Remove the domain-layer enum types and use the Ent-generated types (`entknowledge.Category`, `entlearning.Category`) directly in `KnowledgeEntry.Category` and `LearningEntry.Category` fields.

## Scope

- `internal/knowledge/types.go` — remove `KnowledgeCategory`/`LearningCategory` types, use Ent types in struct fields
- `internal/knowledge/store.go` — remove boundary casts
- `internal/learning/` — update function return types and constant references
- `internal/librarian/` — update function return types and constant references
- `internal/app/tools.go` — update boundary casts to use Ent types directly
- Test files — update all constant references
