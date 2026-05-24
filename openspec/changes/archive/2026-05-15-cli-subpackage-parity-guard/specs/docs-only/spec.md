## ADDED Requirements

### Requirement: README and architecture inventory stay in CLI-subpackage parity
The public README internal tree and architecture project-structure inventory SHALL both mention every shipped `internal/cli/` subpackage so that new CLI subpackages, including shared helper packages, do not silently appear in only one document.

#### Scenario: Every CLI subpackage appears in both inventories
- **WHEN** the repository still ships subpackages under `internal/cli/`
- **THEN** `README.md` SHALL include every `internal/cli/` subpackage row
- **AND** `docs/architecture/project-structure.md` SHALL include every `internal/cli/` subpackage row
- **AND** that parity SHALL cover both command families and helper packages such as `cliboot/`, `clihttp/`, `clitypes/`, `tuicore/`, `workbench/`, and `workbenchstart/`
