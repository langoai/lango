## ADDED Requirements

### Requirement: README internal package tree includes current mission-projection packages
The README internal package tree SHALL include the shipped durable mission projection packages instead of omitting them.

#### Scenario: Mission projection packages stay visible
- **WHEN** a maintainer updates the README internal package tree
- **THEN** it SHALL include `proposal/`, `loopview/`, and `collabview/`
- **AND** those rows SHALL describe transient proposal flow, deterministic operator-loop projection, and deterministic mission-collaboration projection truthfully
