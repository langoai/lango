## MODIFIED Requirements

### Requirement: Gateway address formatting
Gateway server and CLI surfaces SHALL format configured host/port addresses with bracket-safe host/port joining so IPv6 hosts are valid in both HTTP URLs and listen addresses. Doctor server port checks SHALL reuse the same listen-address formatting policy when probing configured gateway bind addresses.

#### Scenario: Gateway CLI URL formats IPv6 hosts
- **WHEN** gateway CLI code resolves configured `server.host` as `::1` and `server.port` as `18789`
- **THEN** the resulting HTTP URL SHALL be `http://[::1]:18789`

#### Scenario: Gateway listen address formats IPv6 hosts
- **WHEN** the gateway server listens on configured `server.host` `::1` and port `18789`
- **THEN** the listen address SHALL be `[::1]:18789`

#### Scenario: Doctor server port check formats IPv6 listen addresses
- **WHEN** the doctor server port check probes configured `server.host` `::1` and port `18789`
- **THEN** the listen address SHALL be `[::1]:18789`
