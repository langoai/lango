## MODIFIED Requirements

### Requirement: CLI status reference page exists
A dedicated CLI reference page SHALL exist for the `lango status` command.

#### Scenario: Status CLI reference documents current flags
- **WHEN** a user navigates to `docs/cli/status.md`
- **THEN** the page SHALL document `--output` flag, `--addr` flag, output sections, and JSON schema
- **AND** the `--addr` documentation SHALL state that omission uses configured `server.host` and `server.port` before falling back to localhost/18789

### Requirement: Metrics CLI docs use current built-in names
Public metrics CLI docs SHALL stay aligned with the implemented metrics command family.

#### Scenario: Metrics docs use configured gateway default wording
- **WHEN** a user reads `docs/cli/metrics.md`
- **THEN** the page SHALL document that `--addr` overrides the configured gateway address
- **AND** it SHALL not describe `http://localhost:18789` as the only default for `lango metrics`

### Requirement: Alerts CLI docs describe current gateway targeting
Public alerts CLI docs SHALL stay aligned with the implemented alerts command family.

#### Scenario: Alerts docs use configured gateway default wording
- **WHEN** a user reads `docs/cli/alerts.md`
- **THEN** the page SHALL document that `--addr` overrides the configured gateway address
- **AND** it SHALL not describe `http://localhost:18789` as the only default for `lango alerts`
