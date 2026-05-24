## ADDED Requirements

### Requirement: README and architecture inventory stay in top-level internal-package parity
The public README internal tree and architecture project-structure inventory SHALL both mention every shipped top-level `internal/` package so that new package additions do not silently appear in only one document.

#### Scenario: Every top-level internal package appears in both inventories
- **WHEN** the repository still ships top-level packages under `internal/`
- **THEN** `README.md` SHALL include every top-level package row
- **AND** `docs/architecture/project-structure.md` SHALL include every top-level package row
- **AND** the architecture inventory SHALL continue to include `automation/`, `deadline/`, and `llm/` alongside the other shipped top-level packages
