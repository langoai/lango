# Design

## Approach

The CLI will open the same local workspace persistence model used by the
runtime: `internal/p2p/workspace.Manager` backed by
`<workspace data dir>/workspaces.db`. The data directory resolves from
`p2p.workspace.dataDir`; when empty, it uses the runtime default
`~/.lango/workspaces` layout.

## Boundaries

- The CLI owns only local workspace lifecycle commands:
  `create`, `list`, `status`, `join`, and `leave`.
- The CLI does not start the P2P node, subscribe to GossipSub, or expose direct
  git bundle control in this change.
- The existing server-backed `p2p_workspace_*` tools remain the live
  distributed path.

## Command Behavior

- `create <name> --goal <text>` persists a forming workspace and prints the ID,
  name, goal, status, and member count.
- `list` reads persisted local workspaces and prints an empty state or a compact
  table. JSON output returns `{ "workspaces": [...], "count": N }`.
- `status <workspace-id>` reads one workspace and prints details including
  members.
- `join <workspace-id>` and `leave <workspace-id>` mutate local membership and
  print actionable success output.
- Feature gates remain strict: `p2p.enabled` and `p2p.workspace.enabled` must
  both be true.

## Testing

Tests use a temporary `p2p.workspace.dataDir` so they exercise real BoltDB
persistence without touching the user's home directory. The regression tests
verify JSON output, persistence across command invocations, and membership
mutation.
