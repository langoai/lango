## ADDED Requirements

### Requirement: Sessions page with session list
SessionsPage SHALL display sessions from Store.ListSessions() with cursor navigation. Each entry SHALL show the session key and relative time since last update.

#### Scenario: View sessions
- **WHEN** SessionsPage is active
- **THEN** it SHALL display all sessions ordered by most recent update

#### Scenario: Empty session list
- **WHEN** no sessions exist
- **THEN** it SHALL display "No sessions found."

### Requirement: Page interface compliance
SessionsPage SHALL implement the Page interface with Activate() refreshing the session list.

#### Scenario: Activate refreshes
- **WHEN** SessionsPage.Activate() is called
- **THEN** it SHALL call the list function and populate the session list

### Requirement: Store interface extension
session.Store SHALL include a ListSessions(ctx context.Context) method returning []SessionSummary. SessionSummary SHALL contain Key, CreatedAt, and UpdatedAt fields.

#### Scenario: List sessions from store
- **GIVEN** sessions exist in the store
- **WHEN** ListSessions is called
- **THEN** it SHALL return SessionSummary entries ordered by UpdatedAt descending

### Requirement: Sidebar integration
The sessions sidebar item SHALL be enabled (Disabled: false) and navigable via PageSessions PageID.

#### Scenario: Navigate to sessions
- **WHEN** sessions is selected in the sidebar
- **THEN** the cockpit SHALL switch to the SessionsPage
