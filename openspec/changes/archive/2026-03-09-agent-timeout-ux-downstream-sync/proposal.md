## Why

Agent Timeout UX (Phase 1-4) implementation added `AutoExtendTimeout` and `MaxRequestTimeout` config fields, progressive thinking indicators, and structured error events, but downstream artifacts (docs, TUI settings, WebSocket event docs) were not updated. Users cannot discover or configure these features without documentation and TUI support.

## What Changes

- Add `agent.autoExtendTimeout` and `agent.maxRequestTimeout` to README.md config table and docs/configuration.md
- Add 3 new WebSocket events (`agent.progress`, `agent.warning`, `agent.error`) to docs/gateway/websocket.md
- Add progressive thinking indicator to channel features in docs/features/channels.md
- Add TUI form fields for auto-extend timeout and max request timeout in settings
- Add state update handlers for the 2 new config keys

## Capabilities

### New Capabilities

_(none — this is a docs/TUI sync, not a new capability)_

### Modified Capabilities

- `auto-extend-timeout`: Document config fields in README and docs; add TUI settings form fields and state update handlers
- `progress-indicators`: Document progressive thinking in channel features

## Impact

- `README.md` — config table updated
- `docs/configuration.md` — JSON example and config table updated
- `docs/gateway/websocket.md` — 3 new events documented
- `docs/features/channels.md` — channel features list updated
- `internal/cli/settings/forms_impl.go` — 2 new form fields
- `internal/cli/tuicore/state_update.go` — 2 new case handlers
- `internal/cli/settings/forms_impl_test.go` — test updated for new fields
