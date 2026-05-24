## ADDED Requirements

### Requirement: Approvals page distinguishes unavailable from empty stores
The cockpit Approvals page SHALL distinguish between an unavailable approval subsystem and a configured-but-empty approval history.

#### Scenario: Nil stores render unavailable message
- **WHEN** the Approvals page renders with both `HistoryStore` and `GrantStore` absent
- **THEN** the page SHALL explain that approval history and grants are not configured

#### Scenario: Empty configured stores render empty-history message
- **WHEN** the Approvals page renders with configured stores but no history entries and no grants
- **THEN** the page SHALL display `No approval history yet.`
