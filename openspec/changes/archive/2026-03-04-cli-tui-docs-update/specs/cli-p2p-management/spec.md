## MODIFIED Requirements

### Requirement: Team subcommand group addition
The existing `lango p2p` command group SHALL gain a new `team` subcommand group containing list, status, and disband subcommands for P2P team lifecycle management. The team subcommand group uses bootLoader for config access but does NOT initialize a full P2P node.

#### Scenario: P2P help includes team
- **WHEN** user runs `lango p2p --help`
- **THEN** the help output lists team alongside existing P2P subcommands (status, peers, connect, disconnect, firewall, discover, identity, session)

### Requirement: ZKP subcommand group addition
The existing `lango p2p` command group SHALL gain a new `zkp` subcommand group containing status and circuits subcommands for ZKP inspection. The zkp status subcommand uses cfgLoader; the zkp circuits subcommand requires no loader.

#### Scenario: P2P help includes zkp
- **WHEN** user runs `lango p2p --help`
- **THEN** the help output lists zkp alongside existing P2P subcommands

### Requirement: Existing P2P commands unaffected
The addition of team and zkp subcommand groups SHALL NOT change the behavior or registration of any existing P2P subcommands.

#### Scenario: Existing P2P status still works
- **WHEN** user runs `lango p2p status`
- **THEN** the command behaves identically to before the team/zkp additions

### Requirement: P2P disabled gating
The team and zkp status subcommands SHALL respect the existing P2P disabled error pattern: when `p2p.enabled` is false, the commands SHALL return the standard error "P2P networking is not enabled (set p2p.enabled = true)". The zkp circuits subcommand SHALL NOT be gated by p2p.enabled since it returns static data.

#### Scenario: Team command with P2P disabled
- **WHEN** user runs `lango p2p team list` with P2P disabled
- **THEN** system returns the standard P2P disabled error

#### Scenario: ZKP circuits with P2P disabled
- **WHEN** user runs `lango p2p zkp circuits` with P2P disabled
- **THEN** system still displays the circuit list since it is static data
