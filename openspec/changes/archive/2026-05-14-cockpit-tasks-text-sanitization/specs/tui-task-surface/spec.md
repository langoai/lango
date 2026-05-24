## MODIFIED Requirements

### Requirement: Tasks cockpit page
The cockpit SHALL include a Tasks page (PageTasks) showing a table of all background tasks with status, elapsed time, and prompt preview.

#### Scenario: Rendered tasks-page text stays plain and single-line
- **WHEN** task IDs, prompt previews, statuses, origin labels, result text, error text, or transient task-action messages contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the Tasks page SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before rendering it
- **AND** wrapped detail content SHALL be derived from that sanitized single-line text
