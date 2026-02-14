
## ADDED Requirements

### Requirement: ADK Agent Abstraction
The system SHALL wrap the Google ADK Agent to integrate with the application.

#### Scenario: Agent Initialization
- **WHEN** the server starts
- **THEN** it SHALL initialize an ADK Agent instance
- **AND** configure it with the selected model and tools from the configuration

### Requirement: Ent State Adapter
The system SHALL adapt the Ent-based session store to the ADK State interface.

#### Scenario: Load Session
- **WHEN** ADK requests state for a session ID
- **THEN** the adapter SHALL retrieve the session and messages from Ent
- **AND** map them to ADK's expected message format

#### Scenario: Save Session
- **WHEN** ADK persists state updates
- **THEN** the adapter SHALL save new messages and state to Ent

### Requirement: Tool Adapter
The system SHALL adapt existing internal tools to the ADK Tool interface.

#### Scenario: Execute Legacy Tool
- **WHEN** ADK invokes a tool
- **THEN** the adapter SHALL translate the inputs and call the internal tool implementation
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
