## Why

Currently only the memory section has a token budget (4000 tokens). Knowledge, RAG, and runSummary sections dump all retrieved content into the system prompt with no limit. On smaller models (8k-32k), this can overflow the context window and cause truncation or errors. A budget manager is needed to allocate available context window across all sections proportionally.

## What Changes

- Add `ContextBudgetManager` type with model window registry, response reserve calculation, and ratio-based section allocation
- Add `ContextConfig` to config with `context.modelWindow`, `context.responseReserve`, `context.allocation.*` fields
- Modify `ContextAwareModelAdapter.GenerateContent()` to compute per-section budgets before retrieval and truncate each section to its budget
- Add item-level truncation for knowledge (`TruncateResult` before `AssemblePrompt`), RAG (drop lowest-rank results), memory (dynamic budget), and runSummary (drop older summaries)
- Legacy `memoryTokenBudget=4000` preserved as fallback when budget manager is not set
- Graceful degradation: when `available <= 0`, skip budget enforcement entirely (legacy unbounded mode)

## Capabilities

### New Capabilities
- `context-budget`: Token budget management for context assembly — model window registry, section allocation, per-section truncation, allocation validation, graceful degradation

### Modified Capabilities
- `knowledge-store`: Knowledge search results gain item-level truncation via `TruncateResult` before prompt assembly
- `feature-status`: Knowledge feature status enriched with budget mode indicator (budgeted vs unbounded)

## Impact

- **New files**: `internal/adk/budget.go`, `internal/adk/budget_test.go`
- **Modified packages**: `internal/adk/` (context_model, context_assembly, context_retrieval), `internal/knowledge/` (retriever), `internal/config/` (types, loader), `internal/app/` (wiring)
- **Config**: New `context.*` section (advanced settings). `contextProfile` remains top-level unchanged
- **Breaking**: None — budget manager is additive, legacy behavior preserved when not configured
- **Docs**: README context settings section update. TUI settings deferred to future step
