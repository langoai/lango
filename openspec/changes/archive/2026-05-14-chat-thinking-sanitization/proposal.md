## Why

Thinking transcript rows already clamp to width and collapse multiline previews, but they still render preview text without stripping terminal control sequences first. Malformed thought summaries can still leak raw escape sequences into the transcript.

## What Changes

- Sanitize thinking transcript preview/fallback text with terminal-control stripping plus single-line normalization.
- Add regression coverage for escaped and multiline thinking text.
- Record the thinking-row sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Extends the plain-text rendering baseline to model thinking summaries.
- Prevents malformed thought content from destabilizing the transcript surface.
