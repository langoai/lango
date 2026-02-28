## ADDED Requirements

### Requirement: Tool selection priority in prompts
The TOOL_USAGE.md prompt SHALL include a "Tool Selection Priority" section that instructs agents to always prefer built-in tools over skills. The section SHALL state that skills wrapping `lango` CLI commands will fail due to passphrase authentication requirements in agent mode.

#### Scenario: Agent reads tool usage prompt
- **WHEN** the agent processes TOOL_USAGE.md during system prompt assembly
- **THEN** the prompt SHALL contain a "Tool Selection Priority" section before the "Exec Tool" section

#### Scenario: Agent encounters a skill with built-in equivalent
- **WHEN** a skill provides functionality already available as a built-in tool
- **THEN** the prompt guidance SHALL direct the agent to use the built-in tool instead

### Requirement: Tool selection directive in agent identity
The AGENTS.md prompt SHALL include a tool selection directive stating that built-in tools MUST be preferred over skills, and skills are extensions for specialized use cases only.

#### Scenario: Agent reads identity prompt
- **WHEN** the agent processes AGENTS.md during system prompt assembly
- **THEN** the prompt SHALL contain a tool selection directive before the knowledge system description

### Requirement: Runtime skill priority note
The `AssemblePrompt()` method in `ContextRetriever` SHALL prepend a note to the "Available Skills" section advising agents to prefer built-in tools over skills.

#### Scenario: Skills section rendered with priority note
- **WHEN** the assembled prompt includes skill pattern items
- **THEN** the "Available Skills" section SHALL begin with a note stating to prefer built-in tools over skills
