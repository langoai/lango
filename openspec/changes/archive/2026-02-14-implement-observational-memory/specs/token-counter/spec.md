## ADDED Requirements

### Requirement: Token Count Estimation
The system SHALL provide approximate token counting for text content.

#### Scenario: ASCII text estimation
- **WHEN** counting tokens for ASCII/Latin text
- **THEN** the system SHALL estimate 1 token per 4 characters

#### Scenario: CJK text estimation
- **WHEN** counting tokens for CJK (Chinese, Japanese, Korean) text
- **THEN** the system SHALL estimate 1 token per 2 characters

#### Scenario: Mixed text estimation
- **WHEN** counting tokens for text containing both ASCII and CJK characters
- **THEN** the system SHALL count each character segment with the appropriate ratio and sum the results

#### Scenario: Empty text
- **WHEN** counting tokens for empty text
- **THEN** the system SHALL return 0

### Requirement: Message Token Counting
The system SHALL count tokens for session messages including all content.

#### Scenario: Text message token count
- **WHEN** counting tokens for a message with text content
- **THEN** the system SHALL count tokens for the content string plus a role overhead (estimated 4 tokens per message for role/formatting)

#### Scenario: Tool call message token count
- **WHEN** counting tokens for a message with tool calls
- **THEN** the system SHALL count tokens for the content plus the serialized tool call input and output

#### Scenario: Batch message token count
- **WHEN** counting tokens for a slice of messages
- **THEN** the system SHALL return the sum of individual message token counts
