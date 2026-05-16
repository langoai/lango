## ADDED Requirements

### Requirement: Workbench empty state guides incomplete profiles
The standalone workbench empty state SHALL guide the operator toward setup and verification when the active profile is obviously incomplete.

#### Scenario: Incomplete profile shows setup guidance
- **WHEN** bare `lango` renders an empty Mission Control workbench state with an incomplete profile
- **THEN** the empty state SHALL mention `lango onboard`, `lango settings`, and `lango doctor`
- **AND** SHALL keep the existing chat and cockpit navigation hints

### Requirement: Workbench empty state stays clean for ready profiles
The standalone workbench empty state SHALL omit setup guidance when the active profile already has a usable provider/model path.

#### Scenario: Ready profile omits setup guidance
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **THEN** the empty state SHALL NOT mention `lango onboard`, `lango settings`, or `lango doctor`
