## Why

`lango agent list` still writes human-readable and JSON output directly to process stdout instead of the Cobra command writer. That leaves another high-traffic inspection surface inconsistent with the hardened agent CLI commands.

## What Changes

- route `agent list` table and JSON output through `cmd.OutOrStdout()`
- route empty/filter-miss and remote section spacing through the same writer
- add command-level writer capture tests for representative list flows
- sync agent CLI docs and specs with the output-writer contract

## Impact

- improves automation compatibility and testability for agent discovery output
- keeps user-visible output unchanged while aligning behavior with the rest of the CLI
