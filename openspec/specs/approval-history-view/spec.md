### Requirement: In-memory approval history store
The system SHALL maintain an in-memory ring buffer (default 500 entries) recording every approval decision with timestamp, request ID, tool name, session key, summary, safety level, outcome (open set), and provider.

#### Scenario: Append and list
- **WHEN** 3 approval decisions are recorded
- **THEN** List() returns all 3 entries in newest-first order

#### Scenario: Ring buffer overflow
- **WHEN** more than maxSize entries are appended
- **THEN** the oldest entries are evicted and only the most recent maxSize entries are retained

#### Scenario: Count by outcome
- **WHEN** CountByOutcome() is called
- **THEN** a map of outcome string to count is returned

### Requirement: Approval middleware history recording
The approval middleware SHALL record every decision point to the HistoryStore: session grant bypass, turn-local grant bypass, turn-local denial/timeout replay, spending limiter auto-approve, approval granted, approval denied.

#### Scenario: Session grant bypass recorded
- **WHEN** a tool call is auto-approved by a session grant
- **THEN** a history entry with outcome="bypass" and provider="grant_store" is appended

#### Scenario: Spending limiter bypass recorded
- **WHEN** a payment tool is auto-approved by the spending limiter
- **THEN** a history entry with outcome="bypass" and provider="spending_limiter" is appended

#### Scenario: Approval granted recorded
- **WHEN** a user approves a tool call
- **THEN** a history entry with the actual outcome and provider is appended

### Requirement: Approvals cockpit page
The cockpit SHALL include an Approvals page with two sections: History (top) and Grants (bottom), accessible via sidebar and ctrl+6 keybinding.

#### Scenario: History section display
- **WHEN** the Approvals page is active and history entries exist
- **THEN** a table showing Time, Tool, Summary, Outcome, Provider is rendered

#### Scenario: Grants section display
- **WHEN** the user switches to the Grants section via `/`
- **THEN** a table showing Session, Tool, Granted time is rendered

#### Scenario: Grant revocation
- **WHEN** user presses `r` in the grants section on a selected grant
- **THEN** the grant is revoked from the GrantStore

#### Scenario: Session grant revocation
- **WHEN** user presses `R` in the grants section
- **THEN** all grants for the selected session are revoked

#### Scenario: Empty state
- **WHEN** no history and no grants exist
- **THEN** the page displays "No approval history yet."

#### Scenario: Independent section cursors
- **WHEN** user switches between history and grants sections
- **THEN** each section preserves its cursor position independently
