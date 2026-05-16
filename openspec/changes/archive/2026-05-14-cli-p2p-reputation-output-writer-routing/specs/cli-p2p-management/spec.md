## ADDED Requirements

### Requirement: P2P reputation command

The system SHALL provide `lango p2p reputation --peer-did <did> [--json]` that displays a peer trust score, exchange history, and interaction timeline.

#### Scenario: Reputation in text format
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:abc123`
- **THEN** the command prints peer DID, trust score, successes, failures, timeouts, first seen, and last interaction

#### Scenario: Reputation in JSON format
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:abc123 --json`
- **THEN** the JSON output SHALL include `peerDid`, `trustScore`, `successfulExchanges`, `failedExchanges`, `timeoutCount`, `firstSeen`, and `lastInteraction`

#### Scenario: Reputation missing record
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:missing`
- **THEN** the command reports that no reputation record was found for that DID

### Requirement: P2P reputation output routing

`lango p2p reputation` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture missing-record, text, and JSON output without intercepting process-global stdout.

#### Scenario: Reputation missing-record output writes to command output
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:missing`
- **THEN** the command writes the missing-record message to the Cobra command output stream

#### Scenario: Reputation text output writes to command output
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:abc123`
- **THEN** the command writes the reputation text output to the Cobra command output stream

#### Scenario: Reputation JSON output writes to command output
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:abc123 --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream
