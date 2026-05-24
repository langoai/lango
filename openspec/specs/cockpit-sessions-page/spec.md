## Purpose

Capability spec for cockpit-sessions-page. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Sessions page with session list
The cockpit SHALL include a Sessions page showing session key and relative last update time, ordered newest-first by `UpdatedAt`.

#### Scenario: Sessions are sorted newest-first after load
- **WHEN** the Sessions page loads multiple session summaries with different `UpdatedAt` values
- **THEN** it SHALL render the most recently updated session first
- **AND** older sessions SHALL appear below newer ones

#### Scenario: Sessions help appears only when another row exists
- **WHEN** the Sessions page help is rendered with two or more loaded session rows
- **THEN** it SHALL advertise `↑/k` and `↓/j`

#### Scenario: Sessions help hides inert navigation with fewer than two rows
- **WHEN** the Sessions page help is rendered with zero or one loaded session rows
- **THEN** it SHALL omit vertical navigation bindings

#### Scenario: Rendered sessions-page text stays plain and single-line
- **WHEN** session keys or configured-source error text contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the Sessions page SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before rendering it

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

### Requirement: Sessions page distinguishes unavailable from empty session history
The cockpit Sessions page SHALL distinguish between an unavailable session-list source, a configured-source failure, and a configured-but-empty session history.

#### Scenario: Nil list function renders unavailable message
- **WHEN** the Sessions page renders with no configured list function
- **THEN** the page SHALL explain that the session list is not configured

#### Scenario: Configured list failure renders explicit failure message
- **WHEN** the Sessions page renders with a configured list function that failed to load sessions
- **THEN** the page SHALL explain that the session list failed to load
- **AND** it SHALL include the underlying error text

#### Scenario: Empty configured list renders no-sessions message
- **WHEN** the Sessions page renders with a configured list function that returns zero sessions
- **THEN** the page SHALL display `No sessions found.`
