## Why

Tool lifecycle rows already sanitize tool names and parameter previews, but their preview/output detail text still only collapses whitespace before truncation. Malformed tool output can still leak terminal control sequences into the transcript.

## What Changes

- Sanitize tool preview and output detail text with terminal-control stripping plus single-line normalization.
- Add regression coverage for escaped preview/output text.
- Record the detail-text sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Completes the plain-text rendering baseline for tool lifecycle rows.
- Prevents malformed tool output from destabilizing compact tool detail lines.
