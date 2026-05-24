## Why

The README quick reference already listed the newer P2P git guidance commands, but it still omitted the implemented `workspace`, `team`, and `zkp` P2P command families. That makes the top-level operator reference less complete than the public CLI index and feature docs.

## What Changes

- add the implemented `p2p workspace`, `p2p team`, and `p2p zkp` entries to the README quick reference
- add an executable guard so those implemented P2P command families cannot silently disappear from the README again

## Impact

- README discoverability better matches the actual shipped CLI surface
- reduced operator confusion when using the top-level command list
- stronger regression protection for README completeness drift
