## Why

`lango agent status` still writes table and JSON output directly to process stdout instead of the Cobra command writer. That weakens wrapper capture and leaves this inspection surface inconsistent with the hardened CLI baseline.

## What Changes

- route `agent status` table and JSON output through `cmd.OutOrStdout()`
- convert command-level status tests to capture the Cobra output writer instead of process stdout
- sync agent CLI docs and specs with the output-writer contract

## Impact

- improves automation compatibility and testability for agent status inspection
- keeps user-visible output unchanged while aligning behavior with the rest of the CLI
