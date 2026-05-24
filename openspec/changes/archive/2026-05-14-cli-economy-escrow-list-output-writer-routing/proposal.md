## Why

`lango economy escrow list` still writes its summary and disabled-state output directly to process stdout. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango economy escrow list` output through `cmd.OutOrStdout()`
- Add regression coverage for summary, economy-disabled, and escrow-disabled output capture
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the economy escrow summary surface consistent with the CLI writer-routing hardening work
- Improves testability without changing escrow semantics
