## Why

The README quick reference still omitted the implemented `lango agent` inspection commands and the `lango graph` command family even though the public CLI index already documented both surfaces. That left the top-level operator reference incomplete for two shipped inspection families.

## What Changes

- add the implemented `lango agent status/list/tools/hooks` commands to the README quick reference
- add the implemented `lango graph` commands to the README quick reference
- add executable guards so those README entries cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped inspection and graph CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
