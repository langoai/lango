## MODIFIED Requirements

### Requirement: Agent configuration supports prompts directory
The `AgentConfig` struct SHALL include a `PromptsDir` field (mapstructure: "promptsDir") specifying the directory containing section `.md` files. The system SHALL support three-tier precedence: PromptsDir > SystemPromptPath > built-in defaults.

#### Scenario: PromptsDir configured
- **WHEN** AgentConfig.PromptsDir is set to a valid directory path
- **THEN** the system SHALL load prompt sections from .md files in that directory

#### Scenario: Legacy SystemPromptPath only
- **WHEN** AgentConfig.PromptsDir is empty but SystemPromptPath is set
- **THEN** the file content SHALL replace the Identity section only, and all other default sections SHALL remain

#### Scenario: No prompt configuration
- **WHEN** both PromptsDir and SystemPromptPath are empty
- **THEN** the system SHALL use the built-in default sections including conversation rules
