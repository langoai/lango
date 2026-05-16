## ADDED Requirements

### Requirement: P2P provenance command group

The system SHALL provide `lango p2p provenance push` and `lango p2p provenance fetch` for exchanging signed provenance bundles through the running gateway using authenticated P2P sessions.

#### Scenario: Provenance push
- **WHEN** user runs `lango p2p provenance push <peer-did> <session-key>`
- **THEN** the command pushes a provenance bundle through the gateway and reports success for the target peer DID

#### Scenario: Provenance fetch
- **WHEN** user runs `lango p2p provenance fetch <peer-did> <session-key>`
- **THEN** the command fetches a provenance bundle through the gateway and reports success for the target peer DID

### Requirement: P2P provenance output routing

`lango p2p provenance push` and `lango p2p provenance fetch` SHALL write success confirmations through the Cobra command output stream so wrappers and test harnesses can capture them without intercepting process-global stdout.

#### Scenario: Provenance push confirmation writes to command output
- **WHEN** user runs `lango p2p provenance push <peer-did> <session-key>`
- **THEN** the command writes the push confirmation to the Cobra command output stream

#### Scenario: Provenance fetch confirmation writes to command output
- **WHEN** user runs `lango p2p provenance fetch <peer-did> <session-key>`
- **THEN** the command writes the fetch confirmation to the Cobra command output stream
