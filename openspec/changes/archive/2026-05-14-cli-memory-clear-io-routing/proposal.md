## Why

`lango memory clear` still prints prompts and results directly to process stdout and reads confirmation input directly from process stdin. That makes interactive automation and command-level testing inconsistent with the hardened CLI surfaces.

## What Changes

- route `memory clear` prompt and result output through `cmd.OutOrStdout()`
- read confirmation input through `cmd.InOrStdin()`
- add command-level capture tests for abort, confirm, and force flows
- sync memory CLI docs and specs with the command stream contract

## Impact

- improves automation compatibility and testability for destructive memory operations
- keeps user-visible prompt text unchanged while aligning stream behavior with Cobra conventions
