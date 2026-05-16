## Why

The README quick reference still omitted the implemented `lango approval` command family even though the public CLI index and dedicated approval docs already documented it. That left the top-level operator reference incomplete for a shipped control surface.

## What Changes

- add the implemented `lango approval status` command to the README quick reference
- add an executable guard so that approval entry cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped approval CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
