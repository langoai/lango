## Why

The cockpit Mission Activity buffer currently normalizes whitespace and truncates summaries, but it does not strip ANSI/OSC control sequences before storing them. Mission Control rendering is now hardened, yet the buffered activity model itself can still retain raw control text that later surfaces through replay paths or future consumers.

## What Changes

- Sanitize activity summaries at append time inside the Mission Activity buffer.
- Add regression coverage for malformed buffered and assistant-derived activity summaries.
- Record the sanitized activity-summary replay contract in OpenSpec and downstream docs.

## Impact

- Aligns Mission Control activity storage with the same plain-text baseline already enforced at render time.
- Prevents raw control text from persisting inside cockpit activity snapshots.
