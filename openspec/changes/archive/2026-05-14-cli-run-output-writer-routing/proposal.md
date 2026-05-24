## Why

`lango run` subcommands still write directly to process stdout instead of the Cobra command writer. That makes RunLedger inspection harder to capture in wrappers, harnesses, and tests even though the commands are otherwise read-only operator tools.

## What Changes

- route `lango run list`, `status`, and `journal` output through `cmd.OutOrStdout()`
- add command-level capture tests for all three subcommands
- sync RunLedger specs and public CLI docs with the output-writer contract

## Impact

- improves testability and automation compatibility for RunLedger CLI inspection
- keeps user-visible output unchanged while aligning with the rest of the CLI
