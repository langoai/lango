## 1. FTS5 Search Package (`internal/search/`)

- [x] 1.1 Create `internal/search/probe.go` — `ProbeFTS5(db *sql.DB) bool` function that tests FTS5 availability via temp table create+drop
- [x] 1.2 Create `internal/search/fts5.go` — `FTS5Index` struct with constructor `NewFTS5Index(db *sql.DB, tableName string, columns []string)`
- [x] 1.3 Implement `EnsureTable()` — `CREATE VIRTUAL TABLE IF NOT EXISTS` with `unicode61` tokenizer
- [x] 1.4 Implement `DropTable()` — `DROP TABLE IF EXISTS`
- [x] 1.5 Implement `Insert(ctx, rowid, values []string)`, `Update(ctx, rowid, values []string)`, `Delete(ctx, rowid)`
- [x] 1.6 Implement `BulkInsert(ctx, records []Record)` — single transaction batch insert
- [x] 1.7 Implement `Search(ctx, query string, limit int) ([]SearchResult, error)` — FTS5 MATCH with `rank` ordering, returns `RowID` + `Rank`
- [x] 1.8 Handle phrase queries (quoted), prefix queries (trailing `*`), and plain keywords in Search
- [x] 1.9 Create `internal/search/fts5_test.go` — table-driven tests: probe, CRUD, search (keyword/phrase/prefix), empty query, bulk insert, concurrent access

## 2. Knowledge Store FTS5 Integration

- [x] 2.1 Add `fts5Index *search.FTS5Index` and `learningFTS5Index *search.FTS5Index` fields to `knowledge.Store`
- [x] 2.2 Add `SetFTS5Index(idx *search.FTS5Index)` and `SetLearningFTS5Index(idx *search.FTS5Index)` setter methods
- [x] 2.3 Modify `SearchKnowledge()` — if fts5Index != nil, search FTS5 first, resolve entries by key from Ent, apply category filter; else use existing LIKE path
- [x] 2.4 Modify `SearchLearnings()` — same FTS5-first pattern with learning FTS5 index; LIKE fallback
- [x] 2.5 Add FTS5 error graceful degradation — log warning, fall back to LIKE on FTS5 search error
- [x] 2.6 Add write-time sync in `SaveKnowledge()` — FTS5 insert (new) or update (existing); log warning on failure, don't block Ent write
- [x] 2.7 Add write-time sync in `DeleteKnowledge()` — FTS5 delete
- [x] 2.8 Add write-time sync in `SaveLearning()` — FTS5 insert with learning ID as rowid, columns: trigger/error_pattern/fix
- [x] 2.9 Add write-time sync in `DeleteLearning()` — FTS5 delete by learning ID
- [x] 2.10 Create `internal/knowledge/store_fts5_test.go` — tests for FTS5 search path, LIKE fallback, write-time sync, error degradation

## 3. App Wiring

- [x] 3.1 Add FTS5 probe + index creation in `initKnowledge()` (or a new `initFTS5()` in wiring) — probe FTS5, create knowledge FTS5 table (columns: key, content), create learning FTS5 table (columns: trigger, error_pattern, fix)
- [x] 3.2 Bulk-index existing knowledge and learning entries into FTS5 tables on startup
- [x] 3.3 Inject FTS5 indexes into knowledge store via `SetFTS5Index()` / `SetLearningFTS5Index()`
- [x] 3.4 Track FTS5 availability for health reporting (pass to FeatureStatus or StatusCollector)

## 4. Health & Status Integration

- [x] 4.1 Update context health check (`internal/cli/doctor/checks/context_health.go`) to report FTS5 availability as informational detail
- [x] 4.2 Update status command (`internal/cli/status/status.go`) to show FTS5 status in knowledge feature detail
- [x] 4.3 Test health check with FTS5 available and unavailable scenarios

## 5. Benchmarks & Verification

- [x] 5.1 Create `internal/search/fts5_bench_test.go` — benchmark FTS5 vs LIKE search on 1k and 10k corpus sizes
- [x] 5.2 Run `go build ./...` and verify zero compilation errors
- [x] 5.3 Run `go test ./...` and verify all tests pass
- [x] 5.4 Verify CLI: `lango doctor` shows FTS5 status, `lango status` shows search mode
