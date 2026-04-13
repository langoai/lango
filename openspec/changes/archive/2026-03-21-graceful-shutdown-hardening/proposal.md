## Why

`lango serve` can hang during shutdown when one lifecycle-managed component blocks inside its stop path. The current shutdown timeout does not reliably bound the full stop sequence, and repeated `Ctrl+C` presses still leave the process attached to the terminal.

## What Changes

- Make `lango serve` handle shutdown as a two-stage flow: first signal starts graceful shutdown, second signal forces process exit with code `130`.
- Make lifecycle shutdown honor the request deadline per component so one blocked stop handler cannot stall the entire application shutdown.
- Update channel shutdown paths to support context-aware stop behavior and fix Telegram stop ordering so update polling is interrupted before waiting on goroutines.
- Make background and workflow managers support context-aware shutdown instead of waiting indefinitely on internal goroutines.
- Add shutdown observability so logs identify which component is stopping, which completed, and which timed out.

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

- `server`: tighten `lango serve` shutdown behavior so graceful shutdown is deadline-bounded, blocked components do not hang process exit forever, and a second interrupt forces termination.

## Impact

- Affected code: `cmd/lango/main.go`, lifecycle shutdown coordination, app/channel lifecycle registration, Telegram/Slack/Discord channels, background manager, workflow engine.
- Affected behavior: `lango serve` signal handling and shutdown sequencing.
- Affected docs/specs: `openspec/specs/server/spec.md`, README server behavior notes.
