# P2P Workspace Local CLI

## Why

`lango p2p workspace create/list/status/join/leave` currently behaves as a
guidance surface even though the runtime already ships a BoltDB-backed
workspace manager. Operators who use the CLI get instructions to switch to
server-backed tools instead of a directly useful local workspace command path.

## What Changes

- Connect the workspace CLI commands to the existing local workspace manager.
- Keep scope limited to workspace lifecycle inspection and membership mutation:
  create, list, status, join, and leave.
- Preserve existing P2P and workspace feature gates.
- Keep `lango p2p git` guidance-oriented in this change.
- Update public docs and specs so they no longer describe workspace commands
  as guidance-only.

## Impact

- Modified capabilities: `cli-p2p-management`, `p2p-workspace`,
  `downstream-docs-sync`.
- Runtime behavior changes for `lango p2p workspace` commands only.
- No direct network server control is introduced.
