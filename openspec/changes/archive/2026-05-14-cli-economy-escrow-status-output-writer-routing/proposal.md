## Why

`lango economy escrow status` still writes its enabled/disabled output directly to process stdout. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango economy escrow status` output through `cmd.OutOrStdout()`
- Add regression coverage for enabled and disabled output capture
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the economy escrow inspection surface consistent with the CLI writer-routing hardening work
- Improves testability without changing escrow semantics
