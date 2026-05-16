## MODIFIED Requirements

### Requirement: TUI startup notices remain seam-aware
Interactive top-level TUI entrypoints SHALL route their startup notice text through seam-aware stderr writers so wrapper and regression captures do not depend on process-global stderr interception.

#### Scenario: TUI startup bootstrap rendering stays shared and consistent
- **WHEN** chat, cockpit, and workbench prepare logging and render their startup notices
- **THEN** they SHALL use one shared bootstrap helper while preserving the existing startup notice contract
