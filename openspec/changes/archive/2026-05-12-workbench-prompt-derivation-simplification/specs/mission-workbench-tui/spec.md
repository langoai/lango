## ADDED Requirements

### Requirement: Workbench prompt derivation depends on stable workspace inputs

The standalone workbench SHALL derive starter prompt behavior from stable workspace inputs such as `workDir` through the shared prompt helper, rather than relying on a separately transported precomputed prompt slice.

#### Scenario: Mission Control page derives prompts from workDir
- **WHEN** the workbench Mission Control page needs ready-profile starter prompts
- **THEN** it SHALL derive them from the shared prompt helper using the current workspace input
- **AND** the starter prompt behavior SHALL stay consistent with the rest of the workbench shell
