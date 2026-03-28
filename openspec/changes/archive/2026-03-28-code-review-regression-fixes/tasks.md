# Tasks: Code Review Regression Fixes

## Fix 1: DefaultConfig() defaults restore
- [x] 1.1 Add RunLedger defaults to DefaultConfig() (values from dev branch)
- [x] 1.2 Add Provenance defaults to DefaultConfig() (values from dev branch)
- [x] 1.3 Add Sandbox defaults to DefaultConfig() (values from dev branch)
- [x] 1.4 Verify `go test ./internal/cli/doctor/checks/...` passes
- [x] 1.5 Verify `go test ./internal/cli/settings/...` passes

## Fix 2: Settings explicitKeys
- [x] 2.1 Export `ContextRelatedKeys()` in `config/auto_enable.go` (returns slice copy)
- [x] 2.2 Update `settings.go` Save call to pass all context-related keys as explicit
- [x] 2.3 Remove TODO(step8) comment

## Fix 3: RAG enabled flag
- [x] 3.1 Gate RAGService creation on `emb.RAG.Enabled` in `wiring_embedding.go`
- [x] 3.2 Add `cfg.Embedding.RAG.Enabled` check to ContextSearchAgent registration in `wiring_knowledge.go`

## Fix 4: Relevance score clamping
- [x] 4.1 Implement two-step BoostRelevanceScore (add + cap) with error handling
- [x] 4.2 Implement two-step DecayAllRelevanceScores (subtract + floor) with error handling
- [x] 4.3 Verify `go test ./internal/knowledge/...` passes

## Verification
- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go build -tags "fts5,vec" ./...` passes
- [x] 5.3 All affected package tests pass
