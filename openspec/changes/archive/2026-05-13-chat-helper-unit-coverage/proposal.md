## Why

Recent chat-surface hardening introduced several pure helpers (`sortedParamKeys`, `formatParamValue`, `singleLineValue`, `compactRequestID`), but most coverage still reaches them indirectly through renderer tests. Direct unit coverage is missing for the helper contracts themselves.

## What Changes

- Add direct unit tests for the recent chat helper functions.
- Record the helper-coverage expectation in OpenSpec.

## Impact

- Improves test granularity around critical formatting helpers.
- Makes future regressions easier to localize when deterministic rendering changes.
