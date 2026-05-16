## Why

`lango approval status` still writes both table and JSON output directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests, and makes the approval CLI inconsistent with the rest of the hardened command surfaces.

## What Changes

- route approval status table output through `cmd.OutOrStdout()`
- route approval status JSON output through the same writer
- add command-level capture tests for both modes
- sync approval specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for the approval CLI
- keeps user-visible output unchanged while aligning with other commands
