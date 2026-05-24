## Why

The `lango alerts` command family is implemented, wired into the root CLI, and documented in dedicated alerts docs, but it was still missing from the public quick references. That left operational alerting less discoverable than other shipped inspection surfaces.

## What Changes

- add the implemented `lango alerts list` and `lango alerts summary` commands to `README.md`
- add the implemented `lango alerts list` and `lango alerts summary` commands to `docs/cli/index.md`
- add an executable guard so those alerts quick-reference entries cannot silently disappear again

## Impact

- public quick references better match the actual shipped alerts CLI surface
- reduced operator confusion when skimming top-level and CLI index command lists
- stronger regression protection for quick-reference completeness drift
