## ADDED Requirements
### Requirement: Doctor help reflects the current diagnostic surface
The `cli-help-text` main spec SHALL track the current `doctor --help` diagnostic families and total count rather than a stale earlier subset.

#### Scenario: Doctor help lists the current families and count
- **WHEN** a maintainer inspects the `doctor --help` contract
- **THEN** it reflects the current total count and the current diagnostic families shown by the production command
