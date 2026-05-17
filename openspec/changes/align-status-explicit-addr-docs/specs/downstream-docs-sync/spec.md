## MODIFIED Requirements

### Requirement: Gateway-backed CLI docs describe configured gateway resolution
Public CLI docs for gateway-backed metrics, alerts, and status commands SHALL document that `--addr` is an override and that omission uses configured server host and port before falling back to localhost/18789.

#### Scenario: Status docs describe explicit target display
- **WHEN** public CLI docs describe `lango status --addr <url>`
- **THEN** they SHALL state that status probes the normalized explicit address
- **AND** they SHALL state that the `gateway` output field reports that same normalized explicit target
