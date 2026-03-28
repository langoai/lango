## 1. Scored Store APIs

- [x] 1.1 Add `ScoredKnowledgeEntry{Entry, Score, SearchSource}` type to `knowledge/types.go`
- [x] 1.2 Add `ScoredLearningEntry{Entry, Score, SearchSource}` type to `knowledge/types.go`
- [x] 1.3 Add `SearchKnowledgeScored()` to `knowledge/store.go` — FTS5 path (Score=-rank) + LIKE path (Score=RelevanceScore) + is_latest filter
- [x] 1.4 Add `SearchLearningsScored()` to `knowledge/store.go` — Score=Confidence, SearchSource="like"

## 2. Retrieval Package — Types + Interfaces

- [x] 2.1 Create `internal/retrieval/finding.go` — Finding struct
- [x] 2.2 Create `internal/retrieval/agent.go` — RetrievalAgent interface + FactSearchSource interface

## 3. FactSearchAgent

- [x] 3.1 Create `internal/retrieval/fact_agent.go` — wraps FactSearchSource, covers 3 factual layers
- [x] 3.2 Create `internal/retrieval/fact_agent_test.go` — mocked FactSearchSource, correct Layer/Score/SearchSource mapping

## 4. RetrievalCoordinator

- [x] 4.1 Create `internal/retrieval/coordinator.go` — parallel agent execution, (Layer,Key) dedup, TruncateFindings, ToRetrievalResult with Score propagation
- [x] 4.2 Create `internal/retrieval/coordinator_test.go` — parallel agents, dedup, merge, truncation, ToRetrievalResult Score test

## 5. Shadow Mode

- [x] 5.1 Create `internal/retrieval/shadow.go` — CompareShadowResults() logging helper

## 6. Config + Integration

- [x] 6.1 Add `RetrievalConfig{Enabled, Shadow}` to `config/types.go`
- [x] 6.2 Add retrieval defaults to `config/loader.go` (enabled=false, shadow=true)
- [x] 6.3 Add `coordinator` field + `WithCoordinator()` to `adk/context_model.go`
- [x] 6.4 Add shadow goroutine in GenerateContent after g.Wait()
- [x] 6.5 Add `initRetrievalCoordinator()` to `app/wiring_knowledge.go`
- [x] 6.6 Wire coordinator into adapter in `app/wiring.go`

## 7. Docs + Verification

- [x] 7.1 Update README.md — Retrieval Coordinator section
- [x] 7.2 Build: `CGO_ENABLED=1 go build -tags fts5 ./internal/retrieval/ ./internal/knowledge/ ./internal/adk/ ./internal/config/` — zero errors
- [x] 7.3 Test: `CGO_ENABLED=1 go test -tags fts5 ./internal/retrieval/ ./internal/knowledge/ ./internal/adk/` — all pass
