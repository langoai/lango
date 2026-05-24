## Why

The README quick reference still omitted the implemented `lango config` profile-management commands even though the public CLI index and dedicated config docs already documented them. That left the top-level operator reference incomplete for a core setup surface.

## What Changes

- add the implemented `lango config` profile-management commands to the README quick reference
- add an executable guard so those config-profile entries cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped config CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
