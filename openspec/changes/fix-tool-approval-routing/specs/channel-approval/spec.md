## MODIFIED Requirements

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
- **THEN** the request SHALL be denied (return false)
- **AND** an error SHALL be returned with the message `no approval provider for session "<sessionKey>"`

## ADDED Requirements

### Requirement: Session key context propagation
The `runAgent` function SHALL inject the session key into the context via `WithSessionKey` before passing it to the agent pipeline, ensuring downstream components (approval providers, learning engine) can access the session key via `SessionKeyFromContext`.

#### Scenario: Channel message triggers agent with session key
- **WHEN** a Telegram/Discord/Slack handler calls `runAgent(ctx, sessionKey, input)`
- **THEN** `runAgent` SHALL call `WithSessionKey(ctx, sessionKey)` before invoking the agent
- **AND** `SessionKeyFromContext` SHALL return the session key within the agent pipeline

#### Scenario: Session key reaches approval provider
- **WHEN** a tool requiring approval is invoked from a channel message
- **THEN** the `ApprovalRequest.SessionKey` field SHALL contain the channel session key (e.g., "telegram:123:456")
- **AND** `CompositeProvider` SHALL route to the matching channel provider
