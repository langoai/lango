## Why

The public economy quick references still omitted implemented `lango economy escrow list`, `lango economy escrow show`, and `lango economy escrow sentinel status` commands even though the dedicated economy CLI docs already documented them. That made the top-level quick references less complete than the shipped surface.

## What Changes

- add the implemented economy escrow quick-reference entries to `README.md` and `docs/cli/index.md`
- add an executable guard so those implemented entries cannot silently disappear again

## Impact

- public quick-reference docs better match the actual shipped economy CLI surface
- reduced operator confusion when skimming top-level command lists
- stronger regression protection for quick-reference completeness drift
