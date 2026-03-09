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

### Requirement: Economy agent tools registration
The system SHALL register 12 economy agent tools under the "economy" catalog category. Tools SHALL be built from the economyComponents struct.

#### Scenario: Tools registered
- **WHEN** economy is enabled and initEconomy succeeds
- **THEN** 12 tools are added to the tool catalog under category "economy"
