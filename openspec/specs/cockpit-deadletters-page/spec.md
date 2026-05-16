## Purpose

Capability spec for cockpit-deadletters-page. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Dead Letters page renders dead-letter metadata as plain text
The cockpit SHALL include a Dead Letters page that renders backlog rows, summary strips, selected-transaction detail, and retry follow-up messaging from post-adjudication dead-letter data.

#### Scenario: Rendered Dead Letters text stays plain and single-line
- **WHEN** dead-letter transaction identifiers, reasons, adjudication labels, dispatch references, actor labels, subtype/family labels, detail values, or retry status messages contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the Dead Letters page SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before rendering it
