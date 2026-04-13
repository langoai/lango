## Why

The current interactive TUI chat is functional but still feels like a fragile debug surface instead of a production coding-agent interface. Transcript rendering, turn-state visibility, and approval interruptions all compete for space, which makes the experience feel substantially behind terminal-native agent products.

## What Changes

- Redesign the interactive TUI chat as a single-column coding-agent cockpit centered on transcript, turn state, and approval flow.
- Replace role-only transcript rendering with typed transcript items and dedicated renderers for user, assistant, system, status, and approval events.
- Promote turn progress and approval handling to first-class UI elements while keeping tool activity as secondary status information.
- Align layout measurement and rendering so resize, streaming, and approval interruptions remain stable on narrow terminals.

## Capabilities

### New Capabilities
- `tui-cockpit-layout`: Single-column cockpit layout for transcript, state strip, approval card, and composer.

### Modified Capabilities
- `tui-chat-rendering`: Upgrade transcript rendering, assistant markdown reflow, and state/event presentation in TUI mode.
- `tui-cpr-filter`: Preserve CPR filtering while keeping keyboard input, approval interactions, and resize behavior stable in the redesigned cockpit.

## Impact

- Affected code is concentrated in `internal/cli/chat/` and the TUI entrypoint in `cmd/lango/main.go`.
- Public CLI behavior remains `lango` for interactive TUI and `lango serve` for the full runtime; no command surface changes are introduced.
- README and CLI documentation must be updated to reflect the improved cockpit-style TUI experience.
