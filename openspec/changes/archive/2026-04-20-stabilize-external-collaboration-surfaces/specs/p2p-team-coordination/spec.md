## ADDED Requirements

### Requirement: Team operator surfaces remain truthful until live control is implemented
The team operator surfaces SHALL distinguish between the real runtime team subsystem and the current guidance-oriented CLI surfaces.

#### Scenario: Team CLI guidance
- **WHEN** a user reads or runs the `lango p2p team` CLI commands
- **THEN** the system SHALL describe them as guidance-oriented or runtime-backed surfaces rather than claiming direct live control if that control path does not yet exist

### Requirement: Team conflict and payment documentation reflects current implementation
The documented conflict-resolution and payment-coordination behavior SHALL match the current implementation semantics.

#### Scenario: Conflict strategy wording
- **WHEN** a user reads the team conflict-resolution documentation
- **THEN** the descriptions SHALL NOT claim stronger implementation behavior than the current coordinator actually provides

#### Scenario: Team payment threshold wording
- **WHEN** a user reads team payment coordination documentation
- **THEN** the payment threshold wording SHALL match the canonical inclusive `0.8` post-pay rule
