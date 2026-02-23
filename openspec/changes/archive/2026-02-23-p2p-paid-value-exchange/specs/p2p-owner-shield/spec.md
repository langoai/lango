## Purpose

Hard-block privacy layer that prevents owner PII from being leaked through P2P responses, regardless of payment amount.

## Requirements

### Requirement: PII Redaction
The system SHALL redact owner personal information from all P2P responses.

#### Scenario: Owner name in response
- **WHEN** a P2P response contains the configured owner name
- **THEN** the system replaces it with "[owner-data-redacted]"

#### Scenario: Email pattern in response
- **WHEN** a P2P response contains an email address matching the configured owner email or general email patterns
- **THEN** the system replaces it with "[owner-data-redacted]"

#### Scenario: Phone pattern in response
- **WHEN** a P2P response contains a phone number matching the configured owner phone or general phone patterns
- **THEN** the system replaces it with "[owner-data-redacted]"

### Requirement: Conversation Blocking
The system SHALL block conversation history fields from P2P responses by default.

#### Scenario: Conversation data in response
- **WHEN** a P2P response contains keys like "conversation", "message_history", "chat_log", "session_history", or "chat_history"
- **THEN** the system replaces the value with "[owner-data-redacted]"

#### Scenario: Conversation blocking disabled
- **WHEN** blockConversations is explicitly set to false
- **THEN** conversation fields are not redacted

### Requirement: Recursive Scanning
The system SHALL recursively scan nested maps and slices for owner data.

#### Scenario: Nested PII
- **WHEN** owner data appears in a deeply nested map within the response
- **THEN** the system detects and redacts it
