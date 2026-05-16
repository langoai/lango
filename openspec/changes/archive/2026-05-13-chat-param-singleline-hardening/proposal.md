## Why

Chat approval surfaces and compact tool previews now render params deterministically, but raw string values can still carry embedded newlines. That lets one param break a supposedly compact single-line preview or inject extra lines into approval layouts.

## What Changes

- Normalize rendered param values into single-line-safe display text before banner/dialog/preview rendering.
- Add regressions for multiline string params.
- Record the single-line hardening contract in OpenSpec.

## Impact

- Prevents multiline param values from breaking compact operator surfaces.
- Keeps approval and tool previews readable and layout-safe.
