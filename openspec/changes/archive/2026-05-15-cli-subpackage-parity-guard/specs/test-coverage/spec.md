## ADDED Requirements

### Requirement: CLI-subpackage parity guard stays executable
Repository-level regressions that let README or architecture inventory docs fall out of parity with the shipped `internal/cli/` subpackage tree SHALL be enforced by an executable test.

#### Scenario: Every CLI subpackage remains represented in both inventories
- **WHEN** the repository still ships subpackages under `internal/cli/`
- **THEN** an executable repository test SHALL fail if any `internal/cli/` subpackage disappears from `README.md`
- **AND** it SHALL fail if any `internal/cli/` subpackage disappears from `docs/architecture/project-structure.md`
- **AND** it SHALL therefore catch omissions affecting both command families and helper packages such as `cliboot`, `clihttp`, `clitypes`, `tuicore`, `workbench`, or `workbenchstart`
