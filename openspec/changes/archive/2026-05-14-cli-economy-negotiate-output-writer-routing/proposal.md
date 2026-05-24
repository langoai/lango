## Why

`lango economy negotiate status` still writes its enabled/disabled output directly to process stdout. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango economy negotiate status` output through `cmd.OutOrStdout()`
- Add regression coverage for enabled and disabled output capture
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the economy negotiation inspection surface consistent with the CLI writer-routing hardening work
- Improves testability without changing negotiation semantics
