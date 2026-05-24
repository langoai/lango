## Why

The `lango p2p workspace` guidance commands still use vague phrases like "agent tools" or "workspace runtime integrations" for create/list/status. The actual live operator surface is the concrete `p2p_workspace_*` tool cluster, and the CLI should point users there directly.

## What Changes

- Replace vague `p2p workspace` guidance strings with concrete `p2p_workspace_*` tool references.
- Add CLI regressions covering the new guidance strings.
- Sync the public P2P CLI docs and CLI P2P management spec to the same concrete workspace-tool contract.

## Impact

- `cli-p2p-management`: workspace guidance becomes more specific and actionable.
- Operator UX: users are pointed to real tool names instead of generic runtime wording.
