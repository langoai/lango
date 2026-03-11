## Purpose

Wiring layer that connects all 5 economy subsystems (budget, risk, pricing, negotiation, escrow) into the application lifecycle, event bus, and P2P protocol handler.

## Requirements

### Requirement: Economy component initialization
The system SHALL initialize all 5 economy subsystems (budget, risk, pricing, negotiation, escrow) during app startup via initEconomy(). Initialization SHALL occur after P2P wiring and before agent tool registration. The function SHALL accept `*paymentComponents` to enable on-chain escrow settlement.

#### Scenario: Economy enabled
- **WHEN** economy.enabled is true in config
- **THEN** all 5 engines are created and wired with cross-system callbacks

#### Scenario: Economy disabled
- **WHEN** economy.enabled is false in config
- **THEN** initEconomy returns nil and no economy components are initialized

#### Scenario: Payment components passed to initEconomy
- **WHEN** `app.New()` initializes the economy layer
- **THEN** the `paymentComponents` from `initPayment` is passed as the `pc` parameter to `initEconomy`

#### Scenario: Nil payment components handled gracefully
- **WHEN** `initEconomy` receives nil `paymentComponents`
- **THEN** escrow falls back to `noopSettler` and all other economy components initialize normally

### Requirement: Cross-system callback wiring
The system SHALL wire callbacks between economy subsystems without direct imports: reputation querier from P2P into risk and pricing engines, risk assessor into budget engine, pricing querier into negotiation engine.

#### Scenario: Reputation callback wiring
- **WHEN** initEconomy is called with P2P components containing a reputation store
- **THEN** risk and pricing engines receive a ReputationQuerier that delegates to the P2P reputation store

#### Scenario: Risk-to-budget wiring
- **WHEN** initEconomy creates budget and risk engines
- **THEN** budget engine receives a RiskAssessor callback that delegates to the risk engine

### Requirement: Event bus integration
The system SHALL publish economy events (budget alerts, negotiation state changes, escrow milestones) through the existing eventbus.Bus. 8 event types SHALL be defined.

#### Scenario: Budget alert event
- **WHEN** budget spending crosses a threshold
- **THEN** a BudgetAlertEvent is published to the event bus

### Requirement: P2P negotiation protocol routing
The system SHALL route RequestNegotiatePropose and RequestNegotiateRespond messages from the P2P protocol handler to the negotiation engine via SetNegotiator.

#### Scenario: Negotiate propose arrives via P2P
- **WHEN** a RequestNegotiatePropose message is received by the protocol handler
- **THEN** the message is routed to the negotiation engine's Propose method

### Requirement: Escrow settlement mode selection
The economy wiring SHALL select the settlement executor based on the on-chain mode configuration, supporting `"hub"` (shared escrow contract), `"vault"` (per-deal beacon proxy), and custodian (default USDCSettler) modes.

#### Scenario: Hub mode settler
- **WHEN** `economy.escrow.onChain.mode` is `"hub"` and `hubAddress` is configured
- **THEN** the system SHALL create a `HubSettler` with the configured hub and token addresses

#### Scenario: Vault mode settler
- **WHEN** `economy.escrow.onChain.mode` is `"vault"` and factory/implementation addresses are configured
- **THEN** the system SHALL create a `VaultSettler` with the configured factory, implementation, token, and arbitrator addresses

#### Scenario: Fallback to custodian mode
- **WHEN** on-chain mode is not enabled or required addresses are missing
- **THEN** the system SHALL fall back to the `USDCSettler` (custodian mode)

### Requirement: Economy agent tools registration
The system SHALL register 12 economy agent tools under the "economy" catalog category. Tools SHALL be built from the economyComponents struct.

#### Scenario: Tools registered
- **WHEN** economy is enabled and initEconomy succeeds
- **THEN** 12 tools are added to the tool catalog under category "economy"

### Requirement: DanglingDetector lifecycle wiring
The economy wiring SHALL create and register a DanglingDetector when on-chain escrow mode is enabled, to expire stuck pending escrows.

#### Scenario: DanglingDetector created with on-chain escrow
- **WHEN** on-chain escrow mode is enabled and an RPC client is available
- **THEN** the system SHALL create a `DanglingDetector` and register it with the lifecycle registry at `PriorityAutomation`

#### Scenario: DanglingDetector publishes events
- **WHEN** the DanglingDetector detects a stuck pending escrow
- **THEN** it SHALL publish an `EscrowDanglingEvent` to the event bus

### Requirement: On-chain escrow bridge wiring
The economy initialization SHALL wire the on-chain escrow bridge when on-chain mode is enabled and an RPC client is available.

#### Scenario: Bridge initialized during economy setup
- **WHEN** on-chain escrow is enabled with a valid hub address and RPC client
- **THEN** the system SHALL call `initOnChainEscrowBridge` to subscribe to on-chain events

#### Scenario: EventMonitor lifecycle registration
- **WHEN** an EventMonitor is successfully created
- **THEN** the system SHALL register it with the lifecycle registry at `PriorityNetwork`
