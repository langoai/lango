## Why

The `lango p2p workspace` guidance commands still mention a generic "runtime API" for create/list/join/leave operations, but the stable operator surface is the server-backed runtime plus the actual `p2p_workspace_*` tools. The public docs already describe that truthful path, so the CLI output has drifted behind the documented contract.

## What Changes

- Replace stale "runtime API" wording in `lango p2p workspace create/list/join/leave` with truthful server-backed runtime guidance.
- Add CLI regressions covering the new workspace guidance strings.
- Tighten the CLI P2P management spec so workspace guidance does not imply a nonexistent public control API.

## Impact

- `cli-p2p-management`: workspace guidance stays aligned with the actual runtime surface.
- Operator UX: users are no longer sent toward an invented generic runtime API path for workspace control.
