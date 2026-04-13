## MODIFIED Requirements

### Requirement: Tool selection priority in prompts
The TOOL_USAGE.md prompt SHALL include a "Tool Selection Priority" section that instructs agents to always prefer built-in tools over skills. The section SHALL state that skills wrapping `lango` CLI commands will fail due to passphrase authentication requirements in agent mode.

#### Scenario: Agent reads tool usage prompt
- **WHEN** the agent processes TOOL_USAGE.md during system prompt assembly
- **THEN** the prompt SHALL contain a "Tool Selection Priority" section before the "Exec Tool" section

#### Scenario: Agent encounters a skill with built-in equivalent
- **WHEN** a skill provides functionality already available as a built-in tool
- **THEN** the prompt guidance SHALL direct the agent to use the built-in tool instead

#### Scenario: Approval failure guidance for browser actions
- **WHEN** the agent processes browser guidance in TOOL_USAGE.md or navigator-specific instructions
- **THEN** it SHALL be instructed not to immediately reissue the same browser action after approval denial or expiry
- **AND** it SHALL instead explain the approval issue or choose a lower-risk alternative only when that alternative materially changes the action
