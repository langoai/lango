## Context

Agent Timeout UX (Phase 1-4) added `AutoExtendTimeout` and `MaxRequestTimeout` config fields to `internal/config/types.go`, progressive thinking indicators to channels, and structured error events to WebSocket gateway. These features are fully implemented in core but lack documentation and TUI settings exposure.

## Goals / Non-Goals

**Goals:**
- Sync all downstream artifacts (README, docs, TUI) with the already-implemented core changes
- Ensure users can discover and configure auto-extend timeout via `lango settings`
- Document new WebSocket events for gateway API consumers

**Non-Goals:**
- No changes to core logic or agent runtime behavior
- No new CLI commands
- No changes to default config values

## Decisions

1. **TUI field placement**: Add `auto_extend_timeout` and `max_request_timeout` fields directly after `tool_timeout` in the Agent form, keeping timeout-related fields grouped together.

2. **MaxRequestTimeout display**: Show `0s` when unset (Go zero value for `time.Duration`). The placeholder text explains the 3× default behavior.

3. **WebSocket event docs**: Document `agent.progress`, `agent.warning`, and `agent.error` in the existing events table rather than creating a separate section, maintaining the flat event list pattern.

## Risks / Trade-offs

- [Risk] `MaxRequestTimeout.String()` shows "0s" when unset → Acceptable; placeholder text clarifies default behavior, consistent with other duration fields like `toolTimeout`.
