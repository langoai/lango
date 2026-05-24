## Why

The README quick reference still omitted the implemented `lango payment` command family even though the public CLI index and dedicated payment docs already documented it. That left the top-level operator reference incomplete for a major shipped surface.

## What Changes

- add the implemented `lango payment` commands to the README quick reference
- add an executable guard so those payment entries cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped payment CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
