## MODIFIED Requirements

### Requirement: System prompt construction
The system SHALL construct the system prompt using a structured `prompt.Builder` instead of a single string. The `ContextAwareModelAdapter` constructor SHALL accept a `*prompt.Builder` and call `Build()` to produce the base prompt string. Dynamic context injection (knowledge, memory, RAG) SHALL continue to append to the built prompt at runtime.

#### Scenario: Builder produces base prompt
- **WHEN** ContextAwareModelAdapter is created with a prompt.Builder
- **THEN** the basePrompt field SHALL equal the builder's Build() output

#### Scenario: Dynamic context still appended
- **WHEN** knowledge retrieval returns context during GenerateContent
- **THEN** the retrieved context SHALL be appended to the builder-produced base prompt
