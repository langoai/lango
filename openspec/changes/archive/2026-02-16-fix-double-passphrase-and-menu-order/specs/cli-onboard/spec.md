## MODIFIED Requirements

### User Interface
- **Navigation**:
    - Users MUST be able to navigate between configuration categories freely.
    - Uses a menu-based system (e.g., Main Menu -> Category -> Form).
    - The menu SHALL include categories in this order: Providers, Agent, Server, Channels, Tools, Session, Security, Knowledge, Observational Memory, Embedding & RAG, Save & Exit, Cancel.
    - Form cursor navigation SHALL NOT panic when navigating past the first or last field.

#### Scenario: Providers category appears first
- **WHEN** user views the configuration menu
- **THEN** "Providers" SHALL be the first category in the menu, before "Agent"

#### Scenario: Session category in menu
- **WHEN** user views the configuration menu
- **THEN** "Session" category SHALL be listed before "Security"

#### Scenario: Knowledge category in menu
- **WHEN** user views the configuration menu
- **THEN** "Knowledge" category SHALL be listed after "Security" and before "Observational Memory"
