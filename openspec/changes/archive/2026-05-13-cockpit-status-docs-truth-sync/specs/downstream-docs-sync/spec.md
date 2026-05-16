## ADDED Requirements

### Requirement: Public cockpit docs describe Status page unavailable messaging
Public cockpit documentation SHALL describe the concrete unavailable-state behavior of the Status page.

#### Scenario: Cockpit feature page describes status dependency gaps
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** it SHALL explain that a missing feature-status provider surfaces explicit unavailable messaging in the Feature Status section
- **AND** it SHALL explain that a missing observability collector surfaces explicit unavailable messaging in the Token Usage, Tool Execution, and Graph Admission sections
