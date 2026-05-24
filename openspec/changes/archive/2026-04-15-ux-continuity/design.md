## Context

Phase 1 gave agents in-turn elasticity (interrupt/redirect, retry, stale-stream detection, inline emergency compaction). Phase 2 gave them per-turn capability shaping (session modes, discovery, cost). Phase 3 pushes that flexibility across turn and session boundaries so users don't silently pay for bloated contexts and don't lose the agent's hard-won knowledge between runs.

Three infrastructures already exist and are the foundation here:

- **`AnalysisBuffer` pattern** (`internal/learning/analysis_buffer.go`) — async, drain-on-signal buffer used today for learning analysis after turn completion. The post-turn hook (`eventbus.TurnCompletedEvent`) fires at `runner.go` and is already subscribed by `wiring_session_usage.go`.
- **`EntStore.CompactMessages(key, upToIndex, summary)`** (`internal/session/ent_store.go:424`) — atomic message compaction that replaces a prefix of messages with a summary entry. Used today only by `ContextAwareModelAdapter` inline emergency compaction.
- **`FTS5Index`** (`internal/search/`) — domain-agnostic virtual-table wrapper with `Insert`, `Update`, `Delete`, `BulkInsert`, `Search` (BM25). Currently used by knowledge; the spec explicitly permits reuse across collections with separate table names.

What Phase 3 adds is the connective tissue: a buffer for async compaction, a session-end trigger, a recall retriever, and eventbus plumbing for learning suggestions. No new storage engines, no schema migrations.

## Goals / Non-Goals

**Goals:**
- Keep session message token footprint below the model window without the user noticing — unless the cost is unavoidable, in which case a 2 s guard keeps the next turn responsive.
- Surface prior-session context when it is relevant, without user config. Retrieval is read-only, opt-out via config.
- Give the learning engine a user-visible voice through the same approval pipe that tools already use, so TUI and channels share the surface.
- Reuse existing infrastructure (AnalysisBuffer, CompactMessages, FTS5, approval, eventbus) — zero new storage, zero new external dependencies.

**Non-Goals:**
- Cross-session *memory* (facts/entities). That is `agent-memory` / `ontology-*` territory and remains unchanged here. Recall here is specifically FTS-based conversation snippets, not a knowledge graph.
- Automatic learning *application* without user consent. All learning suggestions route through approval. The 0.7 auto-apply threshold for `GetFixForError` is untouched.
- Compaction strategy changes. We reuse the existing summarize-and-replace model from inline emergency compaction.
- TUI-only UX. Per the `multichannel-ux-trap` rule, every new user-visible surface ships through the eventbus with at least one non-TUI renderer path (channels already subscribe to approval events).

## Decisions

### 1. Background hygiene compaction uses the `AnalysisBuffer` pattern, not a bespoke worker
**Decision:** Introduce `session.CompactionBuffer` modeled on `learning.AnalysisBuffer`. It accepts `EnqueueCompaction(key string, upToIndex int)` calls from a `TurnCompletedEvent` subscriber, runs a bounded worker pool, and calls `EntStore.CompactMessages` inline.

**Rationale:** `AnalysisBuffer` has already solved the hard problems we'd hit — bounded queue, drop policy on overflow, graceful drain on shutdown, waitable flush for tests. Rebuilding those properties in a new worker is a trap (see `feedback_extend_not_duplicate`). The pattern is also what `CLAUDE.md` means by "extend existing assets."

**Alternatives considered:**
- *Inline at `OnTurnComplete`:* rejected — compaction can take hundreds of ms to seconds; blocks the event-loop user sees.
- *Long-lived goroutine + channel:* rejected — no existing pattern for bounded drop / shutdown drain; see `feedback_no_dead_abstraction_layer`.

### 2. Sync point is a bounded wait, not a fence
**Decision:** `ContextAwareModelAdapter.GenerateContent()` acquires a per-session "compaction in flight" handle and `select`s on `{done, time.After(2s)}`. Timeout → proceed with current (possibly stale, possibly oversized) context. Compaction continues in background; the next turn re-checks.

**Rationale:** Blocking the user's next turn for compaction inverts the UX priority. The 2 s bound matches the plan and is long enough for typical compactions (a few hundred ms on tested transcripts) while short enough that a genuinely slow compaction doesn't feel like a freeze. When inline emergency compaction (Phase 1B-iii) is still needed, that path remains — it runs synchronously because the model request *can't proceed without it*.

**Alternatives considered:**
- *Hard fence:* rejected — hostile to the "elastic turns" promise.
- *No sync point, eventual consistency only:* rejected — races produce visible double-appends of the same summary chunk across turns.

### 3. Trigger threshold is estimated message tokens, not adapter "Degraded"
**Decision:** Trigger = sum of `types.EstimateTokens(msg.Content)` across session messages > `modelWindow * 0.5`. Threshold is configurable under `context.compaction.threshold` (default 0.5). `budgets.Degraded` is explicitly not used as a trigger, consistent with `context-budget` spec and the Phase 1B-iii decision: Degraded signals a configuration problem (base prompt too large), which compacting session messages cannot resolve.

**Rationale:** Degraded-as-trigger would stomp on that contract (`feedback_spec_contract_check_before_refactor`). Session message tokens are also the correct proxy for "session is getting heavy" whereas `Degraded` is the *non*-proxy.

### 4. Session end has two modes — hard and soft — and they behave differently
**Decision:**
- *Hard end* (TUI quit, CLI exit, explicit `session.End(key)`): best-effort synchronous flush, bounded by 3 s drain on the `CompactionBuffer`, then generate the recall summary and FTS-index it.
- *Soft end* (channel idle timeout, adaptive idle): lazy. Marks the session with `session_end_pending=true` in metadata. Next session open for the same principal (or the next scheduled sweep) processes it.

**Rationale:** Hard end has an attentive user waiting, so some latency is acceptable if bounded. Soft end has nobody attending — running heavy work opportunistically on next open avoids wasted cycles on sessions that resume quickly. Metadata flag keeps this off the critical path and survives crashes.

**Alternatives considered:**
- *Always synchronous on end:* rejected — channels can idle-out dozens of sessions in a burst.
- *Always lazy:* rejected — TUI quit followed by reopen in another terminal would lose the recent session's recall index.

### 5. Recall surfaces as a retriever, not a tool
**Decision:** `SessionRecallRetriever` implements the existing `context_retriever` interface. Wired into `ContextAwareModelAdapter` through the same `WithMemory`/`WithRetriever` optional pattern. Mode-aware: respects the active session mode's allowlist of retrievers if one is defined (future extension; default = always on).

**Rationale:** A tool invokes at agent discretion — which means prior-session context only shows up when the model remembers to ask. A retriever injects it at turn start, which is the whole point. The retriever pattern also already handles budget enforcement (`SectionBudgets.RAG` via `context-budget`), so we don't re-solve truncation.

### 6. Learning suggestions route through existing approval pipeline
**Decision:** `learning.SuggestionEngine` publishes `LearningSuggestionEvent` on the eventbus when confidence crosses the suggestion threshold (distinct from the 0.7 auto-apply threshold — suggestion threshold default 0.5). Subscribers: TUI (renders in chat surface), channel adapters (render per-channel). User acceptance calls the existing `approval.Request` flow; on approval, the learning engine persists the rule through its normal path.

**Rationale:** The approval system is already multi-channel. Piggybacking means a channel user who approves a learning suggestion inside Slack gets the same persistence path as a TUI user. No duplicated UX code paths (`multichannel-ux-trap`).

## Risks / Trade-offs

- **[Risk]** Compaction buffer overflow under a burst of rapid turns → **Mitigation:** `AnalysisBuffer`-style drop policy logs a warning; the sync point still waits on the most recent in-flight job. Lost enqueues are re-triggered on the next turn that still exceeds threshold.
- **[Risk]** Sync point 2 s timeout masks a truly slow compactor → **Mitigation:** Emit `CompactionSlowEvent` (warn level) when a job exceeds 2 s; include in debug / status surfaces so the user can see it even though the turn proceeded.
- **[Risk]** FTS recall surfaces stale or irrelevant snippets, wastes context budget → **Mitigation:** BM25 rank floor (configurable, default 0.2) + budget-aware truncation via existing `SectionBudgets.RAG`. User can disable with `context.recall.enabled=false`.
- **[Risk]** Learning suggestion spam when many low-confidence patterns trigger → **Mitigation:** rate-limit (default 1 suggestion per 10 turns) + dedup by pattern hash. Thresholds configurable under `learning.suggestions.*`.
- **[Risk]** Hard-end 3 s drain misses compactions on crash → **Mitigation:** Metadata `session_end_pending=true` flag set *before* drain; next start reprocesses. Same code path as soft end, so one algorithm handles both recovery and the normal lazy path.
- **[Trade-off]** FTS recall is string-level, not semantic. Matches on wording, not meaning. Acceptable for Phase 3; a semantic layer would be Phase 4+ or a separate ontology project.
- **[Trade-off]** Compaction summary quality depends on the model used for summarization. Reusing the inline-emergency path means we inherit its summarizer; if the model is cheap/small, summaries may be thin. Out of scope to replace here.

## Migration Plan

All changes are additive and default-safe.

1. **Compaction**: default `context.compaction.enabled=true` with 0.5 threshold and 2 s sync point — behavior change is visible but opt-outable via `context.compaction.enabled=false`. No data format change.
2. **Recall**: default `context.recall.enabled=true`. New FTS5 table (e.g., `fts_session_recall`) created lazily on first session-end. Existing FTS knowledge index unaffected (separate table per spec §).
3. **Learning suggestions**: default `learning.suggestions.enabled=true`. Gated on existing approval infrastructure; if approval is disabled in config, suggestions are suppressed at the subscriber level.
4. **Rollback**: set all three `enabled=false`. No data cleanup needed. FTS5 recall table is safe to leave (harmless if unused) or drop manually.

## Open Questions

- *Should recall inject at every turn or only when prior-session context exists and matches?* — default **only when matches above rank floor exist**, to avoid empty section overhead. Re-evaluate with telemetry after ship.
- *Should the learning suggestion approval have a "never again for this pattern" option?* — desired UX but punts to Phase 4; for now the user can dismiss without an explicit suppression path.
- *Sync-point strategy when multiple sessions share a process (future channel runtime):* current design is per-session; the buffer is global with per-key handles. This is fine for the bubbletea + channel topology today; revisit if a single process runs many parallel channel sessions under load.
