## Why

Channel transcript rendering already sanitizes channel names, sender names, and message text, but `appendChannel()` still stores the raw message body in `rawContent`. That leaves the transcript model out of sync with the already-hardened display contract for channel rows.

## What Changes

- Sanitize stored channel message text at append time for transcript entries.
- Add regression coverage for stored sanitized channel payloads.
- Record the append-time coherence contract in OpenSpec.

## Impact

- Aligns channel transcript storage with the existing plain-text rendering baseline.
- Reduces the chance of future raw-message regressions if alternate render paths reuse stored channel data.
