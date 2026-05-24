## Why

`lango economy budget status` still writes enabled/disabled output and task guidance directly to process stdout. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango economy budget status` output through `cmd.OutOrStdout()`
- Add regression coverage for enabled, disabled, and task-guidance output capture
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the economy budget inspection surface consistent with the CLI writer-routing hardening work
- Improves testability without changing budget semantics
