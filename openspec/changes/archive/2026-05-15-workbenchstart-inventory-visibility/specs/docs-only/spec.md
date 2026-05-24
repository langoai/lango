## ADDED Requirements

### Requirement: Architecture and README inventory docs include workbenchstart support
The public inventory docs SHALL include the shipped `cli/workbenchstart/` package that seeds context-aware prompts for the bare workbench.

#### Scenario: Workbenchstart package stays visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the architecture inventory SHALL include `cli/workbenchstart/`
- **AND** the README internal tree SHALL include `workbenchstart/`
- **AND** those rows SHALL describe starter/recovery prompt builder responsibilities truthfully
