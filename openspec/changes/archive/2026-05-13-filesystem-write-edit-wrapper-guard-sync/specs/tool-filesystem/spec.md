## ADDED Requirements

### Requirement: Filesystem write/edit tools keep actionable wrapper parameter guards

Filesystem write/edit tools SHALL reject missing required wrapper inputs with actionable parameter errors before file mutation begins.

#### Scenario: Filesystem write/edit reject missing required inputs
- **WHEN** `fs_write` or `fs_edit` is invoked without one of its declared required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not proceed into downstream file mutation logic
