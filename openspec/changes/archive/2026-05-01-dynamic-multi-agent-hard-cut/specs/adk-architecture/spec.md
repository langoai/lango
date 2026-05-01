## MODIFIED Requirements

### Requirement: Agent hallucination retry in RunAndCollect
`RunAndCollect` SHALL continue to detect `"failed to find agent"` errors, send a correction message, and retry exactly once when sub-agents are registered. The correction message SHALL continue to exist, but the hard cut narrows its built-in semantics: remote/legacy transfer recovery MAY still name compatible transfer targets, while built-in teammate recovery SHALL steer the root runtime toward `agent_spawn` or direct root answer behavior instead of teaching built-in static transfer retries.

#### Scenario: Hallucinated agent name triggers retry
- **WHEN** a `RunAndCollect` call yields an error matching `"failed to find agent: <name>"`
- **AND** the agent has sub-agents registered
- **THEN** the system SHALL send a correction message for the invalid target
- **AND** retry the run exactly once with that correction message

#### Scenario: Built-in hallucinated target produces spawn-oriented recovery
- **WHEN** built-in routing fails with a hallucinated target name
- **AND** the agent has sub-agents registered
- **THEN** the correction message SHALL steer the root runtime toward `agent_spawn` or direct root answer behavior
- **AND** it SHALL NOT suggest retrying a built-in `transfer_to_agent` target

#### Scenario: Retry succeeds
- **WHEN** the correction message retry produces a successful response
- **THEN** `RunAndCollect` SHALL return the successful response with no error

#### Scenario: Retry also fails
- **WHEN** the correction message retry also produces an error
- **THEN** `RunAndCollect` SHALL return the retry error

#### Scenario: Non-hallucination error is not retried
- **WHEN** `RunAndCollect` encounters an error that does not match `"failed to find agent"`
- **THEN** the error SHALL be returned immediately without retry

#### Scenario: No sub-agents means no retry
- **WHEN** `RunAndCollect` encounters a `"failed to find agent"` error
- **AND** the agent has no sub-agents
- **THEN** the error SHALL be returned immediately without retry
