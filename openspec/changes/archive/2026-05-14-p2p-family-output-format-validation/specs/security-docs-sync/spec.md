## ADDED Requirements
### Requirement: P2P family docs sync uses explicit output selection
The downstream `security-docs-sync` requirement for `docs/cli/p2p.md` SHALL reference the migrated P2P operator family with `--output table|json` instead of boolean `--json` toggles.

#### Scenario: P2P docs sync requirement uses explicit output format
- **WHEN** a maintainer reads the downstream P2P docs sync requirement
- **THEN** any documented machine-readable P2P operator command uses `--output table|json`
