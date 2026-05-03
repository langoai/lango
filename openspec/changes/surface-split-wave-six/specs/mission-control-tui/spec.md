## MODIFIED Requirements

### Requirement: Mission Control is a reusable mission surface across workbench and cockpit

Mission Control SHALL remain the shared mission-native surface used by the interactive workbench and available inside cockpit. It SHALL continue to make ongoing work, the latest live decision, recent activity, and the shared composer path available without requiring the user to navigate to another workflow first.

#### Scenario: Workbench mounts Mission Control directly
- **WHEN** the user runs bare `lango` on an interactive terminal
- **THEN** the workbench SHALL mount Mission Control as its primary body
- **AND** the first screen SHALL include a short hint that `lango chat` remains available as focused chat
- **AND** the first Wave 6 slice SHALL also hint that `lango cockpit` remains available as the explicit dashboard

#### Scenario: `lango chat` remains direct chat fallback
- **WHEN** the user runs `lango chat`
- **THEN** the application SHALL bypass Mission Control and start the focused chat surface directly

#### Scenario: Cockpit still exposes Mission Control as a page
- **WHEN** the user runs `lango cockpit`
- **THEN** Mission Control SHALL remain available through the cockpit page set
- **AND** Wave 6 SHALL NOT require a second Mission Control domain or projection system
