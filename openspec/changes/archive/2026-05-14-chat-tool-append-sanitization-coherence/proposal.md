## Why

Tool lifecycle rows sanitize tool names and detail text during rendering, but `appendToolStart()` still stores the raw tool name in the transcript entry model. That leaves tool transcript storage out of sync with the already-hardened display contract and makes future render-path reuse easier to get wrong.

## What Changes

- Sanitize stored tool names at append time for tool-lifecycle transcript entries.
- Add regression coverage for stored tool-entry names.
- Record the append-time coherence contract in OpenSpec.

## Impact

- Aligns tool transcript storage with the existing plain-text rendering baseline.
- Reduces the risk of future raw-name regressions in alternate transcript consumers.
