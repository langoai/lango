# Tasks: dynamic-budget-reallocation

- [x] Add `SectionTokens` type to budget.go
- [x] Implement `ReallocateBudgets` method with empty-section redistribution
- [x] Add 7 reallocation test cases (all present, one empty, two empty, all empty, degraded, proportional, no recursive)
- [x] Split `assembleMemorySection` into `retrieveMemoryData` + `formatMemorySection`
- [x] Split `assembleRunSummarySection` into `retrieveRunSummaryData` + `formatRunSummarySection`
- [x] Split `assembleRAGSection` into `retrieveRAGData` + `formatRAGSection`
- [x] Split `assembleGraphRAGSection` into `retrieveGraphRAGData` + `formatGraphRAGSection`
- [x] Add token estimation helpers (estimateKnowledgeTokens, estimateRAGResultTokens, estimateMemoryTokens, estimateRunSummaryTokens)
- [x] Restructure GenerateContent to two-phase flow (retrieve → reallocate → assemble)
- [x] Verify build passes (`go build -tags fts5 ./...`)
- [x] Verify all tests pass (`go test -tags fts5 ./internal/adk/... ./internal/app/`)
- [x] Create OpenSpec delta specs
