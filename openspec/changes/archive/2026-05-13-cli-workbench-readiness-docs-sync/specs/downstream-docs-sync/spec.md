## MODIFIED Requirements

### Requirement: First-touch public docs mirror current CLI and cockpit operator surfaces
The repository SHALL keep first-touch public docs aligned with the runtime behavior and discoverability of the current CLI and cockpit entry surfaces.

#### Scenario: CLI index describes workbench readiness split
- **WHEN** the public CLI index describes bare `lango`
- **THEN** it SHALL mention that incomplete profiles are guided toward `lango onboard`, `lango settings`, and `lango doctor`
- **AND** SHALL mention that ready profiles expose starter prompts and the `Enter` / `1-3` quick-start path

#### Scenario: Quickstart guide describes workbench readiness split
- **WHEN** the getting-started quickstart guide introduces bare `lango`
- **THEN** it SHALL mention the same setup-recovery path for incomplete profiles
- **AND** SHALL mention the starter-prompt quick-start path for ready profiles
