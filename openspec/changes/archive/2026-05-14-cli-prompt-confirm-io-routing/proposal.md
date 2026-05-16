## Why

`prompt.Confirm(...)` still uses process-global stdin/stdout. Recovery setup now routes most output through Cobra streams, but the written-down confirmation prompt still bypasses command-level input/output capture.

## What Changes

- Add `prompt.ConfirmIO(...)` for stream-aware yes/no confirmation
- Update recovery setup to use command streams for the written-down confirmation prompt
- Add regression coverage and align docs/OpenSpec with the actual stream contract

## Impact

- Removes another process-global I/O dependency from a security-critical CLI flow
- Improves command-level testability without changing confirmation semantics
