## Why

`lango p2p git fetch` still tells operators to use a nonexistent `p2p_git_fetch` tool. That is a real CLI truth bug: the public docs already describe fetch as a server-backed runtime path without a dedicated fetch tool, but the command output still points users to an invalid recovery path.

## What Changes

- Replace the stale `p2p_git_fetch` guidance in `lango p2p git fetch` with truthful server-backed runtime guidance.
- Add a CLI regression that locks the new message and proves the stale tool name is gone.
- Sync the CLI P2P management spec to the same fetch guidance contract.

## Impact

- `cli-p2p-management`: git fetch guidance becomes truthful and actionable.
- Operator UX: users are no longer sent to a nonexistent tool surface.
