## Why

The README quick reference still omitted the implemented `lango mcp` command family even though the public CLI index and dedicated MCP docs already documented it. That left the top-level operator reference incomplete for another shipped control surface.

## What Changes

- add the implemented `lango mcp` commands to the README quick reference
- add an executable guard so those MCP entries cannot silently disappear again

## Impact

- README discoverability better matches the actual shipped MCP CLI surface
- reduced operator confusion when skimming the top-level command list
- stronger regression protection for README completeness drift
