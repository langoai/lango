## MODIFIED Requirements

### Requirement: P2P workspace CLI commands manage local workspace state
The `lango p2p workspace` command family SHALL use the local workspace manager
for lifecycle and membership operations instead of only printing guidance.

#### Scenario: Workspace create persists a local workspace
- **WHEN** `lango p2p workspace create <name>` is run with P2P and workspaces enabled
- **THEN** the command SHALL persist a workspace record in the configured workspace data directory
- **AND** table and JSON output SHALL include the workspace ID, name, goal, status, and member count

#### Scenario: Workspace list reads local persisted state
- **WHEN** `lango p2p workspace list` is run after one or more local workspaces exist
- **THEN** the command SHALL list those workspaces from local persistence
- **AND** JSON output SHALL return a `workspaces` array and `count`

#### Scenario: Workspace status reads local persisted state
- **WHEN** `lango p2p workspace status <workspace-id>` is run for an existing local workspace
- **THEN** the command SHALL print workspace details and member rows from local persistence

#### Scenario: Workspace membership commands mutate local state
- **WHEN** `lango p2p workspace join <workspace-id>` or `leave <workspace-id>` is run
- **THEN** the command SHALL update local workspace membership through the workspace manager

#### Scenario: Workspace commands remain feature gated
- **WHEN** P2P or workspace support is disabled
- **THEN** workspace commands SHALL fail before opening or mutating workspace state

### Requirement: P2P git CLI remains guidance-oriented
The `lango p2p git` command family SHALL remain a guidance-oriented surface
until direct repository control is explicitly designed.

#### Scenario: Git command guidance remains explicit
- **WHEN** a user runs `lango p2p git` commands
- **THEN** the output SHALL continue to direct the operator to the server-backed runtime or `p2p_git_*` tools
