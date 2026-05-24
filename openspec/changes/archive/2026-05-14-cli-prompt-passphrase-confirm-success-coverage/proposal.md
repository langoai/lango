## Why

The prompt seam work already covered read errors and mismatch behavior, but the successful `PassphraseConfirm(...)` path still lacked an explicit deterministic regression.

## What Changes

- Add a seam-based success regression for `PassphraseConfirm(...)`
- Verify that the prompt sequence, hidden-input reads, and returned passphrase all match the expected happy path

## Impact

- Completes coverage of the seam-aware prompt confirmation flow
- Makes future prompt refactors safer without changing runtime behavior
