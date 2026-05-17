## MODIFIED Requirements

### Requirement: README P2P completeness guard stays executable
Repository-level regressions that drop implemented `p2p workspace`, `p2p team`, or `p2p zkp` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented P2P command families remain listed
- **WHEN** the repository still ships the implemented `workspace`, `team`, and `zkp` P2P CLI families
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

#### Scenario: P2P quick-reference required operands remain listed
- **WHEN** the repository still ships P2P firewall and session commands that require peer identifiers
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits `--peer-did <did>` or `<peer-did>` from those quick-reference entries
- **AND** the test SHALL fail if `docs/features/p2p-network.md` or `docs/features/zkp.md` shows affected P2P commands without required peer operands

### Requirement: Memory quick-reference completeness guard stays executable
Repository-level regressions that drop required memory command operands from public quick references SHALL be enforced by an executable test.

#### Scenario: Memory clear required session key remains listed
- **WHEN** the repository still ships `lango memory clear <session-key>`
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` lists the command without `<session-key>`
- **AND** the test SHALL fail if `docs/features/observational-memory.md` shows `lango memory clear` without `<session-key>`

### Requirement: Config quick-reference completeness guard stays executable
Repository-level regressions that drop `config get` output or secret flags from public quick references SHALL be enforced by an executable test.

#### Scenario: Config get full usage remains listed
- **WHEN** the repository still ships `lango config get <dot.path>` with `--output plain|json` and `--show-secrets`
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits the full usage string
