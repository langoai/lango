## Why

The README quick reference still omitted the implemented `lango version` and `lango health` top-level utility commands even though the public CLI index and existing README prose already treated them as shipped entry points.

## What Changes

- add the implemented `lango version` and `lango health` commands to the README quick reference
- add an executable guard so those top-level utility entries cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped top-level utility CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
