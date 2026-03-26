## Why

Knowledge entries are currently mutable — `SaveKnowledge` does upsert (create or update-in-place), destroying previous content on every update. When the agent corrects a preference, refines a rule, or updates a fact, the original version is permanently lost. This prevents auditing what changed, rolling back erroneous updates, and understanding knowledge evolution over time.

This step converts Knowledge to append-only version history: each save creates a new version row, preserving the full edit trail while all reads and search default to the latest version.

## What Changes

- Add `version` (int) and `is_latest` (bool) fields to the Knowledge Ent schema
- **BREAKING**: Remove `key` UNIQUE constraint → `(key, version)` composite unique index
- `SaveKnowledge` becomes append-only: creates new version row instead of update-in-place
- New `GetKnowledgeHistory(key)` method returns all versions ordered by version descending
- `GetKnowledge`, `SearchKnowledge`, `IncrementKnowledgeUseCount` filter `is_latest=true` only
- `DeleteKnowledge` hard-deletes all versions of a key (unchanged semantics)
- FTS5 write-time policy: only latest version indexed; `source_id` remains `key` (not per-version)
- FTS5 bulk index at startup filters `WHERE is_latest = 1`
- `ContentSavedEvent` gains `Version int` field (additive-only, existing subscribers unaffected)
- `KnowledgeEntry` domain type gains `Version int` and `CreatedAt time.Time` fields
- `save_knowledge` tool description and return format updated (includes version number)
- New `get_knowledge_history` tool exposes version browsing to the agent
- Concurrency: retry-once on `(key, version)` unique constraint violation for concurrent same-key saves

## Capabilities

### New Capabilities
- `temporal-knowledge`: Append-only version history for knowledge entries, version tracking, history retrieval, concurrent-safe versioning with retry-on-conflict

### Modified Capabilities
- `knowledge-store`: SaveKnowledge becomes append-only, GetKnowledge/Search/Increment filter is_latest, new GetKnowledgeHistory, KnowledgeEntry gains Version/CreatedAt, schema adds version/is_latest fields
- `knowledge-fts5-integration`: FTS5 bulk index filters is_latest=true, write-time sync unchanged (latest-only by design)

## Impact

- **Schema**: `internal/ent/schema/knowledge.go` — field additions, index restructuring, Ent codegen
- **Store**: `internal/knowledge/store.go` — SaveKnowledge rewrite (tx + retry), read path filters, new GetKnowledgeHistory
- **Types**: `internal/knowledge/types.go` — KnowledgeEntry field additions
- **Events**: `internal/eventbus/events.go` — ContentSavedEvent.Version field
- **Wiring**: `internal/app/wiring_knowledge.go` — bulk index SQL filter
- **Tools**: `internal/app/tools_meta.go` — save_knowledge description/return, new get_knowledge_history tool
- **Docs**: README.md knowledge section
- **No impact on**: embedding/resolver.go, librarian/, learning/, knowledge/retriever.go, adk/ (all use latest-only APIs)
- **Migration**: Ent auto-migration adds columns with defaults (version=1, is_latest=true) to existing rows
- **Scope boundary**: Knowledge only — Learning temporal is a separate follow-up step
