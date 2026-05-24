## Why

Public quick references now list both `lango security change-passphrase` and `lango security migrate-passphrase`, but the wording still treated the deprecated path like a normal passphrase rotation command. That obscured the canonical migration path and the operational cost difference between the two commands.

## What Changes

- update `README.md` and `docs/cli/index.md` so `change-passphrase` is described as the canonical non-reencrypting path
- mark `migrate-passphrase` as a deprecated legacy full re-encryption path in those same quick references
- add an executable guard so the canonical/deprecated distinction does not drift again

## Impact

- public quick references better match the actual security command contract
- operators are less likely to pick the deprecated heavy-weight migration path by mistake
- stronger regression protection for security quick-reference truthfulness
