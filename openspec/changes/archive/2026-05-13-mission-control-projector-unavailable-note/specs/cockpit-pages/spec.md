## ADDED Requirements

### Requirement: Mission Control surfaces unavailable projector state explicitly
The cockpit Mission Control page SHALL distinguish an unavailable mission-control projector from a normal empty mission state.

#### Scenario: Nil projector renders degraded note
- **WHEN** Mission Control renders after activation with no configured projector
- **THEN** the page SHALL still render its empty-state shell
- **AND** it SHALL include a degraded note explaining that Mission Control data is not configured
