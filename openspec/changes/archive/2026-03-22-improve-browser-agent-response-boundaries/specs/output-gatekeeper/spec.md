## MODIFIED Requirements

### Requirement: System prompt output principles
The system SHALL include an "Output Principles" section in the system prompt at priority 350 (between Conversation Rules at 300 and Tool Usage at 400). The section SHALL instruct the model to never echo raw tool output, keep internal reasoning internal, summarize large results, avoid system markers, present structured data in natural language, explain errors without full stack traces, and never emit role-labeled prompt/tool dumps such as system prompt, user prompt, assistant response, or tool output blocks in final user-visible replies.

#### Scenario: Output principles in system prompt
- **WHEN** the system prompt is built using DefaultBuilder
- **THEN** the prompt SHALL contain an "Output Principles" section with instructions about output behavior

#### Scenario: Priority ordering
- **WHEN** the system prompt is rendered
- **THEN** Output Principles SHALL appear after Conversation Rules and before Tool Usage Guidelines

#### Scenario: File override support
- **WHEN** a custom `OUTPUT_PRINCIPLES.md` file exists in the prompts directory
- **THEN** it SHALL override the default embedded output principles content
