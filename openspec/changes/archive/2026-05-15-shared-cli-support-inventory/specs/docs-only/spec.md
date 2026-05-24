## ADDED Requirements

### Requirement: Architecture and README inventory docs include shared CLI support packages
The public inventory docs SHALL include the shipped shared CLI support packages that back gateway-oriented commands.

#### Scenario: Shared CLI support packages stay visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the architecture inventory SHALL include `cli/cliboot/` and `cli/clihttp/`
- **AND** the README internal tree SHALL include `cliboot/` and `clihttp/`
- **AND** those rows SHALL describe bootstrap loader callbacks and shared HTTP/JSON helper responsibilities truthfully
