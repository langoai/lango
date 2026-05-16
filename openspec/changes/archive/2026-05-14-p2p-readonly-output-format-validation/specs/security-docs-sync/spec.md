## ADDED Requirements
### Requirement: Session management docs reference the current output contract
The `security-docs-sync` downstream requirement for `docs/cli/p2p.md` SHALL describe `lango p2p session list` with `--output table|json`.

#### Scenario: Session docs sync requirement uses explicit output format
- **WHEN** a maintainer reads the security docs sync requirement for session commands
- **THEN** it references `lango p2p session list [--output table|json]`
