## MODIFIED Requirements

### Requirement: P2P Network settings form

#### Scenario: User enables PQ handshake
- **WHEN** user navigates to "P2P Network" and sets "Enable PQ Handshake" to true
- **THEN** the config's `p2p.enablePqHandshake` field SHALL be set to true upon save
- **AND** the default value SHALL be false (opt-in)
