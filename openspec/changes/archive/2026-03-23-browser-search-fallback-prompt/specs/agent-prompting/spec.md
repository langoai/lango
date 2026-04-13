## MODIFIED Requirements

### Requirement: Tool selection priority in prompts
The TOOL_USAGE.md prompt SHALL include a "Tool Selection Priority" section that instructs agents to always prefer built-in tools over skills. The section SHALL state that skills wrapping `lango` CLI commands will fail due to passphrase authentication requirements in agent mode.

#### Scenario: Agent reads tool usage prompt
- **WHEN** the agent processes TOOL_USAGE.md during system prompt assembly
- **THEN** the prompt SHALL contain a "Tool Selection Priority" section before the "Exec Tool" section

#### Scenario: Agent encounters a skill with built-in equivalent
- **WHEN** a skill provides functionality already available as a built-in tool
- **THEN** the prompt guidance SHALL direct the agent to use the built-in tool instead

#### Scenario: Browser search fallback guidance
- **WHEN** the agent processes the browser section of TOOL_USAGE.md
- **THEN** it SHALL be instructed to prefer `browser_search` for open-ended live web queries
- **AND** it SHALL also be instructed to fall back to `browser_navigate` with a search URL plus `browser_extract` in `search_results` mode when `browser_search` is unavailable
- **AND** it SHALL be instructed to continue with low-level `browser_action` or `eval` instead of stopping if equivalent browser tools are still available
