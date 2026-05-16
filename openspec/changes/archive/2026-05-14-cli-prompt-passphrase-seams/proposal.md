## Why

`prompt.Passphrase(...)` and `PassphraseConfirm(...)` still depended directly on `os.Stdout`, `syscall.Stdin`, and `term.ReadPassword`. That left the hidden-input prompt path harder to verify than the other recently seam-aware confirmation flows.

## What Changes

- Add small seams for passphrase prompt output, stdin fd lookup, and password reading
- Add prompt package tests for success, read error, and confirmation mismatch
- Record the hidden-input seam contract in OpenSpec

## Impact

- Improves testability of a common interactive foundation without changing runtime behavior
- Makes future prompt refactors safer
