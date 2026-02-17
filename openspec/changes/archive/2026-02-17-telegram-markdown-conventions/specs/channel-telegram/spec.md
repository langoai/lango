## ADDED Requirements

### Requirement: Channel-safe Markdown formatting via agent prompt
The system SHALL include Markdown formatting conventions in the agent's conversation prompt that prevent Telegram Markdown v1 parse errors. The prompt rules SHALL instruct the LLM to avoid patterns that produce malformed Telegram markup.

#### Scenario: Bullet lists use dash syntax
- **WHEN** the agent generates an unordered list
- **THEN** the agent SHALL use `-` as the list marker, never `*`

#### Scenario: Code identifiers wrapped in backticks
- **WHEN** the agent references variable names, function names, or file paths containing underscores
- **THEN** the agent SHALL wrap them in backtick code spans to prevent underscore-as-italic parsing

#### Scenario: All formatting markers are paired
- **WHEN** the agent uses `*`, `_`, or `` ` `` markers
- **THEN** every opening marker SHALL have a corresponding closing marker

#### Scenario: Fenced code blocks are closed
- **WHEN** the agent opens a fenced code block with ` ``` `
- **THEN** the agent SHALL always include a closing ` ``` ` marker
