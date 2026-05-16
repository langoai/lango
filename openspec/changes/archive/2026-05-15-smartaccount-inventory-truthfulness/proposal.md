## Why

The public smart-account inventory docs still described abbreviated command subsets even though the shipped CLI already exposed `session create/revoke`, `module install`, `policy set`, and `paymaster approve`. That made the smart-account surface look smaller than it really is.

## What Changes

- expand the representative command list in `docs/cli/smartaccount.md`
- update `docs/architecture/project-structure.md` and the README internal tree inventory to the current smart-account command surface
- add an executable guard so those inventory docs keep the current command set

## Impact

- smart-account docs better match the shipped command surface
- reduced operator confusion when scanning command inventories
- stronger regression protection for smart-account docs drift
