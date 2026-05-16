## Why

The recovery transcript row already clamps to width and collapses multiline metadata, but `causeClass` still renders without stripping terminal control sequences first. Malformed orchestration metadata can still leak raw escape text into the compact recovery row.

## What Changes

- Sanitize recovery `causeClass` text with terminal-control stripping plus single-line normalization.
- Add regression coverage for escaped and multiline recovery metadata.
- Record the recovery-row sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Extends the plain-text rendering baseline to the remaining unsanitized recovery metadata.
- Prevents malformed orchestration metadata from destabilizing compact recovery rows.
