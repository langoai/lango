## Why

`lango learning status` and `lango learning history` still write table and JSON output directly to process stdout instead of the Cobra command writer. That makes capture inconsistent with the hardened CLI surfaces and weakens wrapper/test integration.

## What Changes

- route `learning status` table and JSON output through `cmd.OutOrStdout()`
- route `learning history` table and JSON output through the same writer
- add command-level writer capture tests for representative status/history flows
- sync learning CLI docs and specs with the output-writer contract

## Impact

- improves automation compatibility and testability for learning inspection commands
- keeps user-visible output unchanged while aligning runtime behavior with other commands
