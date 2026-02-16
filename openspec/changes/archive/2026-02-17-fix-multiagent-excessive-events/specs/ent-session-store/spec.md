## ADDED Requirements

### Requirement: Message Author Field in Ent Schema
The Message ent schema SHALL include an optional `author` string field with a default empty value, used to persist the ADK agent name for multi-agent routing.

#### Scenario: New message with author
- **WHEN** a message is saved with a non-empty Author
- **THEN** the `author` column SHALL store the agent name

#### Scenario: Legacy message without author
- **WHEN** an existing message has no author value
- **THEN** the `author` column SHALL default to empty string
