## ADDED Requirements

### Requirement: Sessions page distinguishes unavailable from empty session history
The cockpit Sessions page SHALL distinguish between an unavailable session-list source and a configured-but-empty session history.

#### Scenario: Nil list function renders unavailable message
- **WHEN** the Sessions page renders with no configured list function
- **THEN** the page SHALL explain that the session list is not configured

#### Scenario: Empty configured list renders no-sessions message
- **WHEN** the Sessions page renders with a configured list function that returns zero sessions
- **THEN** the page SHALL display `No sessions found.`
