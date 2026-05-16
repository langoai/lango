# Spec: Callback Wiring Completion

## Purpose

Capability spec for callback-wiring. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Session on-chain registration and revocation callbacks
When `SessionValidatorAddress` is configured, the session manager SHALL wire `WithOnChainRegistration` and `WithOnChainRevocation` options that call the `SessionValidatorClient`.

#### Scenario: Session creation registers key on-chain
- **WHEN** SessionValidator address is configured and a session key is created
- **THEN** `RegisterSessionKey` SHALL be called on-chain

#### Scenario: Session revocation revokes key on-chain
- **WHEN** SessionValidator address is configured and a session key is revoked
- **THEN** `RevokeSessionKey` SHALL be called on-chain

### Requirement: Budget engine sync uses OnChainTracker callback
The `OnChainTracker.SetCallback` SHALL forward spending data to the budget engine's `Record()` method instead of only logging it.

#### Scenario: Spending callback reaches budget engine
- **WHEN** the on-chain tracker emits spending data through its callback
- **THEN** the budget engine SHALL record that data

### Requirement: P2P CardFn provides agent information
The protocol handler SHALL receive a `CardFn` that returns the agent's name, DID, and peer ID.

#### Scenario: Protocol handler resolves local card information
- **WHEN** the protocol handler needs local card information
- **THEN** `CardFn` SHALL return the configured agent name, DID, and peer ID

### Requirement: Gossip service starts after creation
After creation, `gossip.Start()` SHALL be called to begin the publish/subscribe loops.

#### Scenario: Gossip loops begin after initialization
- **WHEN** the gossip service is created successfully
- **THEN** `gossip.Start()` SHALL be invoked

### Requirement: Team invoke uses the real handler path
The team coordinator's `invokeFn` SHALL route through the P2P protocol handler to send real remote tool invocation requests instead of returning a stub error.

#### Scenario: Team invoke dispatches through protocol handler
- **WHEN** the team coordinator invokes a remote tool
- **THEN** the request SHALL be routed through the P2P protocol handler

### Requirement: SmartAccount components are publicly accessible
All smart account sub-components (session manager, policy engine, module registry, bundler, paymaster, on-chain tracker) SHALL be accessible via public accessor methods from the `App` struct.

#### Scenario: App exposes smart-account sub-components
- **WHEN** callers need smart-account support components from the application
- **THEN** the app SHALL provide public accessors for the configured sub-components

### Requirement: Cross-domain callbacks are replaced by EventBus

The following SetXxxCallback methods SHALL be removed and replaced by EventBus publish/subscribe:
- `SetEmbedCallback` on knowledge and memory stores → `ContentSavedEvent`
- `SetGraphCallback` on knowledge, memory, and learning stores → `ContentSavedEvent` (NeedsGraph) + `TriplesExtractedEvent`
- `SetOnChangeCallback` on reputation store → `ReputationChangedEvent`

Stores SHALL accept `*eventbus.Bus` via `SetEventBus(bus)` method. When bus is nil, publish is silently skipped.

#### Scenario: Knowledge store publishes ContentSavedEvent
- **WHEN** a knowledge store with EventBus set saves content
- **THEN** `ContentSavedEvent` SHALL be published

#### Scenario: Nil bus is a no-op
- **WHEN** a store has no bus and content is saved
- **THEN** no panic SHALL occur
- **AND** no event SHALL be published

#### Scenario: Domain-internal hooks remain
- **WHEN** domain-internal hooks such as `negotiation.SetEventCallback` or `SessionStore.SetInvalidationCallback` are evaluated
- **THEN** they SHALL NOT be removed by this EventBus migration
