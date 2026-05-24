## Why

`lango graph add` still writes human-readable and JSON output directly to process stdout instead of the Cobra command writer. That leaves another graph mutation surface inconsistent with the hardened CLI output contract.

## What Changes

- route `graph add` text and JSON output through `cmd.OutOrStdout()`
- add command-level writer capture tests for representative add flows
- sync graph CLI docs and specs with the add output-writer contract

## Impact

- improves automation compatibility and testability for graph mutation commands
- keeps user-visible output unchanged while aligning behavior with the rest of the CLI
