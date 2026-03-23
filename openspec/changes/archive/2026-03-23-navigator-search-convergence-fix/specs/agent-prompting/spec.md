## MODIFIED Requirements

### Requirement: Tool selection priority in prompts
The TOOL_USAGE.md prompt SHALL include a "Tool Selection Priority" section that instructs agents to always prefer built-in tools over skills. The section SHALL state that skills wrapping `lango` CLI commands will fail due to passphrase authentication requirements in agent mode.

#### Scenario: Agent reads tool usage prompt
- **WHEN** the agent processes TOOL_USAGE.md during system prompt assembly
- **THEN** the prompt SHALL contain a "Tool Selection Priority" section before the "Exec Tool" section

#### Scenario: Agent encounters a skill with built-in equivalent
- **WHEN** a skill provides functionality already available as a built-in tool
- **THEN** the prompt guidance SHALL direct the agent to use the built-in tool instead

#### Scenario: Bounded browser search guidance
- **WHEN** the agent processes browser guidance in TOOL_USAGE.md
- **THEN** it SHALL be instructed to run `browser_search` once for topic queries
- **AND** it SHALL be instructed to use the current page with `browser_extract(search_results)` before issuing another search
- **AND** it SHALL be instructed to reformulate at most once when the first results are empty or clearly unrelated
- **AND** it SHALL be instructed to stop once the requested number of credible results has been collected
