## ADDED Requirements

### Requirement: Budget mode in knowledge feature status
The knowledge feature status SHALL include whether context budget management is active or not. When a budget manager is configured, the status reason SHALL include "budgeted (<window>k)". When no budget manager is set, existing reason text SHALL be preserved.

#### Scenario: Budget manager active
- **WHEN** knowledge is enabled and a budget manager is configured with 128k model window
- **THEN** the knowledge FeatureStatus reason SHALL include "budgeted (128k)" in addition to existing FTS5 status

#### Scenario: Budget manager not active
- **WHEN** knowledge is enabled but no budget manager is configured
- **THEN** the knowledge FeatureStatus reason SHALL not include budget information
