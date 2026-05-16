## MODIFIED Requirements

### Requirement: Approvals cockpit page
The cockpit SHALL include an Approvals page with two sections: History (top) and Grants (bottom), accessible via sidebar and ctrl+6 keybinding.

#### Scenario: Rendered approvals-page text stays plain and single-line
- **WHEN** approval-history tool names, summaries, outcomes, providers, or active-grant session/tool labels contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the Approvals page SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before rendering it
