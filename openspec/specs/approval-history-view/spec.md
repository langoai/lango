## Purpose

Capability spec for approval-history-view. See requirements below for scope and behavior contracts.
## Requirements
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

#### Scenario: Help advertises both section-toggle keys readably
- **WHEN** the Approvals page help is rendered
- **THEN** it SHALL advertise both `tab` and `/` as the section-toggle keys
- **AND** the rendered help key label SHALL be `tab /`

#### Scenario: Section toggle accepts Tab and slash
- **WHEN** the operator is viewing the Approvals page
- **AND** presses `tab` or `/`
- **THEN** the page SHALL switch between the History and Grants sections

#### Scenario: Navigation help appears only when another row exists
- **WHEN** the active Approvals section has two or more rows
- **THEN** the help SHALL advertise `↑/k` and `↓/j`

#### Scenario: Navigation help hides inert keys with zero or one row
- **WHEN** the active Approvals section has zero or one row
- **THEN** the help SHALL omit `↑/k` and `↓/j`

#### Scenario: Revoke help appears only when grant rows exist
- **WHEN** the Approvals page help is rendered for the Grants section with one or more grant rows
- **THEN** it SHALL advertise `r` and `R`

#### Scenario: Revoke help hides inert actions in empty grants section
- **WHEN** the Approvals page help is rendered for the Grants section with zero grant rows
- **THEN** it SHALL omit `r` and `R`

#### Scenario: Rendered approvals-page text stays plain and single-line
- **WHEN** approval-history tool names, summaries, outcomes, providers, or active-grant session/tool labels contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the Approvals page SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before rendering it

### Requirement: Approvals page distinguishes unavailable from empty stores
The cockpit Approvals page SHALL distinguish between an unavailable approval subsystem and a configured-but-empty approval history.

#### Scenario: Nil stores render unavailable message
- **WHEN** the Approvals page renders with both `HistoryStore` and `GrantStore` absent
- **THEN** the page SHALL explain that approval history and grants are not configured

#### Scenario: Empty configured stores render empty-history message
- **WHEN** the Approvals page renders with configured stores but no history entries and no grants
- **THEN** the page SHALL display `No approval history yet.`

### Requirement: Approvals page distinguishes partial unavailable from empty section data
The cockpit Approvals page SHALL distinguish a missing history store or missing grant store from a configured-but-empty section.

#### Scenario: Missing history store renders section-level unavailable message
- **WHEN** the Approvals page renders with no `HistoryStore` but a configured `GrantStore`
- **THEN** the history section SHALL explain that approval history is not configured

#### Scenario: Empty history section uses unified empty wording
- **WHEN** the Approvals page renders with a configured but empty `HistoryStore`
- **AND** a configured `GrantStore`
- **THEN** the history section SHALL display `No approval history yet.`

#### Scenario: Missing grant store renders section-level unavailable message
- **WHEN** the Approvals page renders with no `GrantStore` but a configured `HistoryStore`
- **THEN** the grants section SHALL explain that active grants are not configured
