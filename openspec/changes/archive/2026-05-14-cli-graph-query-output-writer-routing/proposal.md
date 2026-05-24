## Why

`lango graph query` still writes human-readable and JSON output directly to process stdout instead of the Cobra command writer. That leaves another graph inspection surface inconsistent with the hardened CLI output contract.

## What Changes

- route `graph query` text and JSON output through `cmd.OutOrStdout()`
- add command-level writer capture tests for empty, table, and JSON query flows
- sync graph CLI docs and specs with the query output-writer contract

## Impact

- improves automation compatibility and testability for graph query
- keeps user-visible output unchanged while aligning behavior with the rest of the CLI
