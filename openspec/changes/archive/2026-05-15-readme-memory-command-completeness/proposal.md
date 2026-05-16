## Why

The README quick reference still omitted the implemented `lango memory` command family even though the public CLI index and dedicated memory docs already documented it. That left the top-level operator reference incomplete for another shipped surface.

## What Changes

- add the implemented `lango memory` commands to the README quick reference
- add an executable guard so those memory entries cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped memory CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
