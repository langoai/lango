## Context

`SuggestionEmitter` (`learning/suggestion.go`) publishes `LearningSuggestionEvent` when a pattern crosses the confidence threshold. It has rate-limiting (per-session turn counter), dedup (pattern hash within sliding window), and dismissal tracking.

`learning/engine.go` has `OnToolResult(ctx, sessionKey, toolName, params, result, err)` called after every tool execution. It accumulates error patterns and learns fixes. The engine calls `SuggestionEmitter.MaybeEmit` when a fix candidate is ready.

The drift detection is conceptually different from suggestions: a suggestion proposes a new rule, while a drift signal says "existing spec no longer matches observed behavior." The signal path is the same (EventBus) but the event type and threshold logic differ.

## Goals / Non-Goals

**Goals:**
- Publish `SpecDriftDetectedEvent` when recurring error patterns suggest spec staleness
- Reuse `SuggestionEmitter`'s dedup/rate-limit infrastructure
- Event-only — no file writes to `openspec/`

**Non-Goals:**
- Automatic OpenSpec draft generation (explicitly forbidden by plan)
- Modifying existing `LearningSuggestionEvent` flow
- UI rendering of drift events (future work — TUI/channel adapters can subscribe later)

## Decisions

### D1: Drift detection heuristic

**Choice**: Track error pattern frequency per tool across sessions. When the same `(toolName, errorClass)` pair recurs ≥ N times (const, default 5) within the dedup window, emit `SpecDriftDetectedEvent`. The `errorClass` is derived from the existing `CauseClassifier` in the learning engine.

**Why frequency-based?** A one-off error is noise. Recurring errors across sessions indicate systematic divergence between spec expectations and runtime reality.

### D2: Extend SuggestionEmitter (not new type)

**Choice**: Add `EmitSpecDrift(ctx, toolName, errorClass, sampleErr)` to `SuggestionEmitter`. Reuse its dedup map (pattern hash = `drift-{toolName}-{errorClass}`), rate limit, and bus. Separate from `MaybeEmit` because the threshold and semantics differ.

### D3: New event type in continuity_events.go

**Choice**: `SpecDriftDetectedEvent` with toolName, errorClass, occurrences, sampleError, and affectedSpec (best-guess spec name derived from tool category/name — empty if unknown).
