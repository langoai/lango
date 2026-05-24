## Context

The Phase 1-4 UX roadmap (Elastic Turns, Capability Concierge, Continuity, Extension Packs) was implemented across 4 OpenSpec changes on `feature/elastic-ux`. Codex automated review against `dev` found 14 bugs across 3 review rounds. All fixes have been applied in-tree. This design documents the key technical decisions made during the fix process.

## Goals / Non-Goals

**Goals:**
- Fix all P1/P2 findings from Codex review without introducing new regressions
- Maintain backward compatibility with existing config and behavior
- Keep fixes minimal and scoped — no feature additions or refactoring beyond what's needed

**Non-Goals:**
- Addressing deferred findings (non-context adapter, local source snapshot) — separate change
- Adding new test infrastructure beyond what's needed to verify fixes
- Changing public APIs or config schema

## Decisions

### 1. RAG budget split ratio (1/3 recall, 2/3 RAG)

When both recall and RAG results exist, recall gets 1/3 of `budgets.RAG` and RAG gets the remainder. When only one source exists, it gets the full budget.

**Rationale**: Recall is supplementary context (prior sessions), while RAG is direct semantic retrieval for the current query. 2:1 weighting favors the more relevant source. The split only applies when both sources are present — no change to single-source behavior.

### 2. AllowedExtPacks filter with nil-means-skip-all default

`FileSkillStore.AllowedExtPacks` is a `map[string]bool`. When nil (default), ALL ext-packs are skipped. This required moving `wireExtensionRegistry(app)` before module build so the registry is available when skills are initialized.

**Rationale**: Safe-by-default. If extensions are disabled or the registry fails to load, no extension skills are loaded. An empty non-nil map means extensions are enabled but no packs are healthy.

**Alternative considered**: Passing the full registry to `FileSkillStore` — rejected because it creates a dependency from the skill package to the extension package.

### 3. Stale-stream retry via staleTriggered flag

`wrapChunkCallbackWithStale` returns a third value: `*atomic.Bool` that's set when the stale timer fires. The retry guard allows retry when `chunksEmitted && staleTriggered` (stream stalled after output, recovery appropriate) but blocks retry when `chunksEmitted && !staleTriggered` (genuine mid-stream crash, user saw partial content).

**Rationale**: The two guards (stale detection vs partial-output protection) serve orthogonal purposes. A stale stream means no progress — retry is the right action even if some output was shown. A mid-stream crash means the provider is broken — retry would produce garbled output.

### 4. Live session key read in event closures (not captured local)

Event subscription handlers in `continuity_events.go` read `m.sessionKey` at event time rather than capturing it as a local variable at subscription time. The `defer Store.End()` in `runChat`/`runCockpit` reads `model.SessionKey()` instead of the initial local.

**Rationale**: `/clear` changes the session key mid-session. Captured locals go stale. Live reads ensure the correct key is always used for event filtering and session end processing.

### 5. Token accumulation via turnActive pattern in plain chat

Plain chat subscribes to `TokenUsageEvent` via EventBus and accumulates tokens when `turnActive=true`. Flush happens on `DoneMsg`, reset on `ErrorMsg`.

**Rationale**: Mirrors cockpit's `RuntimeTracker` pattern but without the full tracker infrastructure. `TokenUsageEvent` doesn't populate `SessionKey`, so session-key filtering isn't viable — the `turnActive` window is the correct discriminator.

### 6. Pack mirror copies full skill directories

`copyPackFiles` now uses `copyTree` when the skill path resolves to a directory (or SKILL.md with siblings). This matches `fetchFromDir`'s hash-all-files behavior, preventing false-positive tamper detection on next startup.

**Rationale**: Hash coverage and copy coverage must match exactly. The round-1 fix expanded hash coverage without expanding copy coverage — this completes the pair.

## Risks / Trade-offs

- [Risk] `wireExtensionRegistry` moved earlier → If extension loading has side effects that depend on module build output, ordering could break.
  → Mitigation: `wireExtensionRegistry` only reads config + filesystem, no module dependencies.

- [Risk] Live session key reads in closures are not thread-safe if `/clear` races with event delivery.
  → Mitigation: bubbletea processes events sequentially in the Update loop. `/clear` and event handlers never run concurrently.

- [Risk] `AllowedExtPacks = nil` means existing ext-pack skills silently disappear if the extension registry fails to load.
  → Mitigation: `wireExtensionRegistry` logs a warning on failure. This is the correct behavior — a failed registry should not load potentially tampered packs.
