## ADDED Requirements

### Requirement: P2P session command group

The system SHALL provide `lango p2p session` subcommands for listing active sessions and revoking one or all authenticated peer sessions.

#### Scenario: Session list in text format
- **WHEN** user runs `lango p2p session list`
- **THEN** the command prints active sessions with peer DID, created timestamp, expiry timestamp, and ZK verification state

#### Scenario: Session list in JSON format
- **WHEN** user runs `lango p2p session list --json`
- **THEN** the JSON output SHALL contain the active session records

#### Scenario: Session revoke
- **WHEN** user runs `lango p2p session revoke --peer-did did:lango:abc123`
- **THEN** the command reports that the session was revoked

#### Scenario: Session revoke-all
- **WHEN** user runs `lango p2p session revoke-all`
- **THEN** the command reports that all sessions were revoked

### Requirement: P2P session output routing

`lango p2p session` subcommands SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture empty-state, table, JSON, and revoke confirmations without intercepting process-global stdout.

#### Scenario: Session list empty-state writes to command output
- **WHEN** user runs `lango p2p session list` with no active sessions
- **THEN** the command writes `No active sessions.` to the Cobra command output stream

#### Scenario: Session list table writes to command output
- **WHEN** user runs `lango p2p session list` with active sessions
- **THEN** the command writes the session table to the Cobra command output stream

#### Scenario: Session list JSON writes to command output
- **WHEN** user runs `lango p2p session list --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Session revoke confirmation writes to command output
- **WHEN** user runs `lango p2p session revoke --peer-did did:lango:abc123`
- **THEN** the command writes the revoke confirmation to the Cobra command output stream

#### Scenario: Session revoke-all confirmation writes to command output
- **WHEN** user runs `lango p2p session revoke-all`
- **THEN** the command writes the revoke-all confirmation to the Cobra command output stream
