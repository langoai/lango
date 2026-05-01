## MODIFIED Requirements

### Requirement: Hierarchical agent tree with sub-agents

The system SHALL support a hierarchical dynamic teammate path when `agent.multiAgent` is true. The static specialist tree remains available as a compatibility baseline, but the production execution path SHALL allow the runtime to spawn built-in teammate types in-process through the control plane while preserving parent-child session isolation and the canonical run identity chain.

The canonical built-in teammate type registry SHALL be `operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`. Spawn-time validation, prompt defaults, and role max scope for built-in teammates SHALL derive from that registry.

#### Scenario: Dynamic teammate created under multi-agent mode
- **WHEN** multi-agent mode is enabled and the runtime needs a specialist execution path
- **THEN** the system SHALL be able to create an in-process teammate run through the control plane instead of requiring a predeclared static delegation-only hop

#### Scenario: Parent and child sessions remain isolated
- **WHEN** a teammate is spawned from an active parent run
- **THEN** the child execution SHALL run in its own `ChildSession`
- **AND** the parent session SHALL remain the submission and observation anchor

#### Scenario: Legacy static fallback remains available
- **WHEN** the dynamic teammate runtime is unavailable or a legacy static path is already active
- **THEN** the system MAY continue through the existing static specialist routing behavior

#### Scenario: Built-in teammate registry is authoritative
- **WHEN** the runtime resolves a built-in teammate type for spawn or policy evaluation
- **THEN** it SHALL use the canonical built-in registry of `operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`

### Requirement: Tool partitioning by prefix

Tool prefix partitioning SHALL define role maximum scope and default teammate affinity, not a fixed execution authority. Spawn-time `allowed_tools` may narrow the role scope, but the runtime SHALL reject any request that attempts to widen beyond the role max scope.

#### Scenario: Prefix partition defines max scope
- **WHEN** a built-in teammate type resolves to the operator role
- **THEN** operator-prefixed tools SHALL define the maximum tool scope available for that teammate type

#### Scenario: Spawn-time allowlist narrows scope
- **WHEN** a caller spawns an operator teammate with `allowed_tools: ["fs_read"]`
- **THEN** the teammate SHALL run with only `fs_read` plus runtime essentials even if the operator role supports additional tools

#### Scenario: Spawn-time allowlist cannot widen scope
- **WHEN** a caller spawns a librarian teammate with `allowed_tools` containing an operator-only tool
- **THEN** the runtime SHALL reject the spawn request before execution begins

### Requirement: Orchestrator re-routing protocol

The orchestrator re-routing protocol SHALL be reframed as a recovery and synthesis policy. When a teammate or legacy specialist returns control, the root runtime SHALL re-evaluate the state, avoid immediate same-target repetition, and either spawn or route to a different eligible specialist, synthesize a direct answer, or preserve the current failure state for operator visibility.

#### Scenario: Re-routing avoids same teammate loop
- **WHEN** a teammate returns without resolving the request
- **THEN** the root runtime SHALL NOT immediately re-send the request to the same teammate without new recovery context

#### Scenario: Recovery synthesizes direct answer
- **WHEN** no eligible teammate remains after re-evaluation
- **THEN** the runtime SHALL be allowed to synthesize a direct response from the available state instead of forcing another handoff

#### Scenario: Recovery preserves observable failure state
- **WHEN** the runtime stops after a blocked or failed teammate path
- **THEN** the current recovery condition SHALL remain visible to operators through projected run state

### Requirement: Event Author Identity

Assistant-side event author identity SHALL preserve the runtime identity chain. Event authoring SHALL use the active teammate or root run identity projected from `AgentRun` and background execution state, falling back to the configured root identity only for legacy messages that lack stored author metadata.

#### Scenario: Spawned teammate emits authored events
- **WHEN** a spawned teammate produces assistant events during execution
- **THEN** those events SHALL use the teammate's projected runtime identity rather than a hardcoded orchestrator name

#### Scenario: Legacy assistant message falls back cleanly
- **WHEN** a legacy assistant message lacks stored author metadata
- **THEN** the system SHALL fall back to the configured root agent name

### Requirement: Remote agents as sub-agents

Remote A2A routing SHALL be preserved in v1. The dynamic teammate runtime applies to in-process built-in teammate types, while existing remote A2A sub-agent paths continue to use the current routing and compatibility behavior.

#### Scenario: Remote A2A path remains unchanged
- **WHEN** a request is routed to a configured remote A2A agent in v1
- **THEN** the system SHALL continue to use the existing remote routing behavior without forcing the in-process teammate runtime

#### Scenario: Built-in dynamic runtime and remote A2A coexist
- **WHEN** both built-in teammate types and remote A2A agents are configured
- **THEN** the runtime SHALL support the dynamic in-process path for built-ins and preserve the current remote path for A2A agents

#### Scenario: Dynamic runtime status reflects built-in availability only
- **WHEN** multi-agent mode is enabled and the built-in in-process teammate runtime path is configured
- **THEN** status surfaces MAY report `dynamic-v1` even if legacy fallback and remote A2A paths also remain available
