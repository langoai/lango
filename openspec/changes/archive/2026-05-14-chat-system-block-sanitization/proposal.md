## Why

System transcript blocks already clamp to width, but their body text still renders raw terminal control sequences. A malformed system/runtime note can leak ANSI/OSC escapes into the transcript even though the rest of the compact chat surfaces are now sanitized.

## What Changes

- Strip terminal control sequences from system transcript block body text before rendering.
- Add regression coverage for escaped system content.
- Record the system-block sanitization contract in OpenSpec.

## Impact

- Extends the plain visible-text baseline to system transcript entries.
- Prevents malformed system notes from destabilizing the transcript surface.
