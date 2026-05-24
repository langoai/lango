## ADDED Requirements

### Requirement: Replay tool enforces actor policy
The system SHALL require actor resolution and outcome-aware replay authorization before allowing `retry_post_adjudication_execution`.

#### Scenario: Replay fails closed on actor policy
- **WHEN** actor resolution fails or the actor is not permitted
- **THEN** replay SHALL be denied
