## ADDED Requirements

### Requirement: Message Author Field
The `session.Message` struct SHALL include an `Author string` field (JSON tag `"author,omitempty"`) to store the ADK agent name that produced the message.

#### Scenario: Author preserved through AppendEvent
- **WHEN** an ADK event with `Author: "lango-orchestrator"` is appended
- **THEN** the stored message SHALL have `Author: "lango-orchestrator"`

#### Scenario: Author loaded from storage
- **WHEN** a session is loaded from the ent store
- **THEN** each message's Author field SHALL be populated from the stored `author` column
