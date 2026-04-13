## Why

The TUI still leaks raw terminal control responses into the composer and transcript, which makes the interface feel broken and untrustworthy. At the same time, user and assistant messages are still too visually similar, so the transcript does not read like a polished coding-agent interface.

## What Changes

- Eliminate OSC 11 terminal background-color response leakage by removing markdown auto style detection from the TUI renderer.
- Extend the existing terminal-response input guard so idle composer input is protected from both CPR and OSC response sequences.
- Improve transcript readability with stronger but still restrained visual separation between user, assistant, status, and approval blocks.

## Capabilities

### New Capabilities

### Modified Capabilities
- `tui-chat-rendering`: Update markdown style selection and transcript block presentation for clearer message separation.
- `tui-cpr-filter`: Expand the composer input guard to consume OSC terminal response sequences in addition to CPR.

## Impact

- Affected code is concentrated in `internal/cli/chat/`.
- Public CLI behavior stays the same: `lango` still opens the TUI and `lango serve` still runs the full server.
- README and CLI docs should be updated only where they describe the TUI visual experience.
