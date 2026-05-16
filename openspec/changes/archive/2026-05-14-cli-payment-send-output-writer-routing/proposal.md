## Why

`lango payment send` still writes its confirmation prompt and success output directly to process stdout instead of the Cobra command writer and input stream. That breaks output capture for wrappers and tests on one of the most operator-visible payment flows.

## What Changes

- route the confirmation prompt and success output through `cmd.OutOrStdout()`
- read confirmation input from `cmd.InOrStdin()`
- route JSON output through the same writer
- add lightweight send seams and command-level capture tests
- sync payment CLI specs and docs with the output/input-writer contract

## Impact

- improves testability and automation compatibility for payment send flows
- keeps user-visible behavior unchanged while aligning with the rest of the CLI
