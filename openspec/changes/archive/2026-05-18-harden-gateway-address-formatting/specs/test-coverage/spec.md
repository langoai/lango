## MODIFIED Requirements

### Requirement: Gateway-backed CLI address resolution coverage stays executable
Repository-level regressions in gateway-backed CLI default address resolution SHALL be enforced by executable tests.

#### Scenario: Gateway IPv6 formatting remains covered
- **WHEN** gateway URL or listen address helpers receive an IPv6 host
- **THEN** executable tests SHALL fail if the formatted address omits required IPv6 brackets

#### Scenario: P2P provenance gateway resolution remains covered
- **WHEN** P2P provenance push or fetch is constructed with explicit or configured gateway inputs
- **THEN** executable tests SHALL fail if explicit trailing slashes reach the gateway POST client
- **AND** executable tests SHALL fail if omitted `--addr` ignores configured `server.host` and `server.port`
