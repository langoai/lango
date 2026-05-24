## Why

The README quick reference still omitted the implemented `lango a2a` command family even though the public CLI index and dedicated A2A docs already documented it. That left the top-level operator reference incomplete for a shipped interoperability surface.

## What Changes

- add the implemented `lango a2a` commands to the README quick reference
- add an executable guard so those A2A entries cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped A2A CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
