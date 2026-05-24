## ADDED Requirements

### Requirement: P2P CLI docs describe command-scoped ephemeral node startup
Public P2P CLI documentation SHALL describe cancellation behavior for commands that start ephemeral P2P nodes.

#### Scenario: Docs mention command cancellation during P2P node startup
- **WHEN** public docs describe `lango p2p` commands that create ephemeral nodes
- **THEN** the docs SHALL state that command cancellation also cancels DHT bootstrap, bootstrap peer dials, and mDNS discovered-peer connection attempts for ephemeral nodes
