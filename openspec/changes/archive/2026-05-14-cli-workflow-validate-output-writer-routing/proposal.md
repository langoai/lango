## Why

`lango workflow validate` still writes both table and JSON output directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests and leaves this validation-only CLI surface inconsistent with the rest of the hardened command set.

## What Changes

- route `workflow validate` table output through `cmd.OutOrStdout()`
- route `workflow validate --json` output through the same writer
- add temp-file based command-level capture tests for both modes
- sync workflow CLI specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for workflow validation
- keeps user-visible output unchanged while aligning with other commands
