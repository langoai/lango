## Why

The README quick reference still omitted the implemented `lango metrics` command family even though the public CLI index and dedicated metrics docs already documented it. That left the top-level operator reference incomplete for another stable surface.

## What Changes

- add the implemented `lango metrics` commands to the README quick reference
- add an executable guard so those metrics entries cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped metrics CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
