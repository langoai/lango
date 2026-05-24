## ADDED Requirements

### Requirement: Top-level internal-package parity guard stays executable
Repository-level regressions that let README or architecture inventory docs fall out of parity with the shipped top-level `internal/` package tree SHALL be enforced by an executable test.

#### Scenario: Every top-level internal package remains represented in both inventories
- **WHEN** the repository still ships top-level packages under `internal/`
- **THEN** an executable repository test SHALL fail if any top-level package disappears from `README.md`
- **AND** it SHALL fail if any top-level package disappears from `docs/architecture/project-structure.md`
- **AND** it SHALL therefore catch omissions such as leaving `automation/`, `deadline/`, or `llm/` out of the architecture inventory while they still exist in the codebase
