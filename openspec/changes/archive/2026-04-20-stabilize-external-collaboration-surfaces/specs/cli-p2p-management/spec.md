## MODIFIED Requirements

### Requirement: P2P identity command
The system SHALL provide `lango p2p identity [--json]` that displays the active DID when available, the local peer ID, key storage mode, and listen addresses.

#### Scenario: Identity in text format
- **WHEN** user runs `lango p2p identity`
- **THEN** the command SHALL print the active DID when one is available
- **AND** it SHALL print peer ID, key storage mode, and listen addresses

#### Scenario: Identity in JSON format
- **WHEN** user runs `lango p2p identity --json`
- **THEN** the JSON output SHALL include keys `did`, `peerId`, `listenAddrs`, and `keyStorage`
- **AND** `did` SHALL be `null` when no active DID is available

## ADDED Requirements

### Requirement: Team CLI is guidance-oriented until live team control exists
The `lango p2p team` CLI surface SHALL describe the current runtime honestly: teams are real runtime structures, but the CLI is guidance-oriented until full live control is implemented.

#### Scenario: Team list guidance
- **WHEN** user runs `lango p2p team list`
- **THEN** the command SHALL describe the current guidance-oriented/runtime-backed path instead of implying direct live team control

### Requirement: Workspace and git CLI are guidance-oriented until live control exists
The `lango p2p workspace` and `lango p2p git` CLI surfaces SHALL describe the current runtime honestly: the runtime subsystems are real, but the CLI commands mainly guide operators toward server-backed or tool-backed flows until fuller live control exists.

#### Scenario: Workspace create guidance
- **WHEN** user runs `lango p2p workspace create`
- **THEN** the command SHALL explain the server-backed or tool-backed creation path instead of implying a fully direct live CLI operation

#### Scenario: Git bundle guidance
- **WHEN** user runs `lango p2p git push` or `lango p2p git fetch`
- **THEN** the command SHALL describe the current server-backed or tool-backed exchange path instead of implying a fully direct live CLI operation
