## Why

Compact and metadata-driven chat surfaces are now heavily sanitized, but assistant markdown content still goes into `renderMarkdown()` raw. If a model emits terminal control sequences, they can survive into the rendered transcript or plain-text fallback path.

## What Changes

- Strip terminal control sequences from assistant markdown input before rendering or fallback display.
- Add regression coverage for sanitized markdown input across normal and fallback paths.
- Record the markdown-input sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Extends the plain visible-text baseline to assistant transcript bodies.
- Prevents malformed model output from destabilizing the transcript surface.
