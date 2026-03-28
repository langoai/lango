## Context

Step 2 (`knowledge-search-fts-baseline`) added FTS5 search. Step 3 adds token budget management across all prompt sections. Currently `ContextAwareModelAdapter.GenerateContent()` assembles 4 sections in parallel (knowledge, RAG, memory, runSummary) and concatenates them into the system instruction with no global budget — only memory has a 4000-token budget.

Key code paths:
- `internal/adk/context_model.go:141-240` — GenerateContent orchestrates parallel retrieval and assembly
- `internal/adk/context_assembly.go:17-84` — assembleMemorySection with token budget
- `internal/adk/context_retrieval.go` — assembleRAGSection/assembleGraphRAGSection, no budgets
- `internal/knowledge/retriever.go` — ContextRetriever.Retrieve() + AssemblePrompt()
- `internal/types/token.go:15-30` — EstimateTokens() character-based approximation
- `internal/config/types.go:141` — Agent.MaxTokens (output tokens, 4096 default)
- `provider.ModelInfo.ContextWindow` — available but requires API call

## Goals / Non-Goals

**Goals:**
- Allocate available context window across knowledge, RAG, memory, runSummary sections with configurable ratios
- Enforce per-section token budgets via item-level truncation (before text assembly, not after)
- Provide static model window registry (no API calls needed)
- Preserve legacy behavior when budget manager is not set
- Graceful degradation when available budget is zero or negative

**Non-Goals:**
- Dynamic budget reallocation based on actual retrieval results (Step 12)
- Tiktoken-accurate token counting (current character-based approximation is sufficient)
- TUI settings for budget config (deferred)
- Per-query adaptive budgets

## Decisions

### D1: Static model window registry, not runtime API call

Model window sizes are looked up from a static map keyed by model name prefix. No `Provider.ListModels()` call at budget computation time.

**Rationale:** Budget computation happens in the hot path (`GenerateContent`). API calls would add latency and failure modes. Model windows rarely change.

**Alternative considered:** Query provider at startup — adds initialization complexity and requires provider to be ready before budget manager.

### D2: Item-level truncation BEFORE assembly, not post-assembly text cutting

Each section's content is truncated at the item level (dropping lower-priority items) before being formatted into text. Never cut assembled text strings.

**Rationale:** Post-assembly truncation breaks markdown headings, layer boundaries (`## User Knowledge`, `### Summary`), and can produce malformed prompts. Item-level truncation preserves structural integrity.

### D3: ContextConfig in `types.go`, not `types_knowledge.go`

The new `ContextConfig` struct lives in `internal/config/types.go` alongside other top-level configs.

**Rationale:** Budget management is a cross-cutting concern spanning knowledge, RAG, memory, and runSummary. It's not specific to knowledge.

### D4: `contextProfile` remains top-level, `context.*` is advanced settings

`contextProfile` (from Step 1) stays at the config root. Budget allocation is under `context.*`. Most users only set `contextProfile`, power users tune `context.allocation.*`.

**Rationale:** Preserves backward compatibility and keeps the common case simple.

### D5: Allocation sum must equal 1.0, validated at construction

The sum of all allocation ratios (knowledge + rag + memory + runSummary + headroom) must be exactly 1.0. Validation at `ContextBudgetManager` construction time; returns error if invalid.

**Rationale:** Prevents silent misconfiguration where sections either overflow or waste budget.

### D6: Degradation on `available <= 0` → skip budgets entirely

When `modelWindow - responseReserve - basePromptTokens <= 0`, the budget manager returns unlimited budgets for all sections, effectively preserving legacy behavior.

**Rationale:** A model with a tiny window or a very large base prompt shouldn't crash. The system degrades to existing unbounded behavior with a warning log.

## Risks / Trade-offs

- **[Character-based token estimation is approximate]** → Accepted. Exact tiktoken is expensive. ~20% margin is tolerable since headroom allocation (10%) absorbs estimation error.
- **[Static model registry may miss new models]** → Mitigated by configurable `context.modelWindow` override and generous fallback default (128k).
- **[Truncation may drop relevant content]** → Acceptable trade-off. Dropping low-priority items is better than overflowing the context window. Step 12 (dynamic reallocation) will optimize this later.

## Migration Plan

- **Deploy:** No migration needed. New config fields have sensible defaults. Budget manager is created automatically when `context.modelWindow` > 0 or model is in the registry.
- **Rollback:** Remove `context.*` from config. Budget manager won't be created; legacy unbounded behavior resumes.
- **Legacy:** Existing `observationalMemory.memoryTokenBudget` is honored as fallback when budget manager is not set.
