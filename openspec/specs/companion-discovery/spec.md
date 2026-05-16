## Purpose

Capability spec for companion-discovery. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Companion connectivity is gateway-backed, not discovery-backed
The current runtime SHALL use the gateway `/companion` WebSocket endpoint for companion connectivity and SHALL NOT claim that an automatic discovery subsystem is currently shipped.

#### Scenario: Companion connects through gateway endpoint
- **WHEN** a companion app connects to `/companion`
- **THEN** the gateway registers it as a companion client for approval and RPC response routing

#### Scenario: No automatic discovery claim
- **WHEN** the companion-discovery main spec is read
- **THEN** it SHALL NOT claim that Lango currently browses a Bonjour/mDNS companion service type
- **AND** it SHALL NOT claim a legacy dedicated companion-address configuration path

### Requirement: Companion status reporting is connection-oriented
The system SHALL report whether a companion is currently connected through the gateway-facing connection model.

#### Scenario: Query companion connectivity state
- **WHEN** companion connectivity is inspected through the current runtime surfaces
- **THEN** the system reports whether at least one companion is connected
