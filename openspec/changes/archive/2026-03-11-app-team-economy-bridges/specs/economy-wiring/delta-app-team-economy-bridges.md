# Delta Spec: Economy Wiring — Team-Economy Bridges

## Overview

Updates the economy wiring to support hub/vault on-chain settlement modes, DanglingDetector lifecycle, and on-chain escrow bridge initialization.

## MODIFIED Requirements

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

## ADDED Requirements

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
