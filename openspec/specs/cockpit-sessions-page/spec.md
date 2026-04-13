## Purpose

Capability spec for cockpit-sessions-page. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Sessions page with session list
SessionsPage SHALL display sessions from Store.ListSessions() with cursor navigation. Each entry SHALL show the session key and relative time since last update.

#### Scenario: View sessions
- **WHEN** SessionsPage is active
- **THEN** it SHALL display all sessions ordered by most recent update

### Requirement: Page interface compliance
SessionsPage SHALL implement the Page interface with Activate() refreshing the session list.

#### Scenario: Activate refreshes
- **WHEN** SessionsPage.Activate() is called
- **THEN** it SHALL call the list function and populate the session list

### Requirement: Store interface extension
session.Store SHALL include a ListSessions(ctx context.Context) method returning []SessionSummary.

#### Scenario: List sessions from store
- **GIVEN** sessions exist in the store
- **WHEN** ListSessions is called
- **THEN** it SHALL return SessionSummary entries ordered by UpdatedAt descending

### Requirement: Sidebar integration
The sessions sidebar item SHALL be enabled (Disabled: false) and navigable.

#### Scenario: Navigate to sessions
- **WHEN** sessions is selected in the sidebar
- **THEN** the cockpit SHALL switch to the SessionsPage
