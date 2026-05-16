## ADDED Requirements

### Requirement: Settlement progression escalation fails closed on unknown current state

The settlement progression service SHALL return an actionable error instead of panicking when escalation is requested from an unsupported current progression status.

#### Scenario: Escalation from unknown progression status returns an error
- **WHEN** `ApplyReleaseOutcome` maps an `escalate` decision while the current settlement progression status is unknown
- **THEN** the call SHALL return an error identifying the unsupported current settlement progression status
- **AND** SHALL not panic
