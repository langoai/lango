## MODIFIED Requirements

### Requirement: Gateway-backed CLI address resolution coverage stays executable
Repository-level regressions in gateway-backed CLI default address resolution SHALL be enforced by executable tests.

#### Scenario: Metrics CLI configured gateway default remains covered
- **WHEN** `lango metrics` and metrics subcommands are constructed with a config loader
- **THEN** executable tests SHALL fail if they ignore configured `server.host` and `server.port` when `--addr` is omitted
- **AND** executable tests SHALL fail if explicit `--addr` stops overriding the configured gateway

#### Scenario: Alerts CLI configured gateway default remains covered
- **WHEN** `lango alerts list` or `lango alerts summary` is constructed with a config loader
- **THEN** executable tests SHALL fail if they ignore configured `server.host` and `server.port` when `--addr` is omitted
- **AND** executable tests SHALL fail if explicit `--addr` stops overriding the configured gateway

#### Scenario: Status CLI configured gateway probe remains covered
- **WHEN** `lango status` bootstraps a config with custom `server.host` and `server.port`
- **THEN** executable tests SHALL fail if the live `/health` probe still uses a hardcoded localhost/18789 default

### Requirement: Gateway CLI docs default wording guard stays executable
Repository-level docs guards SHALL prevent gateway-backed CLI docs from presenting localhost/18789 as the only default when the command now honors configured server host and port.

#### Scenario: Gateway CLI docs configured-default wording remains covered
- **WHEN** public CLI docs for metrics, alerts, or status are checked
- **THEN** executable tests SHALL fail if the docs omit configured-gateway default wording for `--addr`
