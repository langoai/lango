## ADDED Requirements

### Requirement: Status page renders explicit unavailable messaging for missing dependencies
The cockpit Status page SHALL distinguish missing feature-status or observability dependencies from valid zero-valued data.

#### Scenario: Missing feature-status provider renders unavailable message
- **WHEN** the Status page renders with no feature-status provider
- **THEN** the Feature Status section SHALL explain that the feature status provider is not configured

#### Scenario: Missing metrics collector renders unavailable message
- **WHEN** the Status page renders with no observability metrics collector
- **THEN** the Token Usage, Tool Execution, and Graph Admission sections SHALL explain that the metrics collector is not configured
