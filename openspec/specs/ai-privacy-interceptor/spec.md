## Purpose

Define the AI Privacy Interceptor that filters agent requests to the LLM, providing PII redaction, tool approval workflows, and output secret scanning.
## Requirements
### Requirement: Intercept AI Requests
The system SHALL intercept all agent requests to the LLM and pass them through a filter chain before execution.

#### Scenario: Interception
- **WHEN** an agent initiates a chat completion request
- **THEN** the request parameters (messages, tools) are inspected by the middleware

### Requirement: PII Redaction
The system SHALL automatically redact personally identifiable information (email, phone numbers, API keys) from the prompt content using regex patterns.

#### Scenario: Redact Email
- **WHEN** a user prompt contains "my email is test@example.com"
- **THEN** it is rewritten to "my email is [REDACTED]" before sending to the LLM

### Requirement: Extended PII pattern detection
PIIRedactor SHALL use a PIIDetector interface with 13 builtin regex patterns across contact, identity, financial, and network categories. PIIConfig SHALL support legacy fields (RedactEmail, RedactPhone, CustomRegex) and new fields (DisabledBuiltins, CustomPatterns, PresidioEnabled, PresidioURL, PresidioThreshold, PresidioLanguage). A helper function SHALL list all currently enabled builtin pattern names for diagnostics.

#### Scenario: Korean mobile number redaction
- **WHEN** a user prompt contains "phone: 010-1234-5678"
- **THEN** the phone number SHALL be replaced with [REDACTED]

#### Scenario: Korean RRN redaction
- **WHEN** a user prompt contains "rrn: 900101-1234567"
- **THEN** the RRN SHALL be replaced with [REDACTED]

#### Scenario: Disabled builtins
- **WHEN** PIIConfig has DisabledBuiltins=["email"]
- **THEN** PIIRedactor SHALL not detect email addresses

#### Scenario: Custom named patterns
- **WHEN** PIIConfig has CustomPatterns={"proj_id": "\\bPROJ-\\d{4}\\b"}
- **THEN** PIIRedactor SHALL detect matching text

#### Scenario: Presidio enabled
- **WHEN** PIIConfig has PresidioEnabled=true and PresidioURL set
- **THEN** PIIRedactor SHALL create a CompositeDetector with both RegexDetector and PresidioDetector

#### Scenario: List enabled patterns
- **WHEN** the diagnostic helper is called
- **THEN** it SHALL return the names of all currently enabled builtin patterns

### Requirement: Position-based redaction
PIIRedactor.RedactInput SHALL use match position offsets to replace detected PII, merging overlapping matches into single [REDACTED] markers.

#### Scenario: Non-overlapping matches
- **WHEN** text contains email and phone at separate positions
- **THEN** each SHALL be replaced with [REDACTED] independently

#### Scenario: Overlapping matches
- **WHEN** two patterns match overlapping text regions
- **THEN** they SHALL be merged into a single [REDACTED] replacement

#### Scenario: No matches
- **WHEN** text contains no PII
- **THEN** RedactInput SHALL return the original text unchanged

### Requirement: README documents PII detection capabilities
The README AI Privacy Interceptor section SHALL describe all 13 builtin PII detection patterns organized by category (Contact, Identity, Financial, Network). The section SHALL document pattern customization via `piiDisabledPatterns` and `piiCustomPatterns`. The section SHALL document optional Presidio NER-based detection integration.

#### Scenario: User reads AI Privacy Interceptor section
- **WHEN** a user reads the AI Privacy Interceptor section in README.md
- **THEN** they see the 4 pattern categories with specific pattern names listed
- **THEN** they see how to customize patterns (disable builtin, add custom)
- **THEN** they see how to enable Presidio integration with Docker Compose

### Requirement: README configuration table includes PII fields
The README configuration reference table SHALL include rows for `piiDisabledPatterns`, `piiCustomPatterns`, `presidio.enabled`, `presidio.url`, `presidio.scoreThreshold`, and `presidio.language` with correct types and defaults.

#### Scenario: User looks up PII config fields
- **WHEN** a user searches the configuration reference table for PII settings
- **THEN** they find 6 new rows after `piiRegexPatterns` with type, default, and description for each field

### Requirement: Approval Workflow
The system SHALL determine which tools require approval based on the configured `ApprovalPolicy` and each tool's `SafetyLevel`. The system SHALL use fail-closed semantics: without explicit approval, execution is denied. Approval requests SHALL be routed through a `CompositeProvider` that selects the appropriate approval channel based on the originating session key. The legacy `approvalRequired` boolean is deprecated in favor of `ApprovalPolicy`.

#### Scenario: Policy-based approval gate
- **WHEN** the application initializes tools
- **AND** ApprovalPolicy is not "none"
- **THEN** each tool SHALL be passed through wrapWithApproval which uses needsApproval to determine if wrapping is needed

#### Scenario: None policy disables approval
- **WHEN** ApprovalPolicy is "none"
- **THEN** no tools SHALL be wrapped with approval logic

#### Scenario: Default policy for empty config
- **WHEN** ApprovalPolicy is empty (not set)
- **THEN** the system SHALL treat it as "dangerous"

#### Scenario: Channel-specific approval (Telegram)
- **WHEN** the agent attempts to call a tool that requires approval
- **AND** the request originates from a Telegram session (session key starts with "telegram:")
- **THEN** the approval request SHALL be routed to the Telegram approval provider

#### Scenario: Channel-specific approval (Discord)
- **WHEN** the agent attempts to call a tool that requires approval
- **AND** the request originates from a Discord session (session key starts with "discord:")
- **THEN** the approval request SHALL be routed to the Discord approval provider

#### Scenario: Channel-specific approval (Slack)
- **WHEN** the agent attempts to call a tool that requires approval
- **AND** the request originates from a Slack session (session key starts with "slack:")
- **THEN** the approval request SHALL be routed to the Slack approval provider

#### Scenario: Companion approval granted
- **WHEN** the agent attempts to call a tool that requires approval
- **AND** no channel-specific provider matches
- **AND** a companion is connected
- **AND** the companion approves the request
- **THEN** the tool execution SHALL proceed

#### Scenario: Companion approval denied
- **WHEN** the agent attempts to call a tool that requires approval
- **AND** no channel-specific provider matches
- **AND** a companion is connected
- **AND** the companion denies the request
- **THEN** the tool execution SHALL be denied with an error

#### Scenario: Companion approval error
- **WHEN** the agent attempts to call a tool that requires approval
- **AND** no channel-specific provider matches
- **AND** a companion is connected
- **AND** the approval request encounters an error
- **THEN** the tool execution SHALL be denied (fail-closed)

#### Scenario: TTY fallback approval
- **WHEN** the agent attempts to call a tool that requires approval
- **AND** no channel-specific provider matches
- **AND** no companion is connected
- **AND** `HeadlessAutoApprove` is false or not set
- **AND** stdin is a terminal (TTY)
- **THEN** the system SHALL prompt the user via stderr with tool name, summary, and "Allow? [y/N]"
- **AND** proceed only if user responds "y" or "yes"

#### Scenario: Headless fallback approval
- **WHEN** the agent attempts to call a tool that requires approval
- **AND** no channel-specific provider matches
- **AND** no companion is connected
- **AND** `HeadlessAutoApprove` is true
- **THEN** the system SHALL auto-approve via `HeadlessProvider`
- **AND** SHALL log a WARN-level audit message including the summary

#### Scenario: No approval source available
- **WHEN** the agent attempts to call a tool that requires approval
- **AND** no channel-specific provider matches
- **AND** no companion is connected
- **AND** `HeadlessAutoApprove` is false
- **AND** stdin is not a terminal
- **THEN** the tool execution SHALL be denied

### Requirement: Approval timeout configuration
The system SHALL support an `ApprovalTimeoutSec` configuration field in `InterceptorConfig` that controls how long to wait for an approval response before timing out.

#### Scenario: Default timeout
- **WHEN** `ApprovalTimeoutSec` is not set or is 0
- **THEN** the system SHALL use a default timeout of 30 seconds

#### Scenario: Custom timeout
- **WHEN** `ApprovalTimeoutSec` is set to a positive value
- **THEN** the system SHALL use that value as the timeout in seconds

