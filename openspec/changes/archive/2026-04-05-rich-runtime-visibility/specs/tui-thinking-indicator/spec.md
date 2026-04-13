## MODIFIED Requirements

### Requirement: Thinking indicator displays active state
The system SHALL display a thinking indicator with summary text preview when the agent is actively thinking. The active state SHALL show "💭 Thinking..." followed by a truncated preview of the thinking summary text using `ansi.Truncate`. If the summary is empty, only "💭 Thinking..." is displayed.

#### Scenario: Active thinking with summary
- **WHEN** a ThinkingStartedMsg arrives with summary="analyzing user query for search terms"
- **THEN** the indicator SHALL display "💭 Thinking..." followed by a truncated preview of the summary in italic muted style

#### Scenario: Active thinking with empty summary
- **WHEN** a ThinkingStartedMsg arrives with summary=""
- **THEN** the indicator SHALL display "💭 Thinking..." with no preview text

#### Scenario: Summary truncated for narrow width
- **WHEN** the thinking summary exceeds the available display width
- **THEN** the preview SHALL be truncated with "..." using `ansi.Truncate` with a minimum preview width of 10
