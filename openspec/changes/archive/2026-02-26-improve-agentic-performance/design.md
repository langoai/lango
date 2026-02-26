## Context

The Lango ADK agent layer wraps Google ADK runners with Lango-specific features (memory, RAG, learning, orchestration). In production-like usage, three performance/reliability issues surfaced:

1. **Unbounded loops**: Agent.Run() delegates to the ADK runner with no turn cap, so a stuck tool-call chain can spin indefinitely.
2. **Redundant computation**: `EventsAdapter.At(i)` iterates the full event list on every call; `truncatedHistory()` recomputes on every access.
3. **Memory bloat**: Long sessions accumulate reflections without consolidation, and the memory section injected into the system prompt grows unbounded.

Additionally, the learning engine's 0.5 confidence threshold produced false-positive auto-fixes, and the orchestrator's 5-round default was too low for multi-step tasks.

## Goals / Non-Goals

**Goals:**
- Prevent infinite agent loops with a configurable turn limit (default 25)
- Enable self-correction by retrying with learned fixes on tool errors
- Scale token budgets per model family to use context windows efficiently
- Cache computed history/events for O(1) repeated access
- Bound memory section growth with token budgeting and auto meta-reflection
- Reduce false-positive learning auto-applies by raising confidence threshold
- Give the orchestrator more delegation headroom with budget guidance

**Non-Goals:**
- Changing the ADK runner internals or upgrading ADK version
- Persisting token budget configuration (derived at runtime from model name)
- Adding new memory types or changing the reflection/observation data model
- Modifying the learning store schema

## Decisions

### 1. Turn limit via event-stream wrapper (not ADK config)

Wrap the ADK runner's event iterator in `Agent.Run()` and count events with function calls. This avoids depending on ADK-internal turn-limit features which don't exist in the current ADK version.

**Alternative**: Modify the ADK runner to accept a max-turn config. Rejected because it couples us to ADK internals and requires upstream changes.

### 2. Self-correction in RunAndCollect (not Run)

The retry-with-fix logic lives in `RunAndCollect` after the initial run fails and no sub-agent fallback applies. This keeps `Run()` (the iterator) stateless and pure.

**Alternative**: Inject correction into the event stream. Rejected because it would make the iterator stateful and harder to reason about.

### 3. sync.Once for lazy caching in EventsAdapter

Use `sync.Once` for both `truncatedHistory()` and the `At()` method's event list. EventsAdapter is created fresh per session access, so no invalidation is needed.

**Alternative**: Pre-compute on construction. Rejected because not all code paths need both truncated history and converted events.

### 4. Token budget via model name heuristic

`ModelTokenBudget(modelName)` uses string matching on model family names (claude, gemini, gpt-4o, etc.) to return ~50% of the model's context window. Simple, zero-config, and correct for known models.

**Alternative**: Config-file-based budget mapping. Rejected as over-engineering for the current need; the heuristic covers all supported providers.

### 5. Memory token budget with reflection-first priority

Reflections are compressed summaries with higher information density. The assembler includes reflections first, then fills remaining budget with observations. Default budget: 4000 tokens.

### 6. Confidence threshold 0.7 (was 0.5)

0.5 produced false positives from low-confidence early learnings. 0.7 requires more corroboration before auto-applying a fix. The `handleSuccess` boost is also scoped to exact tool triggers to avoid cross-contamination.

### 7. Delegation rounds 10 (was 5) with budget guidance

5 rounds was insufficient for multi-step tasks. 10 provides headroom. The orchestrator prompt now includes round-budget guidance (simple: 1-2, medium: 3-5, complex: 6-10) so the LLM self-manages.

## Risks / Trade-offs

- **Turn limit too low for some tasks** → Configurable via `WithMaxTurns()`, default 25 is generous
- **Self-correction retry doubles latency on failure** → Only triggers when a high-confidence fix exists (>0.7), and only once
- **Model name heuristic misidentifies models** → Falls back to `DefaultTokenBudget` (32K) for unknown models
- **Meta-reflection threshold too aggressive** → Default 5 is conservative; only consolidates, doesn't delete
- **sync.Once prevents mid-session cache invalidation** → EventsAdapter is recreated per access, so staleness isn't possible
