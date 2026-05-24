## ADDED Requirements

### Requirement: Approvals page distinguishes partial unavailable from empty section data
The cockpit Approvals page SHALL distinguish a missing history store or missing grant store from a configured-but-empty section.

#### Scenario: Missing history store renders section-level unavailable message
- **WHEN** the Approvals page renders with no `HistoryStore` but a configured `GrantStore`
- **THEN** the history section SHALL explain that approval history is not configured

#### Scenario: Missing grant store renders section-level unavailable message
- **WHEN** the Approvals page renders with no `GrantStore` but a configured `HistoryStore`
- **THEN** the grants section SHALL explain that active grants are not configured
