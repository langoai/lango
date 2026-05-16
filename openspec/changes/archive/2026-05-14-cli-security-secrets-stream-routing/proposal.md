## Why

`lango security secrets list`, `set`, and `delete` still use direct stdout/stderr and ad-hoc stdin reads. That makes wrapper capture and interactive automation inconsistent with the hardened CLI surfaces.

## What Changes

- route `security secrets` table and JSON output through `cmd.OutOrStdout()`
- route delete confirmation through Cobra input/output streams
- keep warnings/errors on `cmd.ErrOrStderr()` where applicable
- add command-level tests using a persistent temp DB-backed bootloader
- sync secrets CLI docs and specs with the command stream contract

## Impact

- improves automation compatibility and testability for secret management
- keeps user-visible behavior unchanged while aligning stream handling with Cobra conventions
