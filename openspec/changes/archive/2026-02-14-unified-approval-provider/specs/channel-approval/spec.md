## ADDED Requirements

### Requirement: Approval Provider interface
The system SHALL define a `Provider` interface with `RequestApproval(ctx, req) (bool, error)` and `CanHandle(sessionKey) bool` methods for handling tool execution approval requests.

#### Scenario: Provider implementation
- **WHEN** a new approval channel is added
- **THEN** it SHALL implement the `Provider` interface
- **AND** `CanHandle` SHALL return true only for session keys it can handle

### Requirement: Composite approval routing
The system SHALL provide a `CompositeProvider` that routes approval requests to the first registered provider whose `CanHandle` returns true for the given session key.

#### Scenario: Route to matching provider
- **WHEN** an approval request has session key "telegram:123:456"
- **AND** a Telegram provider is registered
- **THEN** the request SHALL be routed to the Telegram provider

#### Scenario: Multiple providers registered
- **WHEN** multiple providers are registered
- **THEN** the first provider whose `CanHandle` returns true SHALL handle the request

#### Scenario: No matching provider with TTY fallback
- **WHEN** no registered provider can handle the session key
- **AND** a TTY fallback is configured
- **THEN** the request SHALL be routed to the TTY fallback

#### Scenario: No matching provider without fallback (fail-closed)
- **WHEN** no registered provider can handle the session key
- **AND** no TTY fallback is configured
- **THEN** the request SHALL be denied (return false, nil)

### Requirement: Thread-safe provider registration
The system SHALL allow providers to be registered concurrently without data races.

#### Scenario: Concurrent registration
- **WHEN** multiple providers are registered from different goroutines
- **THEN** all registrations SHALL complete without data races

### Requirement: TTY approval fallback
The system SHALL provide a `TTYProvider` that prompts the user via terminal stdin for approval as a last-resort fallback.

#### Scenario: TTY prompt
- **WHEN** TTY fallback is invoked
- **AND** stdin is a terminal
- **THEN** the system SHALL print a prompt to stderr and read y/N from stdin

#### Scenario: Non-terminal stdin
- **WHEN** TTY fallback is invoked
- **AND** stdin is not a terminal
- **THEN** the request SHALL be denied (return false, nil)

### Requirement: Gateway approval provider
The system SHALL provide a `GatewayProvider` that delegates approval to connected companion apps via WebSocket.

#### Scenario: Companions connected
- **WHEN** a companion app is connected
- **THEN** `CanHandle` SHALL return true
- **AND** the approval request SHALL be forwarded to companions

#### Scenario: No companions connected
- **WHEN** no companion app is connected
- **THEN** `CanHandle` SHALL return false

### Requirement: Approval request context
Each approval request SHALL carry an ID, tool name, session key, parameters, and creation timestamp.

#### Scenario: Request fields
- **WHEN** an approval request is created
- **THEN** it SHALL contain a unique ID, the tool name, the originating session key, tool parameters, and a timestamp
