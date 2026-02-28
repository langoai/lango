## Requirement: System prompt construction
The system SHALL construct the system prompt using a structured `prompt.Builder` instead of a single string. The `ContextAwareModelAdapter` constructor SHALL accept a `*prompt.Builder` and call `Build()` to produce the base prompt string. Dynamic context injection (knowledge, memory, RAG) SHALL continue to append to the built prompt at runtime.

#### Scenario: Prepend system prompt to new session
- **WHEN** a new agent session is started
- **THEN** the system SHALL prepend a message with `role: system` containing the configured prompt to the conversation history

#### Scenario: Default identity prompt
- **WHEN** no custom system prompt is provided
- **THEN** a default prompt describing the agent's identity and tools SHALL be used

#### Scenario: Builder produces base prompt
- **WHEN** ContextAwareModelAdapter is created with a prompt.Builder
- **THEN** the basePrompt field SHALL equal the builder's Build() output

#### Scenario: Dynamic context still appended
- **WHEN** knowledge retrieval returns context during GenerateContent
- **THEN** the retrieved context SHALL be appended to the builder-produced base prompt

### Requirement: SAFETY prompt reflects PII detection scope
The SAFETY.md prompt SHALL enumerate specific PII categories (email, phone numbers, national IDs, financial account numbers) and mention 13 builtin patterns. The prompt SHALL reference Presidio NER-based detection as additional coverage when enabled.

#### Scenario: Agent responds to PII-related user question
- **WHEN** the agent processes SAFETY.md prompt during system prompt assembly
- **THEN** the agent understands it protects 13 builtin PII pattern categories
- **THEN** the agent can accurately inform users about PII protection coverage including Presidio

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
