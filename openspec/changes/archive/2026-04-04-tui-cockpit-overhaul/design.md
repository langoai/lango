## Context

The Lango TUI is built on Bubble Tea/Lipgloss. The cockpit shell (`internal/cli/cockpit/`) already provides a 3-panel layout (sidebar, main, context panel), 5 pages, and a theme system. The chat model (`internal/cli/chat/`) has a structured transcript (`[]transcriptItem` with user/assistant/system/status/approval kinds) and a state machine (idle/streaming/approving/cancelling/failed). The turnrunner exposes 4 callbacks (`OnChunk`, `OnWarning`, `OnDelegation`, `OnBudgetWarning`) but internal ADK events (tool lifecycle, thinking) are consumed only by the trace recorder.

The spike confirmed that `genai.Part.Thought` (bool) is available in ADK v1.0.0. The background task manager exists but is not wired to the cockpit in local-chat mode.

## Goals / Non-Goals

**Goals:**
- Surface tool execution, thinking, and approval status as distinct visual items in the transcript
- Separate low-risk and high-risk approval requests into two distinct UI tiers
- Make background task activity visible without leaving the chat view
- Transform the footer into an operational status display
- Deliver via 4 sequential implementation waves with hub file ownership rules

**Non-Goals:**
- Channel-to-TUI real-time integration (gateway not started in local-chat mode)
- ChannelEnvelope abstraction (no consumer yet)
- Session switching between main and background task transcripts (complexity vs value)
- Tool progress reporting (ADK tools execute synchronously, no progress producer)
- Replacing the cockpit's chat-wrapper architecture (ChatModel stays as transcript owner)

## Decisions

### D1. No separate UIEvent package — bridge lives in chat package

The only consumer of runtime events is `ChatModel.Update()`. Creating a shared `uievent` package would be a dead abstraction with no second consumer. Instead, `internal/cli/chat/bridge.go` defines `enrichRequest(program, req)` that sets turnrunner callbacks to send chat-specific `tea.Msg` types.

**Alternative considered:** Shared `internal/cli/uievent/` package with typed event hierarchy. Rejected because it adds an indirection layer that no other package imports. If a second consumer appears (e.g., cockpit context panel subscribing to tool events), extract at that time.

### D2. turnrunner.Request gets new callbacks, not a single OnEvent

Adding `OnToolCall`, `OnToolResult`, and `OnThinking` as separate typed callbacks matches the existing pattern (`OnChunk`, `OnWarning`, `OnDelegation`, `OnBudgetWarning`). Each callback has a distinct signature tuned to its payload.

**Alternative considered:** Single `OnEvent func(UIEvent)` callback. Rejected because it forces the caller to type-switch and loses the typed-payload benefit. The existing Request already uses per-event callbacks.

**Duration calculation:** Runner maintains a local `map[string]time.Time` keyed by callID. `OnToolCall` records `startedAt`, `OnToolResult` computes `time.Since(startedAt)`. Map entries are cleaned up on result or turn completion.

### D3. Thinking detection via genai.Part.Thought in recordEvent()

The `recordEvent()` function iterates `event.Content.Parts`. We add a `part.Thought` check: on the first `part.Thought == true` boundary, fire `OnThinking(agentName, true, part.Text)` and start accumulating thought text in a `strings.Builder`. Subsequent thought chunks accumulate without re-firing start. When the next non-thought part arrives, fire `OnThinking(agentName, false, accumulatedSummary)` with the full accumulated text and reset. The bridge tracks `thinkingStart` to compute duration for `ThinkingFinishedMsg`.

A `PendingIndicatorTickMsg` covers the submit-to-first-event gap for responses that don't start with thinking or tool calls. The pending indicator is dismissed (with layout recalculation) on the first chunk, tool, thinking, approval, done, or error event.

### D4. Renderer stub pattern for approval surfaces

Wave 3 creates `approval_strip.go` and `approval_dialog.go` as stubs that delegate to `renderApprovalBanner()`. Wave 3 also plants the tier dispatch in `renderApproval()` and key dispatch stubs (`handleApprovalDialogKey`, `scrollApprovalDialog`) as no-ops. Wave 4 replaces the stub files entirely with real implementations. This ensures hub files (chat.go, approval.go) are only modified in Wave 3.

**Alternative considered:** Function variable / renderer registry. Rejected as over-engineering for a sequential wave delivery where the stub file replacement is simpler.

### D5. Approval tier classification: SafetyLevel + ToolCapability combination

`ClassifyTier(safetyLevel, category, activity)` returns Fullscreen when `safetyLevel == "dangerous"` AND (`category ∈ {"filesystem", "automation"}` OR `activity ∈ {"execute", "write"}`). Everything else is Inline. This catches exec, fs_write, fs_edit, browser actions, and bg_submit while keeping read-only tools on the compact strip.

### D6. Task strip at ChatModel level, Tasks page at cockpit level

The task strip (1-2 line summary) is part of `ChatModel.View()` so it appears in both `lango` (cockpit) and `lango chat` (standalone). The full Tasks page is a cockpit `Page` registered at `PageTasks` with Ctrl+5. BackgroundManager is passed via `Deps` — nil in minimal configurations.

### D7. Hub file ownership: Wave 3 exclusive

`chat.go`, `chatview.go`, `messages.go`, `statusbar.go`, `approval.go` are modified ONLY in Wave 3. Wave 4 adds new files and modifies cockpit/wiring files. This eliminates merge conflicts between waves. The stub pattern (D4) enables this by pre-planting all dispatch points in Wave 3.

## Risks / Trade-offs

**[R1] Thinking detection reliability** — `genai.Part.Thought` depends on the provider model actually emitting thought parts. If a provider doesn't use extended thinking, the thinking transcript item simply won't appear. The pending indicator covers the gap. → Mitigation: pending indicator as universal fallback.

**[R2] Fullscreen approval dialog focus management** — Modal overlay in Bubble Tea must capture all key routing. The existing `stateApproving` guard already gates all input. → Mitigation: Tier 2 dialog reuses the same state machine, just renders differently.

**[R3] Diff generation for large files** — Reading a file to generate a diff during approval could be slow. → Mitigation: Cap diff at 500 lines with truncation marker.

**[R4] BackgroundManager nil in some configurations** — Local-chat mode may not have background tasks enabled. → Mitigation: All task surface code checks for nil manager and renders empty/hidden.

**[R5] Palette change touches many consumers** — Semantic alias approach minimizes this: existing constants remain, aliases point to them. → Mitigation: No constants deleted, only aliases added. Single wave for easy revert.
