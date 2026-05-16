## MODIFIED Requirements

### Requirement: Approval surfaces keep chat rendering stable
The chat TUI SHALL keep approval rendering stable across inline and fullscreen surfaces, including defensive fallback paths for internal renderer state.

#### Scenario: Fullscreen approval dialog tolerates nil renderer state
- **WHEN** the fullscreen approval dialog renders diff content without a preinitialized approval state
- **THEN** it SHALL still render without panicking
- **AND** SHALL fall back to internal default dialog state as needed
