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

#### Scenario: Provider tags response source
- **WHEN** a provider returns an approval response successfully
- **THEN** the response SHALL include provider metadata indicating which approval backend handled it

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

### Requirement: Turn-local approval replay protection
Each request SHALL maintain turn-local approval state keyed by `tool name + canonical params JSON`. The approval middleware SHALL consult this state before issuing a new approval request.

#### Scenario: Turn-local positive replay
- **WHEN** a request already approved a specific `tool + params` once in the current turn
- **THEN** an identical retry in the same turn SHALL execute without issuing another approval prompt

#### Scenario: Turn-local negative replay block
- **WHEN** a request already received deny, timeout, or unavailable for a specific `tool + params` in the current turn
- **THEN** an identical retry in the same turn SHALL return the same failure immediately without issuing another approval prompt

#### Scenario: Different params require new approval
- **WHEN** the retried tool call uses different params
- **THEN** the middleware SHALL treat it as a new approval request

#### Scenario: Always Allow still uses session-wide grant store
- **WHEN** a user selects `Always Allow`
- **THEN** the approval result SHALL be persisted in the session-wide grant store
- **AND** future matching calls MAY bypass approval in later turns according to existing grant-store behavior

### Requirement: Structured approval failures
The approval system SHALL expose structured sentinel errors for deny, timeout, and unavailable outcomes.

#### Scenario: User deny returns denied sentinel
- **WHEN** the user denies the approval request
- **THEN** the middleware SHALL return an error wrapping `approval.ErrDenied`
- **AND** the user-facing message SHALL state that execution was denied by user approval

#### Scenario: Approval timeout returns timeout sentinel
- **WHEN** the approval request expires without response
- **THEN** the middleware SHALL return an error wrapping `approval.ErrTimeout`
- **AND** the user-facing message SHALL state that approval expired

#### Scenario: No provider available returns unavailable sentinel
- **WHEN** no approval provider can handle the request
- **THEN** the middleware SHALL return an error wrapping `approval.ErrUnavailable`
- **AND** the user-facing message SHALL state that no approval channel is available

### Requirement: Approval observability logs
The approval flow SHALL emit structured logs for request, callback, final outcome, turn-local bypass, and replay-block events.

#### Scenario: Approval request logged
- **WHEN** the middleware issues an approval request
- **THEN** it SHALL log session, request ID, tool, summary, params hash, provider, outcome=`requested`, and grant scope

#### Scenario: Turn-local bypass logged
- **WHEN** the middleware reuses a turn-local positive grant
- **THEN** it SHALL log outcome=`bypass` and grant scope=`turn`

#### Scenario: Replay-block logged
- **WHEN** the middleware short-circuits an identical denied/expired/unavailable request
- **THEN** it SHALL log outcome=`replay_blocked` and the cached failure kind
