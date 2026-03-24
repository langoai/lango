## Context

`lango` already starts an interactive TUI chat, but the current interface is still organized like a minimally working transcript viewport with ad hoc banners. The system already has the right runtime pieces: `TurnRunner` for execution, a fallback-based approval provider, markdown rendering, slash commands, and alt-screen entry. The missing part is presentation architecture: transcript items are not typed enough, state transitions are not surfaced clearly, and layout measurement is still easy to desynchronize from rendering.

The redesign must preserve the existing single-column form factor because that best fits the current codebase, terminal constraints, and implementation risk profile. The goal is not to clone `crush`, Claude Code, or Codex component-for-component, but to adopt the same product discipline: clear state hierarchy, visible interruptions, and transcript-first interaction.

## Goals / Non-Goals

**Goals:**
- Make the TUI read like a coding-agent cockpit, not a raw chat log.
- Keep transcript as the primary surface while elevating turn state and approval interruptions.
- Ensure assistant output survives streaming, failure, cancel, and resize paths without layout corruption.
- Keep implementation scoped to `internal/cli/chat/` plus minimal entrypoint and docs changes.

**Non-Goals:**
- No multi-panel or sidebar layout in this change.
- No diagnostics pane, tool timeline panel, or command palette yet.
- No changes to `TurnRunner`, tool middleware contracts, or approval provider interfaces.
- No new slash commands beyond polishing the existing ones.

## Decisions

### 1. Single-column cockpit layout

Use four vertical regions in a single column:
- Header
- Turn status strip
- Transcript viewport
- Composer/help footer

Approval is rendered as an interrupt card in the transcript stack area rather than a separate modal or side panel. This keeps focus on the conversation flow and avoids introducing multi-surface state management.

Alternative considered:
- Multi-panel layout with separate state and approval panes
Why not now:
- Better long-term information density, but too much layout and focus complexity for this phase.

### 2. Typed transcript items

Replace simple role/content entries with typed transcript items carrying:
- `kind`
- rendered `content`
- optional `rawContent`
- optional metadata for status/approval rendering

Assistant entries keep raw markdown for resize reflow. Status and approval events remain compact and intentionally lower-noise than assistant prose.

Alternative considered:
- Keep role/content entries and infer rendering style from state
Why not:
- Too fragile for approval events, status updates, and future tool activity.

### 3. Shared render/layout model

`View()` and `recalcLayout()` will operate on the same fixed-part model. Height calculation is derived from the same blocks that are rendered so that approval entry/exit and terminal resize do not drift.

Alternative considered:
- Maintain manual separator constants and fixed input/banner heights
Why not:
- This has already proven brittle and creates recurring overflow bugs.

### 4. Assistant append unification

All assistant-visible output goes through a single helper that stores raw markdown and computes rendered content for the current content width. This includes:
- successful streaming completion
- non-streaming `ResponseText`
- partial output preserved on failure
- partial output preserved on cancel

Alternative considered:
- Keep separate finalize and fallback append paths
Why not:
- Causes inconsistent dedupe, reflow, and state handling.

## Risks / Trade-offs

- **[Risk] Re-rendering assistant markdown on every resize may be more expensive for long histories** → Accept for now; correctness and stability take priority. Avoid extra caching in this phase and revisit only if profiling shows a real issue.
- **[Risk] More transcript item kinds increase renderer complexity** → Keep the first typed model intentionally small: user, assistant, system, status, approval.
- **[Risk] Approval card placement could still compete with transcript space on short terminals** → Clamp viewport height and keep approval content dense with only critical parameters.

## Migration Plan

1. Introduce typed transcript items and helper APIs without changing CLI entrypoints.
2. Switch `ChatModel` rendering to parts-based cockpit layout.
3. Update tests for transcript rendering, state transitions, resize reflow, and approval layout.
4. Refresh README and CLI docs to describe the cockpit-style TUI rather than a generic interactive chat.
5. Verify with full `go build ./...` and `go test ./...`.

Rollback is straightforward because the change is isolated to TUI presentation files and docs.

## Open Questions

None for this phase. Multi-panel layout, diagnostics surfaces, and richer tool timelines are intentionally deferred.
