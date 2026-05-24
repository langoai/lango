## Why

The README quick reference still omitted the implemented `lango account` command family even though the public CLI index and dedicated smart-account docs already documented it. That left the top-level operator reference incomplete for one of the larger shipped surfaces.

## What Changes

- add the implemented `lango account` commands to the README quick reference
- add an executable guard so those smart-account entries cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped smart-account CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
