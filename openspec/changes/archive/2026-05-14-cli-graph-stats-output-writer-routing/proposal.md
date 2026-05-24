## Why

`lango graph stats` still writes both table and JSON output directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests and leaves the graph inspection commands inconsistent across subcommands.

## What Changes

- route `graph stats` table output through `cmd.OutOrStdout()`
- route `graph stats --json` output through the same writer
- add temp-store based command-level capture tests for both modes
- sync graph CLI specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for graph statistics inspection
- keeps user-visible output unchanged while aligning with other commands
