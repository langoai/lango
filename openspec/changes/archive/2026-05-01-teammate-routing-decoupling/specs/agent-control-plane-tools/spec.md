## MODIFIED Requirements

### Requirement: agent_spawn tool creates AgentRun with enriched prompt and advisory routing

The existing `agent_spawn` response shape and basic ID semantics remain preserved unless explicitly changed by this requirement. Advisory teammate routing SHALL now be carried by `RequestedAgent` metadata rather than by rewriting the stored `Instruction` with an advisory system prefix. The stored `Instruction` SHALL remain the raw user instruction text.

Built-in teammate types SHALL still validate `allowed_tools` against their role maximum scope before execution begins. Custom or non-built-in teammate paths SHALL continue to accept spawn-time `allowed_tools`, but their effective runtime ceiling is defined by the current allowlist rather than a built-in role registry.

#### Scenario: Requested agent remains advisory metadata
- **WHEN** `agent_spawn` is called with `agent: "researcher"`
- **THEN** the created run SHALL persist `RequestedAgent: "researcher"`
- **AND** no code-level enforcement SHALL guarantee routing to `"researcher"`

#### Scenario: Stored instruction remains raw
- **WHEN** `agent_spawn` is called with `instruction: "fix the bug"` and `agent: "researcher"`
- **THEN** the stored `Instruction` SHALL remain exactly `"fix the bug"`
- **AND** the advisory routing signal SHALL be carried outside the raw instruction text

### Requirement: RequestedAgent routing is advisory via enriched prompt

`RequestedAgent` SHALL remain advisory, but the advisory signal is now expressed through metadata and runtime context rather than through an enriched prompt prefix stored in `AgentRun.Instruction`.

#### Scenario: No advisory prefix is stored in AgentRun instruction
- **WHEN** `agent_spawn` is called with an `agent` parameter
- **THEN** the stored `Instruction` SHALL NOT gain a synthetic `"[System: This task is best handled by ...]"` prefix
- **AND** the runtime MAY still use `RequestedAgent` metadata for routing and teammate context
