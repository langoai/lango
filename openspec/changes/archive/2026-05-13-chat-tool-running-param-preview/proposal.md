## Why

The chat transcript already surfaces tool lifecycle state, but the running row currently drops the request context entirely and shows only the tool name. Operators can see params in the approval surfaces, then lose them once execution starts. That weakens traceability during live debugging.

## What Changes

- Preserve a compact, deterministic param preview on running tool transcript rows.
- Reuse stable param ordering for the preview.
- Add regressions and sync docs/spec wording.

## Impact

- Improves tool-execution visibility without changing tool behavior.
- Makes the running transcript row more informative during active execution.
