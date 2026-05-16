## Why

`lango doctor` still writes both human-readable and JSON output directly to process stdout instead of the Cobra command writer. That breaks output capture for tests, wrappers, and automation even though the command otherwise behaves like the rest of the CLI.

## What Changes

- route doctor table and JSON output through `cmd.OutOrStdout()`
- add command-level capture tests for both table and JSON modes
- sync doctor specs and public CLI docs with the command-writer contract

## Impact

- improves testability and wrapper compatibility for `lango doctor`
- keeps user-visible output unchanged while making routing consistent
