## Why

The bootstrap lifecycle already specifies the secure-storage offer and stderr outcomes when storing an interactive passphrase into a hardware keyring provider, but those branches were weakly verified in tests.

## What Changes

- Add narrow seams for passphrase acquisition, confirmation, and bootstrap stderr output
- Add deterministic regression coverage for secure-storage offer, successful store, entitlement warning, and generic store failure warning

## Impact

- Improves confidence in a security-critical bootstrap path without changing operator-visible behavior
- Makes future refactors safer by removing reliance on implicit globals in tests
