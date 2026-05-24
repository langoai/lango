## Why

`AcquireNonInteractive()` still routed keyring warnings directly to process-global stderr, which made the non-interactive fallback path harder to verify than the newly seam-based interactive acquisition code.

## What Changes

- Add an internal `acquireNonInteractiveWithIO(...)` helper that accepts an injected stderr writer
- Add deterministic tests for the keyring warning and no-warning (`ErrNotFound`) branches
- Record the non-interactive stderr seam in OpenSpec

## Impact

- Improves testability of a security-critical non-interactive path
- Aligns `AcquireNonInteractive()` with the same injected-writer philosophy used elsewhere in passphrase acquisition
