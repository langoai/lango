## MODIFIED Requirements

### Requirement: Workspace operator surfaces expose local CLI lifecycle control
Workspace operator surfaces SHALL distinguish direct local CLI lifecycle
control from server-backed distributed workspace tools.

#### Scenario: Workspace CLI local control
- **WHEN** a user reads or runs `lango p2p workspace create`, `list`, `status`, `join`, or `leave`
- **THEN** the CLI SHALL manage local workspace persistence directly
- **AND** docs SHALL not describe those workspace commands as guidance-only

#### Scenario: Distributed workspace tools remain documented
- **WHEN** a user needs live distributed workspace messaging or peer exchange
- **THEN** docs SHALL still point to the running server and `p2p_workspace_*` tools
