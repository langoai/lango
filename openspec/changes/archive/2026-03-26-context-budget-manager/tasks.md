## 1. Budget Core (`internal/adk/budget.go`)

- [x] 1.1 Create `internal/adk/budget.go` — `SectionAllocation` struct with Knowledge, RAG, Memory, RunSummary, Headroom float64 fields
- [x] 1.2 Implement `ContextBudgetManager` struct with modelWindow, responseReserve, basePromptTokens, allocation fields
- [x] 1.3 Implement `NewContextBudgetManager(modelWindow, responseReserve, basePromptTokens int, alloc SectionAllocation) (*ContextBudgetManager, error)` with allocation sum validation (must equal 1.0 within 0.001 tolerance)
- [x] 1.4 Implement `SectionBudgets() SectionBudgets` that computes per-section token budgets (available * ratio). Return 0 for all when available <= 0 (degradation)
- [x] 1.5 Implement `LookupModelWindow(modelName string) int` — static model registry with prefix matching. Known models: gemini-2.0-flash (1M), gemini-1.5-pro (2M), claude-sonnet/opus/haiku (200k), gpt-4o (128k), gpt-4o-mini (128k). Default fallback 128k
- [x] 1.6 Implement `DefaultAllocation() SectionAllocation` returning 0.30/0.25/0.25/0.10/0.10
- [x] 1.7 Implement response reserve clamping: min 1024, max 25% of model window

## 2. Config (`internal/config/`)

- [x] 2.1 Add `ContextConfig` struct to `internal/config/types.go` with ModelWindow int, ResponseReserve int, Allocation (knowledge/rag/memory/runSummary/headroom float64)
- [x] 2.2 Add `Context ContextConfig` field to main `Config` struct in types.go
- [x] 2.3 Add default values in `internal/config/loader.go` — allocation 0.30/0.25/0.25/0.10/0.10, modelWindow 0, responseReserve 0
- [x] 2.4 Verify config loads correctly with `go test ./internal/config/`

## 3. Knowledge Truncation (`internal/knowledge/retriever.go`)

- [x] 3.1 Add `TruncateResult(result *RetrievalResult, budgetTokens int) *RetrievalResult` to retriever.go — item-level truncation before AssemblePrompt
- [x] 3.2 TruncateResult: return unchanged when budgetTokens == 0 (unlimited) or result fits
- [x] 3.3 TruncateResult: remove lowest-priority items from each layer when over budget, using types.EstimateTokens
- [x] 3.4 Test TruncateResult: within budget, over budget, zero budget, empty result

## 4. Assembly Integration (`internal/adk/`)

- [x] 4.1 Add `budgetManager *ContextBudgetManager` field and `WithBudgetManager()` builder method to ContextAwareModelAdapter
- [x] 4.2 In GenerateContent: if budgetManager is set, compute SectionBudgets before parallel retrieval
- [x] 4.3 Pass knowledge budget to knowledge retrieval path — call TruncateResult on RetrievalResult before AssemblePrompt
- [x] 4.4 Pass RAG budget to assembleRAGSection/assembleGraphRAGSection — drop lowest-rank results until within budget (using types.EstimateTokens on resolved content)
- [x] 4.5 Pass memory budget to assembleMemorySection — replace static memoryTokenBudget with dynamic budget when budget manager is set
- [x] 4.6 Pass runSummary budget to assembleRunSummarySection — drop older summaries until within budget
- [x] 4.7 When budget manager is not set, preserve existing unbounded behavior (no truncation)

## 5. App Wiring (`internal/app/wiring.go`)

- [x] 5.1 In wiring.go context adapter setup: create ContextBudgetManager using LookupModelWindow(cfg.Agent.Model), cfg.Agent.MaxTokens as reserve, cfg.Context allocation
- [x] 5.2 If cfg.Context.ModelWindow > 0, use config value instead of registry lookup
- [x] 5.3 Estimate basePromptTokens from builder.Build() base prompt
- [x] 5.4 Inject via ctxAdapter.WithBudgetManager(bm); handle construction errors gracefully (log warning, continue without budget)
- [x] 5.5 Enrich knowledge FeatureStatus with budget mode info

## 6. Tests (`internal/adk/budget_test.go`)

- [x] 6.1 Test NewContextBudgetManager: valid allocation, invalid sum, edge cases
- [x] 6.2 Test SectionBudgets: 8k model, 32k model, 128k model, 200k model
- [x] 6.3 Test SectionBudgets: degradation when available <= 0
- [x] 6.4 Test LookupModelWindow: known models, prefix matching, unknown fallback
- [x] 6.5 Test response reserve clamping: min 1024, max 25% of window
- [x] 6.6 Test legacy compatibility: no budget manager set → unbounded behavior

## 7. Docs

- [x] 7.1 Update README.md context configuration section with new `context.*` settings
- [x] 7.2 Document default allocation ratios and model window auto-detection

## 8. Build Verification

- [x] 8.1 Run `go build -tags fts5 ./...` — zero compilation errors
- [x] 8.2 Run `go test -tags fts5 ./internal/adk/ ./internal/config/ ./internal/knowledge/ ./internal/app/` — all tests pass
