## Context

Knowledge entries are stored as single mutable rows: `SaveKnowledge` upserts by unique `key`, destroying previous content. The Knowledge Ent schema has `key UNIQUE`, and all CRUD methods assume a single row per key. FTS5 integration (Step 2) uses `source_id = key` for indexing. The context budget manager (Step 3) operates on retrieval results, not storage.

This design converts Knowledge to append-only version history while preserving all existing API signatures and caller compatibility.

## Goals / Non-Goals

**Goals:**
- Append-only versioning: each SaveKnowledge creates a new row instead of mutating
- All reads (Get, Search, Increment) default to latest version only
- New GetKnowledgeHistory API for version browsing
- FTS5 latest-only indexing via write-time policy (source_id = key, unchanged)
- Concurrent save safety via retry-on-conflict
- Agent-facing tool contract updated (save_knowledge returns version, new get_knowledge_history tool)

**Non-Goals:**
- Learning temporal versioning (separate follow-up step)
- Soft-delete / version restoration UI
- Version diffing or merge semantics
- Partial unique index on (key, is_latest) — Ent doesn't support SQLite partial unique
- Schema migration scripts — Ent auto-migration handles field additions

## Decisions

### D1. Append-only via transaction + retry

**Choice**: SaveKnowledge uses `client.Tx()` to atomically set old `is_latest=false` and create new version row.

**Alternative considered**: Two separate operations without transaction. Rejected because a crash between the two operations would leave multiple `is_latest=true` rows for the same key, violating the singleton invariant.

**Retry**: On `(key, version)` unique constraint violation (concurrent same-key write), retry once with fresh read. SQLite single-writer serialization makes single retry sufficient.

### D2. is_latest singleton invariant — application-level enforcement

**Choice**: For any key, at most one row has `is_latest=true`. Enforced by the transaction in SaveKnowledge (set old false → create new true), not by a DB constraint.

**Alternative considered**: SQLite partial unique index (`CREATE UNIQUE INDEX ... WHERE is_latest = 1`). Rejected because Ent ORM doesn't support partial indexes natively, and raw SQL migration adds maintenance burden for a constraint that the transaction already guarantees.

### D3. Hard-delete all versions

**Choice**: `DeleteKnowledge(key)` deletes all version rows. Matches existing behavior.

**Alternative considered**: Soft-delete (set all is_latest=false). Rejected — no current use case for version recovery, and soft-delete requires additional "is_deleted" logic in every read path.

### D4. Carry forward use_count and relevance_score

**Choice**: New version inherits `use_count` and `relevance_score` from the previous latest version. These represent cumulative key-level metrics, not per-version metrics.

**Alternative considered**: Reset to 0 / 1.0. Rejected because use_count represents total key usage across its lifetime, and resetting would break relevance ranking.

### D5. FTS5 source_id = key (not per-version)

**Choice**: FTS5 `source_id` remains `key`. Write-time policy ensures only latest content is indexed: version=1 → Insert, version>1 → Update (delete+re-insert same source_id).

**Alternative considered**: Use `key:version` as source_id. Rejected — would require cleaning up old version entries and complicate the key-based resolution in `resolveKnowledgeByKeys`.

### D6. ContentSavedEvent.Version as typed field only

**Choice**: Add `Version int` to `ContentSavedEvent` struct. Do NOT pass version via `Metadata["version"]`.

**Rationale**: Typed field is discoverable, type-safe, and doesn't pollute the metadata map which serves a different purpose (category, etc.).

### D7. Tool contract update

**Choice**: Update `save_knowledge` description and return format to include version number. Add new `get_knowledge_history` tool.

**Rationale**: The semantic change from "overwrite" to "append version" is visible to the agent. The return message should confirm which version was created, and version browsing should be available.

## Risks / Trade-offs

**[Risk] SQLite table recreation on schema migration** → Ent auto-migration may recreate the knowledge table when removing the key UNIQUE constraint. Mitigation: Knowledge row counts are typically small (tens to hundreds). The migration is a one-time cost at startup.

**[Risk] is_latest invariant not DB-enforced** → A bug in SaveKnowledge could leave multiple is_latest=true rows. Mitigation: All write paths go through SaveKnowledge (single code path). GetKnowledge uses `Only()` which will error if multiple rows match, surfacing violations immediately.

**[Risk] Concurrent save conflict** → Two goroutines saving the same key simultaneously. Mitigation: Retry-once on unique constraint violation. Single retry is sufficient due to SQLite single-writer serialization.

**[Trade-off] Storage growth** → Append-only means old versions accumulate. Acceptable for knowledge entries (low volume, high value). Future steps may add compaction/archival if needed.

**[Trade-off] No version restoration** → Hard-delete removes all history. Acceptable — version restoration is not a current requirement.
