## MODIFIED Requirements

### Requirement: Public docs list implemented P2P command surface
Public P2P documentation SHALL include the currently implemented P2P command families and SHALL document gateway-backed P2P provenance address selection.

#### Scenario: P2P provenance docs describe gateway address selection
- **WHEN** a user reads `docs/cli/p2p.md`
- **THEN** the provenance section SHALL state that `--addr` overrides the configured gateway
- **AND** it SHALL state that omitted `--addr` uses configured `server.host` and `server.port`
- **AND** it SHALL state that explicit `--addr` values are normalized before gateway requests
