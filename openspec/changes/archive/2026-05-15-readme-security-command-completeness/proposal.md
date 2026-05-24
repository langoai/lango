## Why

The README quick reference still omitted the implemented `lango security` command family even though the public CLI index and dedicated security docs already documented it. That left the top-level operator reference incomplete for a major shipped surface.

## What Changes

- add the implemented `lango security` commands to the README quick reference
- add an executable guard so those security entries cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped security CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
