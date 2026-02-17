## MODIFIED Requirements

### Requirement: CONVERSATION_RULES.md prevents answer repetition
The `CONVERSATION_RULES.md` file SHALL define conversation behavior rules that prevent answer accumulation/repetition across turns, including current-question focus, self-contained responses, topic change handling, channel message limits, consistency maintenance, and channel-safe Markdown formatting conventions.

#### Scenario: Conversation rules prevent repetition
- **WHEN** the conversation rules section is rendered
- **THEN** it SHALL instruct to answer only the current question and not repeat previous answers

#### Scenario: Conversation rules respect channel limits
- **WHEN** the conversation rules section is rendered
- **THEN** it SHALL mention Telegram (4096) and Discord (2000) character limits

#### Scenario: Conversation rules include Markdown formatting conventions
- **WHEN** the conversation rules section is rendered
- **THEN** it SHALL instruct to use `-` for unordered lists instead of `*`
- **AND** it SHALL instruct to wrap code identifiers in backticks
- **AND** it SHALL instruct to ensure all formatting markers are paired
- **AND** it SHALL instruct to always close fenced code blocks
