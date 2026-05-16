## Why

Compact `status` and `approval` transcript rows already collapse multiline text and clamp to width, but they still render content without stripping terminal control sequences first. Malformed runtime messages can still break the plain-text contract of these compact surfaces.

## What Changes

- Sanitize compact status and approval-event row content with ANSI/OSC stripping plus single-line normalization.
- Add regression coverage for escaped runtime messages.
- Record the compact-row sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Extends the plain-text rendering baseline to the remaining compact runtime rows.
- Prevents malformed status/approval messages from destabilizing the transcript surface.
