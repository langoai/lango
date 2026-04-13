# Tasks: temporal-agent-v1

- [x] Add `UpdatedAt time.Time` to `KnowledgeEntry` domain type
- [x] Update 6 conversion sites in `store.go` to populate `UpdatedAt`
- [x] Add `SearchRecentKnowledge` method to `knowledge.Store`
- [x] Create `TemporalSearchSource` interface in retrieval package
- [x] Implement `TemporalSearchAgent` with recency scoring and content enrichment
- [x] Add compile-time interface compliance check
- [x] Write `temporal_agent_test.go` — agent tests (Name, Layers, Search scenarios, Error)
- [x] Write `TestRecencyScore` — score normalization edge cases
- [x] Write `TestEnrichTemporalContent` — age format variations
- [x] Register `TemporalSearchAgent` in `initRetrievalCoordinator`
- [x] Verify build passes (`go build -tags fts5 ./...`)
- [x] Verify all tests pass (`go test -tags fts5 ./internal/retrieval/ ./internal/knowledge/... ./internal/app/`)
- [x] Create OpenSpec delta specs (temporal-agent-v1 NEW, agentic-retrieval MODIFIED)
