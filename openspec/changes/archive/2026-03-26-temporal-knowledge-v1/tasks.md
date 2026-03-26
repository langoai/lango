## 1. Schema + Codegen

- [x] 1.1 Add `version` (int, default 1) and `is_latest` (bool, default true) fields to Knowledge Ent schema
- [x] 1.2 Remove `key` UNIQUE constraint from Knowledge schema
- [x] 1.3 Add composite unique index on `(key, version)` and non-unique index on `(key, is_latest)`
- [x] 1.4 Run `go generate ./internal/ent/...` to regenerate Ent code

## 2. Domain Type + Event

- [x] 2.1 Add `Version int` and `CreatedAt time.Time` fields to `KnowledgeEntry` in `types.go`
- [x] 2.2 Add `Version int` field to `ContentSavedEvent` in `eventbus/events.go`
- [x] 2.3 Update `publishContentSaved` internal method to accept `version int` param and map to event field
- [x] 2.4 Update all `publishContentSaved` call sites (SaveLearning passes 0)

## 3. Store CRUD — SaveKnowledge

- [x] 3.1 Extract core save logic into `saveKnowledgeOnce` internal method
- [x] 3.2 Implement first-version path: create with `version=1`, `is_latest=true`
- [x] 3.3 Implement append-version path: tx{set old `is_latest=false`, create new version with carry-forward use_count/relevance_score}
- [x] 3.4 Add `isUniqueConstraintError` helper function
- [x] 3.5 Implement retry-once wrapper in public `SaveKnowledge` method

## 4. Store CRUD — Read Operations

- [x] 4.1 Update `GetKnowledge` to filter `is_latest=true`, populate Version and CreatedAt in returned entry
- [x] 4.2 Add `GetKnowledgeHistory(ctx, key)` method — all versions ordered by version desc
- [x] 4.3 Update `searchKnowledgeLIKE` to add `is_latest=true` predicate
- [x] 4.4 Update `resolveKnowledgeByKeys` to add `is_latest=true` predicate (defense-in-depth)
- [x] 4.5 Update `IncrementKnowledgeUseCount` to filter `is_latest=true`

## 5. FTS5 Bulk Index

- [x] 5.1 Update `bulkIndexKnowledge` SQL query in `wiring_knowledge.go` to filter `WHERE is_latest = 1`

## 6. Downstream Artifacts

- [x] 6.1 Update `save_knowledge` tool description in `tools_meta.go` to indicate version append semantics
- [x] 6.2 Update `save_knowledge` tool return to include `version` field and updated message format
- [x] 6.3 Add `get_knowledge_history` tool in `tools_meta.go` — accepts key, returns all versions desc
- [x] 6.4 Update README.md Knowledge section to note version history behavior

## 7. Tests — Existing Updates

- [x] 7.1 Update `TestSaveAndGetKnowledge/upsert` to verify append-version behavior (2 rows, latest returned)
- [x] 7.2 Update `TestDeleteKnowledge` to verify all versions deleted
- [x] 7.3 Update `TestIncrementKnowledgeUseCount` to verify latest-only increment

## 8. Tests — New Temporal

- [x] 8.1 Add `TestSaveKnowledge_VersionHistory` — 3 saves, verify versions 1/2/3, GetKnowledge returns v3
- [x] 8.2 Add `TestGetKnowledgeHistory` — 3 saves, verify descending order, all versions present with CreatedAt
- [x] 8.3 Add `TestGetKnowledgeHistory_NotFound` — returns ErrKnowledgeNotFound
- [x] 8.4 Add `TestSaveKnowledge_CarryForward` — increment use_count on v1, save v2, verify v2 carries use_count
- [x] 8.5 Add `TestSearchKnowledge_LatestOnly` — save v1 "old", save v2 "new", search "old" returns nothing

## 9. Tests — FTS5 Temporal

- [x] 9.1 Update `TestWriteTimeSync_Knowledge` — verify FTS5 contains only latest content after version update
- [x] 9.2 Add `TestFTS5_OnlyLatestVersion` — 3 saves, FTS5 MATCH finds only v3 content

## 10. Build Verification

- [x] 10.1 Run `go generate ./internal/ent/...` — zero codegen errors
- [x] 10.2 Run `CGO_ENABLED=1 go build -tags fts5 ./...` — zero build errors
- [x] 10.3 Run `CGO_ENABLED=1 go test -tags fts5 ./internal/knowledge/ ./internal/app/ ./internal/eventbus/` — all tests pass
