## Why

`lango account module install` still writes its success summary directly to process stdout. That makes wrapper capture and command-level testing inconsistent with the rest of the smart account CLI writer-routing hardening work.

## What Changes

- Route `lango account module install` success output through `cmd.OutOrStdout()`
- Add a small module-install seam for deterministic command-level tests
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the smart account module management surface consistent with the CLI writer-routing hardening work
- Improves testability without changing module installation semantics
