# Tasks: retrieval-aggregation-v2

- [x] Add Source, Tags, Version, UpdatedAt fields to Finding struct
- [x] Update FactSearchAgent to populate provenance from ScoredKnowledgeEntry.Entry
- [x] Update TemporalSearchAgent to populate provenance from KnowledgeEntry
- [x] Add sourceAuthority ranking map to coordinator
- [x] Implement compareFindingPriority function (authority → version → recency → score)
- [x] Replace dedupFindings with mergeFindings using evidence-based priority
- [x] Fix save_knowledge default source from "" to "knowledge"
- [x] Add TestMergeFindings_Authority — 6 evidence merge scenarios
- [x] Add TestCompareFindingPriority — 8 priority chain edge cases
- [x] Verify existing dedup tests still pass (backward compatible)
- [x] Verify build passes (`go build -tags fts5 ./...`)
- [x] Verify all tests pass (`go test -tags fts5 ./internal/retrieval/ ./internal/app/`)
- [x] Create OpenSpec delta specs (retrieval-aggregation-v2 NEW, agentic-retrieval MODIFIED)
