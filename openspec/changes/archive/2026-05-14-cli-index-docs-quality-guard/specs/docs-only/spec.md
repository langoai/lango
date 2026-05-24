## ADDED Requirements
### Requirement: CLI index quick reference completeness and table structure are enforced mechanically
The public CLI quick reference SHALL keep implemented operator commands visible and SHALL keep explanatory prose outside command tables.

#### Scenario: CLI index regressions are rejected
- **WHEN** `docs/cli/index.md` drops implemented KMS wrap/detach or P2P workspace/provenance command rows, or reintroduces prose that splits the Agent & Memory table
- **THEN** an executable repository test SHALL fail
