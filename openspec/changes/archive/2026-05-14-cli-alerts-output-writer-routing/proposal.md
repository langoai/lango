## Why

`lango alerts` still writes table and JSON output directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers, tests, and automation even though the command is otherwise read-only and simple.

## What Changes

- route alerts list/summary table output through `cmd.OutOrStdout()`
- route alerts JSON output through the same writer
- add command-level capture tests using a local HTTP fixture
- sync alerting specs and CLI docs with the output-writer contract

## Impact

- makes `lango alerts` consistent with the hardened status/doctor commands
- improves testability without changing user-facing payloads
