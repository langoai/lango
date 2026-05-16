## Why

The submit-to-first-event pending indicator is still rendered without any width parameter, so it does not follow the same narrow-terminal safety baseline as the other compact chat transcript/status surfaces.

## What Changes

- Make the pending indicator width-aware and clamp it on narrow terminals.
- Add regressions for narrow and zero-width rendering.
- Record the contract in OpenSpec and downstream cockpit docs.

## Impact

- Keeps the pending indicator aligned with the rest of the chat surface hardening.
- Prevents pre-stream waiting UI from overflowing on narrow terminals.
