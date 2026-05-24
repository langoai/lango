## Why

The architecture project-structure row for `cli/security/` already distinguishes canonical `change-passphrase` from deprecated `migrate-passphrase`, but it still omits parts of the current shipped surface such as recovery commands and KMS wrap/detach.

## What Changes

- expand the `cli/security/` architecture row to include recovery and the full KMS subfamily
- strengthen the executable docs-quality guard for the architecture security row
- sync the main docs-only and test-coverage specs with the fuller security surface contract

## Impact

- more truthful architecture inventory docs
- better discoverability of current security operations
- stronger regression protection against partial security-surface drift
