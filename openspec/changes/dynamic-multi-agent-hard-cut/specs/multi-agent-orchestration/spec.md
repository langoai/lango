## MODIFIED Requirements

### Requirement: Hierarchical agent tree with sub-agents
The system SHALL continue to support a hierarchical dynamic teammate path when `agent.multiAgent` is true. For built-in teammate production execution, the normal path SHALL enter through the control plane with `agent_spawn` rather than requiring static ADK specialist delegation. The static specialist tree remains available as a documented compatibility baseline where it already applies, and the production execution path SHALL continue to preserve parent-child session isolation and the canonical run identity chain.

The authoritative built-in teammate registry SHALL be `operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`. Spawn-time validation, prompt defaults, and role max scope for built-in teammates SHALL derive from that registry. Remote A2A agents remain a separate execution model.

#### Scenario: Built-in work uses spawn-only production execution
- **WHEN** the runtime routes built-in specialist work under multi-agent mode
- **THEN** the production path SHALL begin with `agent_spawn`
- **AND** built-in `transfer_to_agent` delegation SHALL NOT be required as the normal production path

#### Scenario: Parent and child sessions remain isolated
- **WHEN** a teammate is spawned from an active parent run
- **THEN** the child execution SHALL run in its own `ChildSession`
- **AND** the parent session SHALL remain the submission and observation anchor

#### Scenario: Legacy static fallback remains available
- **WHEN** the dynamic teammate runtime is unavailable or a legacy static path is already active
- **THEN** the system MAY continue through the existing static specialist routing behavior

#### Scenario: Remote A2A remains separate
- **WHEN** a configured remote A2A agent is selected
- **THEN** the runtime MAY still use the remote compatibility path
- **AND** this SHALL NOT re-open built-in static delegation as the normal production path

#### Scenario: Built-in teammate registry is authoritative
- **WHEN** the runtime resolves a built-in teammate type for spawn or policy evaluation
- **THEN** it SHALL use the authoritative built-in registry of `operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`
