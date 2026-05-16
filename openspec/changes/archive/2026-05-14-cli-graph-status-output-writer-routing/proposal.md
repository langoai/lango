## Why

`lango graph status` still writes both table and JSON output directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests and leaves this graph inspection command inconsistent with the rest of the hardened CLI surfaces.

## What Changes

- route `lango graph status` table output through `cmd.OutOrStdout()`
- route `lango graph status --json` output through the same writer
- add command-level capture tests for both modes
- sync graph CLI specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for graph status inspection
- keeps user-visible output unchanged while aligning with other commands
