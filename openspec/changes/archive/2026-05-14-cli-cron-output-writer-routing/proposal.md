## Why

`lango cron` still writes table and confirmation output directly to process stdout instead of the Cobra command writer. That breaks deterministic capture for wrappers and command-level tests and leaves cron inconsistent with the rest of the hardened CLI surfaces.

## What Changes

- route `cron add`, `list`, `delete`, `pause`, `resume`, and `history` output through `cmd.OutOrStdout()`
- add command-level writer-capture tests for representative table and confirmation flows
- sync cron CLI specs and docs with the output-writer contract

## Impact

- improves automation compatibility and testability for cron management commands
- keeps user-visible output unchanged while aligning routing behavior with other commands
