## Why

The README quick reference still omitted the implemented `lango agent` diagnostics family even though the public CLI index already documented it. That left the top-level operator reference incomplete for one of the more important diagnostic surfaces.

## What Changes

- add the implemented `lango agent` diagnostics commands to the README quick reference
- add an executable guard so those agent-diagnostics entries cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped diagnostics CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
