## Why

`lango agent tools` still writes human-readable and JSON output directly to process stdout instead of the Cobra command writer. That makes wrapper capture inconsistent with the already-hardened `agent status` and `agent hooks` surfaces.

## What Changes

- route `agent tools` table and JSON output through `cmd.OutOrStdout()`
- route empty/filter-miss text output through the same writer
- add command-level writer capture tests for table, JSON, and missing-category flows
- sync agent CLI docs and specs with the output-writer contract

## Impact

- improves automation compatibility and testability for tool category inspection
- keeps user-visible output unchanged while aligning behavior with the rest of the CLI
