## ADDED Requirements
### Requirement: Public config create docs include preset-backed profile creation
Public config CLI docs SHALL describe the implemented `--preset` flag on `lango config create`.

#### Scenario: Config create docs expose preset-backed creation
- **WHEN** a reader consults the public config CLI docs
- **THEN** they SHALL find `lango config create <name> [--preset <name>]`
- **AND** they SHALL find the supported preset values documented
