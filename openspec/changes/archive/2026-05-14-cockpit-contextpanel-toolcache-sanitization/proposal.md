## Why

The context panel now sanitizes tool names at render time, but the cached `sortedTools` entries still store raw tool names from the metrics snapshot. That leaves a replayable cache model dependent on renderer-level sanitization instead of the cached tool-stats snapshot itself being display-safe.

## What Changes

- Sanitize tool names when building the cached `sortedTools` entries.
- Add regression coverage for malformed cached tool names.
- Record the cached tool-label replay-safety contract in OpenSpec and downstream docs.

## Impact

- Aligns the context panel's tool-stats cache with the same plain-text baseline already enforced at render time.
- Prevents raw control text from persisting inside cached tool-stat labels.
