## Why

`lango config get`, `set`, and `keys` still write directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests and leaves the scripting-oriented config CLI inconsistent with the rest of the hardened command set.

## What Changes

- route `config get`, `set`, and `keys` output through `cmd.OutOrStdout()`
- update shared value-printing helpers to accept an explicit writer
- add command-level capture tests for all three subcommands
- sync config CLI specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for scripting-oriented config commands
- keeps user-visible output unchanged while aligning with other commands
