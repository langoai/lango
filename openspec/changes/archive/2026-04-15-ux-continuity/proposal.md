## Why

Phase 3 of the zero-config UX roadmap. Phase 1 (elastic turns) and Phase 2 (capability concierge) gave agents flexibility within a single session, but context and learnings still evaporate between turns and sessions. Users re-explain prior decisions, watch context windows fill silently, and carry "I wish the agent remembered that" friction across runs. The infrastructure — `AnalysisBuffer`, `CompactMessages`, `FTS5Index`, `OnTurnComplete`, `eventbus`, approval plumbing — already exists. What's missing is the connective UX that turns it into continuity the user can feel.

## What Changes

- **Background Hygiene Compaction** — after every turn, estimate current session message tokens; if above 50% of the model window, enqueue an async compaction job. A sync-point guard (max 2s) at the next turn's `GenerateContent()` waits for in-flight compaction without blocking the user; on timeout the turn proceeds and compaction continues in the background. Publishes `CompactionCompletedEvent` on the eventbus.
- **Session Recall (FTS)** — wire session-end triggers (hard end = TUI quit/CLI exit with best-effort 3s drain; soft end = channel idle timeout, lazy processing on next start) to generate a short session summary and index it in the existing FTS5 infrastructure. Add a retrieval hook that surfaces matching prior-session snippets at turn start.
- **Learning Suggestions (approval-gated, multi-channel)** — when the learning engine proposes a new rule/skill/preference, publish `LearningSuggestionEvent` through the eventbus. TUI and channel adapters each render their own surface; acceptance goes through the existing approval path before persistence. No TUI-only code paths (per multichannel-ux-trap rule).

## Capabilities

### New Capabilities
- `background-hygiene-compaction`: post-turn token-threshold-triggered async compaction with a bounded sync point.
- `session-recall`: session-end summarization + FTS-indexed recall hook at turn start.
- `learning-suggestions`: eventbus-delivered approval-gated learning proposals shared across TUI and channels.

### Modified Capabilities
- `interactive-tui-chat`: render `LearningSuggestionEvent` and `CompactionCompletedEvent` in the chat surface.
- `eventbus`: add `CompactionCompletedEvent` and `LearningSuggestionEvent` types.
- `session-store`: add session-end trigger hook used by recall indexing.

## Impact

- **Code**: `internal/session/` (compaction_buffer.go, recall hook, session-end trigger), `internal/app/app.go` (OnTurnComplete wiring, sync point guard), `internal/eventbus/` (two new event types), `internal/cli/chat/` (renderers for the new events), `internal/learning/` (emit learning suggestions over eventbus), `internal/toolcatalog/dispatcher.go` (session-end trigger handoff for soft-end flow).
- **APIs**: No external API changes. Internal `session.Store` gains an `OnSessionEnd` hook; `ContextAwareModelAdapter.GenerateContent()` adds an internal sync-point await before compaction-sensitive phases.
- **Dependencies**: Reuses existing FTS5 index, `AnalysisBuffer`, `CompactMessages`, approval pipeline. No new third-party dependencies.
- **Config**: Additive fields under `context.compaction.*` (threshold ratio, sync timeout) and `learning.suggestions.*` (enabled flag). Defaults match the plan (50% / 2s / enabled).
