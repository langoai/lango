## Why

The repository already had dedicated `docs/cli/core.md` and `docs/cli/status.md` references, but the public CLI index still treated those surfaces only as top-level quick-reference rows. That left the index structure less coherent than the rest of the CLI docs inventory.

## What Changes

- add dedicated `Core Commands` and `Status Dashboard` sections to `docs/cli/index.md`
- add an executable guard so those index sections cannot silently disappear again

## Impact

- CLI index structure better matches the actual dedicated CLI docs inventory
- core and status command families become easier to discover when scanning section-by-section
- stronger regression protection for index structure drift
