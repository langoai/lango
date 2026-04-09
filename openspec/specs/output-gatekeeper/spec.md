## Purpose

Capability spec for output-gatekeeper. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Tool output size management
The system SHALL manage tool output size using token-based tiered compression via `WithOutputManager` middleware instead of character-based truncation. The middleware SHALL classify outputs into Small/Medium/Large tiers and apply content-aware compression.

#### Scenario: Output within budget
- **WHEN** a tool returns output within the token budget
- **THEN** the output SHALL pass through with `_meta` metadata injected

#### Scenario: Output exceeding budget
- **WHEN** a tool returns output exceeding the token budget
- **THEN** the output SHALL be compressed using content-type-specific strategies and `_meta.compressed` SHALL be `true`

### Requirement: Response sanitization
The system SHALL sanitize model responses before delivering them to users. Sanitization SHALL remove internal content that is not intended for end users. Each sanitization rule SHALL be independently configurable via `*bool` toggle (nil defaults to enabled).

#### Scenario: Thought tag removal
- **WHEN** the model response contains `<thought>...</thought>` or `<thinking>...</thinking>` blocks
- **THEN** those blocks SHALL be removed from the response

#### Scenario: Code block preservation
- **WHEN** the model response contains thought/thinking tags inside a code block (``` delimiters)
- **THEN** the tags inside code blocks SHALL NOT be removed

#### Scenario: Internal marker removal
- **WHEN** the model response contains lines starting with `[INTERNAL]`, `[DEBUG]`, `[SYSTEM]`, or `[OBSERVATION]`
- **THEN** those lines SHALL be removed from the response

#### Scenario: Large JSON block replacement
- **WHEN** the model response contains a JSON code block exceeding the rawJsonThreshold (default 500 characters)
- **THEN** the code block SHALL be replaced with `[Large data block omitted]`

#### Scenario: Small JSON block preservation
- **WHEN** the model response contains a JSON code block under the rawJsonThreshold
- **THEN** the code block SHALL be preserved unchanged

#### Scenario: Custom pattern application
- **WHEN** custom regex patterns are configured in `gatekeeper.customPatterns`
- **THEN** matching content SHALL be removed from the response

#### Scenario: Blank line normalization
- **WHEN** the response contains three or more consecutive newlines
- **THEN** they SHALL be collapsed to exactly two newlines

#### Scenario: Disabled sanitizer passthrough
- **WHEN** the gatekeeper is disabled (`gatekeeper.enabled: false`)
- **THEN** the response SHALL pass through unchanged

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

### Requirement: Gateway response sanitization
The system SHALL apply sanitization to both streaming chunks and final responses in the gateway server. Sanitization SHALL be applied only when a sanitizer is configured and enabled.

#### Scenario: Chunk sanitization
- **WHEN** the agent produces a streaming chunk and a sanitizer is configured
- **THEN** the chunk SHALL be sanitized before broadcasting to WebSocket clients

#### Scenario: Empty chunk suppression
- **WHEN** a chunk becomes empty after sanitization
- **THEN** the empty chunk SHALL NOT be broadcast to clients

#### Scenario: Final response sanitization
- **WHEN** the agent completes and returns a final response with a sanitizer configured
- **THEN** the response SHALL be sanitized before returning to the caller

### Requirement: Channel response sanitization
The system SHALL apply sanitization to agent responses in all channel handlers (Telegram, Discord, Slack). Sanitization SHALL be applied after the agent run completes and before the response is returned to the channel adapter.

#### Scenario: Channel response filtering
- **WHEN** the agent returns a response via runAgent() and a sanitizer is configured and enabled
- **THEN** the response SHALL be sanitized before returning to the channel handler
