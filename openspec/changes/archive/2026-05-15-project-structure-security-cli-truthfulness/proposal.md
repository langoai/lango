## Why

The public architecture project-structure reference still described `cli/security/` as if `migrate-passphrase` were the representative passphrase-rotation command. That no longer matched the actual security CLI surface, where `change-passphrase` is canonical and `migrate-passphrase` is deprecated.

## What Changes

- update `docs/architecture/project-structure.md` to describe `change-passphrase` as the canonical security CLI surface
- add a regression guard so the architecture row keeps the current canonical/deprecated distinction

## Impact

- architecture docs better match the shipped security command surface
- reduced confusion when readers inspect module ownership from the project-structure page
- stronger regression protection for architecture-doc truthfulness
