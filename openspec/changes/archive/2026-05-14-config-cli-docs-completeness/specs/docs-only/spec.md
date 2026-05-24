## ADDED Requirements
### Requirement: Public config CLI docs include get/set/keys
Public CLI documentation SHALL include the implemented `lango config get`, `lango config set`, and `lango config keys` commands.

#### Scenario: Config CLI docs expose the implemented read/write/key surfaces
- **WHEN** a reader consults the public config CLI docs or the CLI index
- **THEN** they SHALL find `lango config get`, `lango config set`, and `lango config keys` documented with their real argument shapes
