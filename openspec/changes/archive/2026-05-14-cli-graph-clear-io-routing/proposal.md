## Why

`lango graph clear` still prints prompts and results directly to process stdout and reads confirmation input directly from process stdin. That makes interactive automation and command-level testing inconsistent with the rest of the hardened CLI surfaces.

## What Changes

- route `graph clear` prompt and confirmation output through `cmd.OutOrStdout()`
- read confirmation input through `cmd.InOrStdin()`
- add command-level capture tests for abort, confirm, and force flows
- sync graph CLI docs and specs with the command stream contract

## Impact

- improves automation compatibility and testability for destructive graph operations
- keeps user-visible prompt text unchanged while aligning stream behavior with Cobra conventions
