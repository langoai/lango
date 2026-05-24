## Why

The public P2P CLI reference still showed a workspace/git workflow example using `workspace_id` and `p2p_git_fetch`, but the actual registered tool surface uses `workspaceId` and does not currently expose a `p2p_git_fetch` tool. That example is now misleading for operators trying to follow the documented runtime path.

## What Changes

- Update the CLI reference example to use the actual tool parameter names and only the currently registered git workspace tools.
- Sync downstream docs coverage for the corrected runtime tool example.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `downstream-docs-sync`: the public P2P CLI git-bundle example now matches the actual runtime tool surface.

## Impact

- Affected docs: `docs/cli/p2p.md`
- Affected specs: `openspec/specs/downstream-docs-sync/spec.md`
