## MODIFIED Requirements

### Requirement: Context panel renders token and tool metrics
The context panel SHALL display token usage (input/output/total/cache), top-5 tools by execution count, and system uptime.

#### Scenario: Rendered context-panel labels stay plain and single-line
- **WHEN** top-tool names, runtime active-agent labels, or channel names contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the context panel SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before rendering it
