## MODIFIED Requirements

### Requirement: Provenance CLI
The system SHALL provide server-backed remote provenance exchange commands under `lango p2p provenance`.

#### Scenario: Push bundle to peer
- **WHEN** the operator runs `lango p2p provenance push <peer-did> <session-key> --redaction <level>`
- **THEN** the CLI calls the running gateway
- **AND** the gateway exports a signed bundle locally and sends it to the target peer over the provenance P2P protocol

#### Scenario: Fetch bundle from peer
- **WHEN** the operator runs `lango p2p provenance fetch <peer-did> <session-key> --redaction <level>`
- **THEN** the CLI calls the running gateway
- **AND** the gateway requests a signed bundle from the target peer over the provenance P2P protocol
- **AND** the returned bundle is verify-and-store imported locally

#### Scenario: Active session required
- **WHEN** there is no active authenticated session for the target peer DID
- **THEN** push and fetch fail with an actionable error indicating that an active P2P session is required

### Requirement: Provenance P2P Transport
The provenance transport SHALL support both push and fetch flows.

#### Scenario: Fetch bundle request
- **WHEN** a peer receives a `fetch_bundle` provenance request with `session-key` and `redaction`
- **THEN** it exports a signed provenance bundle for that session and redaction level
- **AND** it returns the bundle over the provenance-specific P2P protocol
