## MODIFIED Requirements

### Requirement: README reflects all implemented features
The README SHALL list all implemented features including Team Health Monitoring, Incremental Git Bundles, Task Branch Management, Config Presets, Event-Driven Bridges, EventMonitor Reorg Protection, and Escrow Hub V2.

#### Scenario: New features in README
- **WHEN** a user reads `README.md`
- **THEN** all 7 new feature areas SHALL be listed in the features section

#### Scenario: CLI commands in README
- **WHEN** a user reads the CLI commands section of `README.md`
- **THEN** `lango status`, `lango onboard --preset`, cron `--timeout`, cron `--deliver`, and cron management by `id-or-name` SHALL be documented

#### Scenario: Provenance quick reference includes required operands
- **WHEN** a user reads the provenance quick reference in `README.md`
- **THEN** it SHALL include `lango provenance checkpoint list --run <id>`, `lango provenance checkpoint create <label> --run <id>`, and `lango provenance checkpoint show <id>`
- **AND** it SHALL include `lango provenance session tree <session-key>` and `lango provenance session list`
- **AND** it SHALL include `lango provenance attribution show <session-key>` and `lango provenance attribution report <session-key>`
- **AND** it SHALL include `lango provenance bundle export <session-key>` and `lango provenance bundle import <file>`

#### Scenario: P2P reputation quick reference includes required peer DID
- **WHEN** a user reads the P2P quick reference in `README.md`
- **THEN** it SHALL include `lango p2p reputation --peer-did <did>`
- **AND** workspace quick-reference summaries SHALL describe direct local workspace actions for create, list, status, join, and leave

#### Scenario: README quick references include required memory and P2P operands
- **WHEN** a user reads the CLI commands section of `README.md`
- **THEN** it SHALL include `lango memory clear <session-key>`
- **AND** it SHALL include `lango p2p firewall add --peer-did <did>`
- **AND** it SHALL include `lango p2p firewall remove <peer-did>`
- **AND** it SHALL include `lango p2p session revoke --peer-did <did>`

#### Scenario: README config get quick reference includes output and secret flags
- **WHEN** a user reads the config quick reference in `README.md`
- **THEN** it SHALL include `lango config get <dot.path> [--output plain|json] [--show-secrets]`

### Requirement: CLI index quick references include required operands
The CLI index SHALL list quick-reference commands with required positional
arguments and required flags for provenance and P2P reputation commands.

#### Scenario: CLI index provenance quick reference includes required operands
- **WHEN** a user reads `docs/cli/index.md`
- **THEN** it SHALL include `lango provenance checkpoint list --run <id>`, `lango provenance checkpoint create <label> --run <id>`, and `lango provenance checkpoint show <id>`
- **AND** it SHALL include `lango provenance session tree <session-key>` and `lango provenance session list`
- **AND** it SHALL include `lango provenance attribution show <session-key>` and `lango provenance attribution report <session-key>`
- **AND** it SHALL include `lango provenance bundle export <session-key>` and `lango provenance bundle import <file>`

#### Scenario: CLI index P2P reputation includes required peer DID
- **WHEN** a user reads `docs/cli/index.md`
- **THEN** it SHALL include `lango p2p reputation --peer-did <did>`
- **AND** workspace quick-reference summaries SHALL describe direct local workspace actions for create, list, status, join, and leave

#### Scenario: CLI index quick references include required memory and P2P operands
- **WHEN** a user reads `docs/cli/index.md`
- **THEN** it SHALL include `lango memory clear <session-key>`
- **AND** it SHALL include `lango p2p firewall add --peer-did <did>`
- **AND** it SHALL include `lango p2p firewall remove <peer-did>`
- **AND** it SHALL include `lango p2p session revoke --peer-did <did>`

#### Scenario: CLI index config get quick reference includes output and secret flags
- **WHEN** a user reads `docs/cli/index.md`
- **THEN** it SHALL include `lango config get <dot.path> [--output plain|json] [--show-secrets]`

### Requirement: Feature docs include required command operands
Public feature docs SHALL include required operands when showing runnable command examples.

#### Scenario: Observational memory docs include clear session key
- **WHEN** a user reads `docs/features/observational-memory.md`
- **THEN** it SHALL include `lango memory clear <session-key>`

#### Scenario: P2P feature docs include firewall and session peer operands
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** it SHALL include `lango p2p firewall add --peer-did <did>`
- **AND** it SHALL include `lango p2p session revoke --peer-did <did>`

#### Scenario: ZKP feature docs include session revoke peer operand
- **WHEN** a user reads `docs/features/zkp.md`
- **THEN** it SHALL include `lango p2p session revoke --peer-did <did>`
