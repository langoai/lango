## ADDED Requirements

### Requirement: P2P connect docs describe bounded timeout
Public P2P CLI documentation SHALL describe that `lango p2p connect` uses a bounded connect attempt tied to the P2P handshake timeout.

#### Scenario: P2P CLI docs mention bounded connect behavior
- **WHEN** a user reads `docs/cli/p2p.md`
- **THEN** the `lango p2p connect` section SHALL mention that connect attempts use `p2p.handshakeTimeout`
- **AND** it SHALL mention the 30 second fallback when the timeout is unset or invalid
- **AND** it SHALL mention that command cancellation stops the connect attempt
