## MODIFIED Requirements

### Requirement: Approval Provider interface
The system SHALL define a `Provider` interface with `RequestApproval(ctx, req) (ApprovalResponse, error)` and `CanHandle(sessionKey) bool` methods for handling tool execution approval requests. `ApprovalResponse` SHALL carry `Approved bool` and `AlwaysAllow bool` fields.

#### Scenario: Provider implementation
- **WHEN** a new approval channel is added
- **THEN** it SHALL implement the `Provider` interface returning `ApprovalResponse`
- **AND** `CanHandle` SHALL return true only for session keys it can handle

#### Scenario: Approve response
- **WHEN** a user approves a request
- **THEN** the provider SHALL return `ApprovalResponse{Approved: true, AlwaysAllow: false}`

#### Scenario: Always Allow response
- **WHEN** a user clicks "Always Allow"
- **THEN** the provider SHALL return `ApprovalResponse{Approved: true, AlwaysAllow: true}`

#### Scenario: Deny response
- **WHEN** a user denies a request
- **THEN** the provider SHALL return `ApprovalResponse{Approved: false, AlwaysAllow: false}`

### Requirement: TTY approval fallback behavior
The `TTYProvider.RequestApproval` SHALL return `(ApprovalResponse{}, error)` when stdin is not a terminal. When stdin is a terminal, it SHALL prompt with `[y/a/N]` where `a` means "always allow".

#### Scenario: Non-terminal environment returns error
- **WHEN** `TTYProvider.RequestApproval` is called and stdin is not a terminal
- **THEN** it SHALL return an empty `ApprovalResponse` and a non-nil error containing "not a terminal"

#### Scenario: Terminal user types 'a'
- **WHEN** the user enters "a" or "always" at the TTY prompt
- **THEN** it SHALL return `ApprovalResponse{Approved: true, AlwaysAllow: true}`

#### Scenario: Terminal user types 'y'
- **WHEN** the user enters "y" or "yes" at the TTY prompt
- **THEN** it SHALL return `ApprovalResponse{Approved: true, AlwaysAllow: false}`

### Requirement: Gateway approval provider
The system SHALL provide a `GatewayProvider` that delegates approval to connected companion apps via WebSocket. The `GatewayApprover` interface SHALL return `(ApprovalResponse, error)`.

#### Scenario: Companions connected
- **WHEN** a companion app is connected
- **THEN** `CanHandle` SHALL return true
- **AND** the approval request SHALL be forwarded to companions

#### Scenario: Companion sends alwaysAllow
- **WHEN** a companion responds with `{"approved": true, "alwaysAllow": true}`
- **THEN** the provider SHALL return `ApprovalResponse{Approved: true, AlwaysAllow: true}`

#### Scenario: Companion omits alwaysAllow (backward compatible)
- **WHEN** a companion responds with `{"approved": true}` without `alwaysAllow`
- **THEN** the provider SHALL return `ApprovalResponse{Approved: true, AlwaysAllow: false}`

## ADDED Requirements

### Requirement: Telegram Always Allow button
The Telegram approval message SHALL include a second row with an "Always Allow" button using the callback data prefix `always:`.

#### Scenario: Always Allow button layout
- **WHEN** an approval message is sent in Telegram
- **THEN** it SHALL have Row 1 with Approve and Deny buttons, and Row 2 with an Always Allow button

#### Scenario: Always Allow callback
- **WHEN** a user clicks "Always Allow" in Telegram
- **THEN** the response channel SHALL receive `ApprovalResponse{Approved: true, AlwaysAllow: true}`
- **AND** the message SHALL be edited to show "Always Allowed"

### Requirement: Discord Always Allow button
The Discord approval message SHALL include a second ActionsRow with a Secondary-style "Always Allow" button using the custom ID prefix `always:`.

#### Scenario: Always Allow button layout
- **WHEN** an approval message is sent in Discord
- **THEN** it SHALL have ActionsRow 1 with Approve and Deny, and ActionsRow 2 with Always Allow

#### Scenario: Always Allow interaction
- **WHEN** a user clicks "Always Allow" in Discord
- **THEN** the response channel SHALL receive `ApprovalResponse{Approved: true, AlwaysAllow: true}`
- **AND** the interaction response SHALL show "Always Allowed"

### Requirement: Slack Always Allow button
The Slack approval message SHALL include an "Always Allow" button in the action block with the action ID prefix `always:`.

#### Scenario: Always Allow button in action block
- **WHEN** an approval message is sent in Slack
- **THEN** the action block SHALL contain Approve, Deny, and Always Allow buttons

#### Scenario: Always Allow interactive action
- **WHEN** a user clicks "Always Allow" in Slack
- **THEN** the response channel SHALL receive `ApprovalResponse{Approved: true, AlwaysAllow: true}`
- **AND** the message SHALL be updated to show "Always Allowed"
