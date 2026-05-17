# Design

## Approach

Keep the runtime pipeline unchanged. Add a focused bootstrap package test that reads the package source files and verifies:

- `phases.go` documents the current 12-phase count.
- `bootstrap.go` documents each current `DefaultPhases()` phase name.
- The old 7-step source comment wording is absent from the public `Run` inventory comment.

Then update the source comments to match the real sequence.

## Why Source-Comment Coverage

The existing `TestDefaultPhases_Returns12Phases` already verifies runtime phase order. It did not catch stale explanatory comments. A source-comment guard is appropriate here because the defect is documentation drift inside production source, not runtime behavior.

## Non-Goals

- Do not change bootstrap phase ordering.
- Do not change phase names.
- Do not update public docs unless the source audit finds public docs with the same stale 7-step sequence.
