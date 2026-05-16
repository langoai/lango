## ADDED Requirements

### Requirement: README config inventory keeps the current command order
The README internal CLI inventory SHALL describe the config family in the current command order.

#### Scenario: Stale config inventory ordering stays removed
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** it SHALL describe `lango config list/create/use/delete/import/export/get/set/keys/validate`
