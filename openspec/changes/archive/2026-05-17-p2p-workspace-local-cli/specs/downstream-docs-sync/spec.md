## MODIFIED Requirements

### Requirement: Public P2P workspace docs match CLI behavior
Public README and CLI docs SHALL describe `lango p2p workspace` commands as
local workspace lifecycle commands once they persist local state directly.

#### Scenario: Quick references use direct action wording
- **WHEN** a user reads the README or CLI index quick references
- **THEN** workspace command descriptions SHALL use direct action wording such as create, list, show, join, and leave
- **AND** they SHALL not use guidance-only phrasing such as `Describe how to` for workspace commands

#### Scenario: Dedicated P2P CLI docs explain the split
- **WHEN** a user reads `docs/cli/p2p.md`
- **THEN** the workspace section SHALL explain that CLI commands manage local workspace records
- **AND** it SHALL explain that distributed messaging and peer exchange still use the running server and `p2p_workspace_*` tools
