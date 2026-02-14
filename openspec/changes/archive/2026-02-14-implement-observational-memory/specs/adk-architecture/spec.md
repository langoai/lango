## MODIFIED Requirements

### Requirement: History Management
The system SHALL manage session history using token-budget-based dynamic truncation to prevent context overflow and optimize token usage.

#### Scenario: History Truncation
- **WHEN** loading session history for the agent
- **THEN** a token budget (configurable via `maxMessageTokenBudget`, default 8000) SHALL be applied
- **AND** messages SHALL be included from most recent to oldest until the budget is exhausted
- **AND** any remaining older messages SHALL be excluded from the LLM context

#### Scenario: Fallback to message count
- **WHEN** Observational Memory is disabled
- **THEN** the system SHALL fall back to the existing hard message count limit (100 messages)

#### Scenario: Event Author Mapping
- **WHEN** adapting historical messages to ADK events
- **THEN** the `Author` field SHALL be populated based on the message role
- **AND** `user` role maps to `user` author
- **AND** `assistant` role maps to the agent name
