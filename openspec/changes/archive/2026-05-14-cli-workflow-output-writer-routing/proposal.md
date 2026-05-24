## Why

`lango workflow list`, `status`, `history`, and `cancel` still write their non-error output directly to process stdout. That makes wrapper capture and command-level testing inconsistent with the rest of the CLI writer-routing hardening work.

## What Changes

- Route workflow list/history table and empty-state output through `cmd.OutOrStdout()`
- Route workflow status detail output through `cmd.OutOrStdout()`
- Route workflow cancel confirmation through `cmd.OutOrStdout()`
- Add a small cancel seam plus command-level regression tests
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes workflow management output consistent with the CLI writer-routing hardening work
- Improves testability without changing workflow runtime semantics
