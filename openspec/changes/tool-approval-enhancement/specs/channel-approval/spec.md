## MODIFIED Requirements

### Requirement: Approval request context
Each approval request SHALL carry an ID, tool name, session key, parameters, a human-readable Summary string, and creation timestamp.

#### Scenario: Request fields
- **WHEN** an approval request is created
- **THEN** it SHALL contain a unique ID, the tool name, the originating session key, tool parameters, a Summary string, and a timestamp

#### Scenario: Summary populated
- **WHEN** a tool approval request is created via wrapWithApproval
- **THEN** the Summary field SHALL be populated by buildApprovalSummary with a human-readable description of the operation

#### Scenario: Empty summary backward compatibility
- **WHEN** an approval request has an empty Summary
- **THEN** providers SHALL display the existing tool-name-only message

## ADDED Requirements

### Requirement: Approval summary rendering
All approval providers SHALL include the Summary field in their approval messages when it is non-empty.

#### Scenario: Gateway provider summary
- **WHEN** a GatewayProvider receives a request with Summary "Execute: ls -la"
- **THEN** the message sent to companions SHALL include the Summary text

#### Scenario: TTY provider summary
- **WHEN** a TTYProvider receives a request with Summary "Delete: /tmp/test"
- **THEN** the terminal prompt SHALL display the Summary on a separate line before the y/N prompt

#### Scenario: Headless provider summary
- **WHEN** a HeadlessProvider receives a request with Summary
- **THEN** the audit log entry SHALL include a "summary" field

#### Scenario: Telegram provider summary
- **WHEN** a Telegram ApprovalProvider receives a request with Summary
- **THEN** the InlineKeyboard message SHALL include the Summary text below the tool name

#### Scenario: Discord provider summary
- **WHEN** a Discord ApprovalProvider receives a request with Summary
- **THEN** the button message Content SHALL include the Summary in a code block

#### Scenario: Slack provider summary
- **WHEN** a Slack ApprovalProvider receives a request with Summary
- **THEN** the Block Kit section text SHALL include the Summary in a code block

### Requirement: Approval summary builder
The system SHALL provide a `buildApprovalSummary(toolName, params)` function that generates human-readable descriptions of tool invocations.

#### Scenario: Exec tool summary
- **WHEN** buildApprovalSummary is called with toolName "exec" and params containing command "curl https://api.example.com"
- **THEN** it SHALL return "Execute: curl https://api.example.com"

#### Scenario: File write summary
- **WHEN** buildApprovalSummary is called with toolName "fs_write" and params containing path "/tmp/test.txt" and content of 100 bytes
- **THEN** it SHALL return "Write to /tmp/test.txt (100 bytes)"

#### Scenario: Unknown tool summary
- **WHEN** buildApprovalSummary is called with an unrecognized toolName
- **THEN** it SHALL return "Tool: <toolName>"

#### Scenario: Long command truncation
- **WHEN** a command string exceeds 200 characters
- **THEN** it SHALL be truncated to 200 characters with "..." appended
