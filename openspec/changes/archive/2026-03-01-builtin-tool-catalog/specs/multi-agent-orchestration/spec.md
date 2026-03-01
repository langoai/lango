## MODIFIED Requirements

### Requirement: Orchestrator universal tools
The orchestration `Config` struct SHALL include a `UniversalTools` field. When `UniversalTools` is non-empty, `BuildAgentTree` SHALL adapt and assign these tools directly to the orchestrator agent.

#### Scenario: Orchestrator receives dispatcher tools
- **WHEN** `Config.UniversalTools` contains builtin_list and builtin_invoke
- **THEN** the orchestrator agent SHALL have those tools available for direct invocation
- **AND** the orchestrator instruction SHALL mention builtin_list and builtin_invoke capabilities

#### Scenario: No universal tools
- **WHEN** `Config.UniversalTools` is nil or empty
- **THEN** the orchestrator SHALL have no direct tools (existing behavior)
- **AND** the instruction SHALL state "You do NOT have tools"

### Requirement: Builtin prefix exclusion from partitioning
`PartitionTools` SHALL skip any tool whose name starts with `builtin_`. These tools SHALL NOT appear in any sub-agent's tool set or in the Unmatched list.

#### Scenario: Builtin tools skipped during partitioning
- **WHEN** tools include `builtin_list` and `builtin_invoke` alongside normal tools
- **THEN** `PartitionTools` SHALL assign normal tools to their respective roles
- **AND** `builtin_*` tools SHALL not appear in any RoleToolSet field
