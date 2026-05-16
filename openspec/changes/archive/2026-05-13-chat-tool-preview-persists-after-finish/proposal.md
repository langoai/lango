## Why

The running tool transcript row now includes a compact param preview, but that context disappears as soon as the same row transitions into success or error state. Operators then lose immediate visibility into which input produced the completed output or failure.

## What Changes

- Keep the compact param preview on success and error tool transcript rows.
- Add regressions for the persisted preview.
- Sync docs/spec wording with the lifecycle-continuity contract.

## Impact

- Improves traceability across tool lifecycle transitions.
- Keeps execution context visible after completion without changing tool behavior.
