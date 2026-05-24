## Why

The README quick reference now lists some newer P2P operator families, but it still omits many implemented core P2P commands that already appear in the public CLI index and feature docs. That leaves the top-level operator reference incomplete.

## What Changes

- add the implemented core P2P command families to the README quick reference
- extend the executable README completeness guard so those implemented commands cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
