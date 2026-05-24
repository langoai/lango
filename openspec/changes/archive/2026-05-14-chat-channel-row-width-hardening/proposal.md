## Why

Channel transcript rows already sanitize ANSI escapes, but they still only flatten literal newlines and do not clamp the final rendered row. Long sender names, badge-heavy prefixes, or whitespace-heavy remote messages can still push the row past the available width.

## What Changes

- Normalize channel transcript text and sender names to single-line whitespace.
- Clamp the final channel transcript row to the available width.
- Add regressions for narrow-width and multiline channel payloads.
- Record the contract in OpenSpec and downstream cockpit docs.

## Impact

- Makes remote channel transcript rows behave like the other hardened compact transcript surfaces.
- Prevents external chat noise from destabilizing the local cockpit transcript layout.
