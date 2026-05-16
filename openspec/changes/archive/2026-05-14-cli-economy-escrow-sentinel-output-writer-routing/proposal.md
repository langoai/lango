## Why

`lango economy escrow sentinel status` still writes active and disabled-state output directly to process stdout. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango economy escrow sentinel status` output through `cmd.OutOrStdout()`
- Add regression coverage for on-chain-disabled and escrow-disabled output capture
- Keep existing active guidance regression coverage
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the escrow sentinel inspection surface consistent with the CLI writer-routing hardening work
- Improves testability without changing sentinel semantics
