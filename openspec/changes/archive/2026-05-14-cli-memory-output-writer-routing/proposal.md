## Why

`lango memory list` and `lango memory status` still write table and JSON output directly to process stdout instead of the Cobra command writer. That makes wrapper capture inconsistent with the hardened CLI surfaces and leaves the observational memory inspection path harder to test.

## What Changes

- route `memory list` table and JSON output through `cmd.OutOrStdout()`
- route `memory status` table and JSON output through the same writer
- add command-level writer capture tests for representative empty-session flows
- sync memory CLI docs and specs with the output-writer contract

## Impact

- improves automation compatibility and testability for observational memory inspection commands
- keeps user-visible output unchanged while aligning runtime behavior with the rest of the CLI
