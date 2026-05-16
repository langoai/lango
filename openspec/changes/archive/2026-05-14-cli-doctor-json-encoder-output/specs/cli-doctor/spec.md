## MODIFIED Requirements

### Requirement: Doctor Command Entry Point
The system SHALL include RunLedger diagnostics in the `lango doctor` command output and help text.

#### Scenario: Doctor JSON output remains decodable
- **WHEN** `lango doctor --json` renders output
- **THEN** the output SHALL be valid pretty-printed JSON that can be decoded directly by wrappers without stripping extra framing text
