## Why

`lango economy escrow show` still writes its detailed-config, disabled-state, and ID-guidance output directly to process stdout. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango economy escrow show` output through `cmd.OutOrStdout()`
- Add regression coverage for detailed-config and disabled-state output capture
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the economy escrow detail surface consistent with the CLI writer-routing hardening work
- Improves testability without changing escrow semantics
