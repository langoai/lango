## Purpose

CLI subcommand and REST endpoint for querying peer reputation details and trust scores from the local database.

## Requirements

### Requirement: CLI reputation command
The system SHALL provide a `lango p2p reputation` CLI command that queries peer reputation details from the local database.

#### Scenario: Query reputation for known peer
- **WHEN** user runs `lango p2p reputation --peer-did "did:lango:abc123"`
- **THEN** system displays trust score, successful exchanges, failed exchanges, timeout count, first seen date, and last interaction date in table format

#### Scenario: Query reputation with JSON output
- **WHEN** user runs `lango p2p reputation --peer-did "did:lango:abc123" --json`
- **THEN** system outputs full PeerDetails as JSON to stdout

#### Scenario: Query reputation for unknown peer
- **WHEN** user runs `lango p2p reputation --peer-did "did:lango:unknown"`
- **THEN** system displays "No reputation record found" message

#### Scenario: Missing peer-did flag
- **WHEN** user runs `lango p2p reputation` without `--peer-did`
- **THEN** system returns an error stating `--peer-did is required`

### Requirement: CLI reputation registered as subcommand
The `reputation` command SHALL be registered as a subcommand of `lango p2p` in `internal/cli/p2p/p2p.go`.

#### Scenario: Help shows reputation command
- **WHEN** user runs `lango p2p --help`
- **THEN** output lists `reputation` as an available subcommand

### Requirement: P2P reputation CLI uses reputation store capability
The `lango p2p reputation` command MUST obtain its reputation store through a storage facade capability instead of constructing it from a generic Ent client in the CLI layer.

#### Scenario: Reputation command reads through facade capability
- **WHEN** the user runs `lango p2p reputation`
- **THEN** the command resolves the reputation store from the storage facade
