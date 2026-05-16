## MODIFIED Requirements

### Requirement: Telegram media download completes successfully
The system SHALL download file content from Telegram's file API via HTTP GET with a 30-second timeout and return the raw bytes.

#### Scenario: File body read failure preserves the I/O cause
- **WHEN** the Telegram file API returns a 200 response but reading the body fails
- **THEN** the system returns an error identifying the body-read failure
- **AND** SHALL preserve the underlying I/O cause instead of collapsing into the empty-body path
