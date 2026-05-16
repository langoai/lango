## Why

Delegation transcript rows already clamp to width and sanitize the delegation reason, but the `from` and `to` agent names still render without stripping terminal control sequences first. Malformed agent names can still leak raw escape text into the transcript.

## What Changes

- Sanitize delegation actor names with control-sequence stripping plus single-line normalization.
- Add regression coverage for escaped and multiline delegation names.
- Record the delegation-name sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Extends the plain-text rendering baseline to the remaining unsanitized delegation metadata.
- Prevents malformed agent names from destabilizing delegation transcript rows.
