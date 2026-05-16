## MODIFIED Requirements
### Requirement: Mission Control surfaces unavailable projector state explicitly
The cockpit Mission Control page SHALL distinguish an unavailable mission-control projector from a normal empty mission state.

#### Scenario: Composer/activity focus routes vertical navigation to the activity lane
- **WHEN** Mission Control focus is on the composer/activity lane
- **AND** multiple activity rows exist
- **THEN** pressing `↑/k` or `↓/j` SHALL move the activity cursor instead of being swallowed by the composer
