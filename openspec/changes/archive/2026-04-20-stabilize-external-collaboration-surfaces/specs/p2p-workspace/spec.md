## ADDED Requirements

### Requirement: Workspace operator surfaces remain truthful until live control is implemented
Workspace and git-bundle operator surfaces SHALL distinguish between the real runtime subsystems and the current guidance-oriented CLI surfaces.

#### Scenario: Workspace CLI guidance
- **WHEN** a user reads or runs `lango p2p workspace` commands
- **THEN** the system SHALL describe the current server-backed or tool-backed path instead of implying a fully direct live CLI control path

#### Scenario: Git bundle CLI guidance
- **WHEN** a user reads or runs `lango p2p git` commands
- **THEN** the system SHALL describe the current server-backed or tool-backed artifact exchange path instead of implying full direct live repository control

### Requirement: Chronicler documentation reflects partial wiring
Workspace chronicler documentation SHALL describe graph-triple persistence as dependent on triple-adder wiring being available.

#### Scenario: Chronicler partial wiring documented
- **WHEN** a user reads workspace chronicler documentation
- **THEN** the documentation SHALL describe the current state as partial wiring rather than guaranteed default live persistence
