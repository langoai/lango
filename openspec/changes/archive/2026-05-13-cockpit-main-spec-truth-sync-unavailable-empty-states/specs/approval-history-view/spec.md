## MODIFIED Requirements

### Requirement: Approvals cockpit page
The cockpit SHALL include an Approvals page with two sections: History (top) and Grants (bottom), accessible via sidebar and ctrl+6 keybinding.

#### Scenario: Empty configured state
- **WHEN** configured approval stores contain no history entries and no grants
- **THEN** the page SHALL display `No approval history yet.`
