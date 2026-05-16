## Why

The fullscreen approval dialog already sanitizes tool names, summaries, origins, and risk metadata, but diff preview lines are still styled from raw `DiffContent`. A file diff containing terminal control sequences can still leak those sequences into the approval surface.

## What Changes

- Strip terminal control sequences from fullscreen approval diff preview lines before styling and caching them.
- Add regression coverage for escaped diff lines.
- Record the diff-preview sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Extends the plain-text rendering baseline to the last rich text surface inside the fullscreen approval dialog.
- Prevents malformed diff content from destabilizing Tier 2 approval UI.
