## Why

The README quick reference still omitted most of the implemented automation surface even though the public CLI index and dedicated automation docs already documented it. That left the top-level operator reference incomplete for a stable family.

## What Changes

- add the implemented `lango cron`, `lango workflow`, and `lango bg` commands to the README quick reference
- add an executable guard so those automation entries cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped automation CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
