## Why

The `lango extension` command family is implemented and wired into the root CLI, and README already describes the extension-pack workflow, but the public quick references still omitted the command surface. That left extension-pack operations less discoverable than other shipped CLI families.

## What Changes

- add the implemented `lango extension inspect/install/list/remove` commands to `README.md`
- add the implemented `lango extension inspect/install/list/remove` commands to `docs/cli/index.md`
- add an executable guard so those extension quick-reference entries cannot silently disappear again

## Impact

- public quick references better match the actual shipped extension CLI surface
- reduced operator confusion when skimming top-level and CLI index command lists
- stronger regression protection for quick-reference completeness drift
