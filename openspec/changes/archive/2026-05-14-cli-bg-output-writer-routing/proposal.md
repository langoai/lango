## Why

`lango bg` subcommands still write directly to process stdout instead of the Cobra command writer. That makes background-task inspection and control output harder to capture in wrappers and tests even though the commands are simple operator-facing tools.

## What Changes

- route `lango bg list`, `status`, `cancel`, and `result` output through `cmd.OutOrStdout()`
- add command-level capture tests using a real in-memory background manager
- sync the background CLI spec and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for background-task CLI management
- keeps user-visible output unchanged while aligning with other commands
