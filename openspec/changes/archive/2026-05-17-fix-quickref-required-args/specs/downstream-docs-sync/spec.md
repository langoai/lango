## MODIFIED Requirements

### Requirement: README reflects all implemented features
The README SHALL list all implemented features with quick-reference commands
that include required positional arguments and required flags.

#### Scenario: Provenance quick reference includes required operands
- **WHEN** a user reads the provenance quick reference in `README.md`
- **THEN** commands that require labels, session keys, bundle files, or `--run`
  SHALL show those required operands

#### Scenario: P2P reputation quick reference includes required peer DID
- **WHEN** a user reads the P2P quick reference in `README.md`
- **THEN** `lango p2p reputation` SHALL include `--peer-did <did>`

### Requirement: CLI index quick references include required operands
The CLI index SHALL list quick-reference commands with required positional
arguments and required flags for provenance and P2P reputation commands.

#### Scenario: CLI index provenance quick reference includes required operands
- **WHEN** a user reads `docs/cli/index.md`
- **THEN** provenance commands that require labels, session keys, bundle files,
  or `--run` SHALL show those required operands

#### Scenario: CLI index P2P reputation includes required peer DID
- **WHEN** a user reads `docs/cli/index.md`
- **THEN** `lango p2p reputation` SHALL include `--peer-did <did>`
