## MODIFIED Requirements

### Requirement: Anthropic provider unknown role handling
The `convertParams` method SHALL handle unknown message roles by logging a warning and skipping the message. The switch statement SHALL include explicit cases for "user", "assistant", and "system" roles, and a `default` case that logs the unknown role via the subsystem logger.

#### Scenario: Unknown role is logged and skipped
- **WHEN** `convertParams` encounters a message with role "tool" or any unrecognized role
- **THEN** it SHALL log a warning containing the unknown role value
- **AND** it SHALL NOT include that message in the Anthropic API request
- **AND** it SHALL NOT return an error

#### Scenario: System role handled separately
- **WHEN** `convertParams` encounters a message with role "system"
- **THEN** it SHALL NOT log a warning (system is handled in a separate loop)
