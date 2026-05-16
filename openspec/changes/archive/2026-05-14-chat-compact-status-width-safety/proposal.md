## Why

Compact `status` and `approval` transcript rows currently compose their label and body without accounting for the final combined width. Approval rows recently gained richer traceability text, which makes narrow-terminal overflow more likely, and status rows have the same structural risk.

## What Changes

- Make compact `status` and `approval` transcript rows width-aware as complete rows, not just body-only truncation.
- Add regressions for narrow-width status and approval rows.
- Record the compact-row width-safety contract in OpenSpec.

## Impact

- Prevents compact transcript rows from overrunning narrow terminals.
- Keeps richer traceability text without destabilizing the transcript layout.
