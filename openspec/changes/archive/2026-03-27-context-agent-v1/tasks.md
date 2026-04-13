## 1. ContextSearchAgent

- [x] 1.1 Create `internal/retrieval/context_agent.go` with ContextSearchAgent, ContextSearchSource interface
- [x] 1.2 Implement `collectionToLayer()` — knowledge + learning only, others return false
- [x] 1.3 Implement `vectorDistanceToScore()` — `max(0, 1.0 - distance)`
- [x] 1.4 Implement `Search()` — call source.Retrieve, filter collections, convert to Finding
- [x] 1.5 Create `internal/retrieval/context_agent_test.go` — mock source, test all scenarios

## 2. Shadow Comparison Enhancement

- [x] 2.1 Add `factualLayers` map to `internal/retrieval/shadow.go`
- [x] 2.2 Enrich `CompareShadowResults` with factual_overlap, factual_old_only, factual_new_only logging

## 3. Wiring

- [x] 3.1 Expand `initRetrievalCoordinator(cfg, kStore, ec)` in `internal/app/wiring_knowledge.go`
- [x] 3.2 Register ContextSearchAgent when ec.ragService available
- [x] 3.3 Update caller in `internal/app/wiring.go` to pass ec

## 4. Verification

- [x] 4.1 `CGO_ENABLED=1 go build -tags fts5 ./...` — full build passes
- [x] 4.2 All retrieval tests pass (context agent + existing)
