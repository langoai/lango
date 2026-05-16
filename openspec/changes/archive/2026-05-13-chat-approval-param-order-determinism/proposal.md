## Why

The chat approval banner and fullscreen dialog iterate over map-backed request params directly. That makes the visible parameter order non-deterministic across renders, which weakens UI reproducibility and makes approval output harder to compare in tests and operator screenshots.

## What Changes

- Render approval request params in stable key order in the fallback banner and fullscreen dialog.
- Add regressions for deterministic param ordering.
- Record the stable param-order contract in OpenSpec.

## Impact

- Improves approval UI consistency without changing semantics.
- Makes approval surfaces easier to test and reason about.
