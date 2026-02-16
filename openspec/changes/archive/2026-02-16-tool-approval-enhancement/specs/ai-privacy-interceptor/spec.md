## MODIFIED Requirements

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
