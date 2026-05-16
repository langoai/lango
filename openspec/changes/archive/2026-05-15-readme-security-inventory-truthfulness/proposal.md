## Why

The README internal CLI inventory still compresses the security family too aggressively. It does not distinguish canonical `change-passphrase` from deprecated `migrate-passphrase`, and it hides the current recovery and KMS sub-surfaces behind vague shorthand.

## What Changes

- update the README internal tree security row to reflect the current canonical/deprecated, recovery, keyring, and KMS command families
- add an executable guard for the README internal security inventory wording
- sync the main docs-only and test-coverage specs with that truthfulness contract

## Impact

- more truthful public package-tree documentation
- better discoverability of the real security surface
- stronger regression protection against stale inventory shorthand
