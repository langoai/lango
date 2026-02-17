## ADDED Requirements

### Requirement: SubAgentPromptFunc type
The orchestration package SHALL define a `SubAgentPromptFunc` function type that takes `(agentName, defaultInstruction string)` and returns the assembled system prompt string for a sub-agent.

#### Scenario: Function receives correct parameters
- **WHEN** `BuildAgentTree` calls the `SubAgentPromptFunc` for each sub-agent
- **THEN** it SHALL pass the agent's spec name and the original spec.Instruction

## MODIFIED Requirements

### Requirement: Config supports SubAgentPrompt field
The orchestration `Config` struct SHALL include a `SubAgentPrompt SubAgentPromptFunc` field. When set, `BuildAgentTree` SHALL use it to build each sub-agent's instruction. When nil, the original `spec.Instruction` is used.

#### Scenario: SubAgentPrompt set
- **WHEN** `Config.SubAgentPrompt` is non-nil
- **THEN** `BuildAgentTree` SHALL call it for every sub-agent and use the returned string as the agent's Instruction

#### Scenario: SubAgentPrompt nil (backward compatible)
- **WHEN** `Config.SubAgentPrompt` is nil
- **THEN** `BuildAgentTree` SHALL use `spec.Instruction` directly, preserving existing behavior
