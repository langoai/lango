## Why

`lango a2a card` and `lango a2a check` still write both table and JSON output directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests and leaves the A2A CLI inconsistent with the other hardened command groups.

## What Changes

- route `a2a card` and `a2a check` table output through `cmd.OutOrStdout()`
- route both JSON modes through the same writer
- add command-level capture tests for local and remote card inspection
- sync A2A CLI specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for A2A inspection commands
- keeps user-visible output unchanged while aligning with the rest of the CLI
